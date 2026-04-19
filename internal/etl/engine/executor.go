package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

const defaultChannelBuffer = 1000

// RunResult contient les métriques d'exécution d'un bloc.
type RunResult struct {
	NodeID   string
	RowsIn   int64
	RowsOut  int64
	Duration time.Duration
	Err      error
}

// ExecutionReport contient le rapport complet d'un run.
type ExecutionReport struct {
	ProjectID string
	StartedAt time.Time
	EndedAt   time.Time
	Results   []RunResult
	Success   bool
	Preview   map[string][]contracts.DataRow `json:"preview"`
}

// Executor exécute un projet ETL à partir de son DAG.
type Executor struct {
	log       zerolog.Logger
	ActiveEnv string
}

func NewExecutor(log zerolog.Logger, activeEnv string) *Executor {
	return &Executor{log: log, ActiveEnv: activeEnv}
}

// teeChannel intercepte les lignes de src sans modifier le canal original.
// Il crée un NOUVEAU canal (retourné) qui reçoit une copie de chaque ligne
// tout en laissant src intact pour le bloc consommateur.
// BUG CORRIGÉ : ne plus écraser p.Ch in-place — le bloc consommateur
// doit continuer à lire depuis le canal original créé par le producteur.
func teeChannel(
	ctx context.Context,
	src chan contracts.DataRow,
	blockID string,
	ps *contracts.PreviewStore,
) chan contracts.DataRow {
	tee := make(chan contracts.DataRow, cap(src))
	go func() {
		defer close(tee)
		for {
			select {
			case row, ok := <-src:
				if !ok {
					return
				}
				ps.Append(blockID, row)
				select {
				case tee <- row:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return tee
}

// Execute exécute un projet ETL complet.
func (e *Executor) Execute(ctx context.Context, project *contracts.Project) (*ExecutionReport, error) {
	preview := contracts.NewPreviewStore(1000)
	report := &ExecutionReport{
		ProjectID: project.ID,
		StartedAt: time.Now(),
	}

	dag, err := BuildDAG(project)
	if err != nil {
		return nil, fmt.Errorf("executor: %w", err)
	}

	ordered, err := dag.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("executor: %w", err)
	}

	// Index des edges désactivés.
	disabledEdges := make(map[string]bool, len(project.Edges))
	for _, edge := range project.Edges {
		if edge.Disabled {
			disabledEdges[edge.From+"->"+edge.To] = true
		}
	}

	// Créer les canaux entre blocs (sauf edges désactivés).
	// ports[key] = canal brut sur lequel le producteur ÉCRIT.
	// Le consommateur lira depuis ce même canal (via inputPorts).
	ports := make(map[string]*contracts.Port)
	for _, node := range ordered {
		for _, succ := range dag.Successors(node.ID) {
			key := node.ID + "->" + succ.ID
			if disabledEdges[key] {
				e.log.Info().Str("edge", key).Msg("edge désactivé : canal non créé")
				continue
			}
			ports[key] = &contracts.Port{
				ID: key,
				Ch: make(chan contracts.DataRow, defaultChannelBuffer),
			}
		}
	}

	for _, node := range ordered {
		start := time.Now()
		result := RunResult{NodeID: node.ID}

		factory, ok := blocks.Registry[node.Type]
		if !ok {
			err := fmt.Errorf("bloc de type '%s' non enregistré", node.Type)
			result.Err = err
			report.Results = append(report.Results, result)
			report.EndedAt = time.Now()
			return report, err
		}
		block := factory()

		// --- Ports d'entrée ---
		// Le consommateur lit depuis le canal brut (ports[key].Ch).
		// La preview est capturée côté SORTIE du producteur, pas ici.
		var inputPorts []*contracts.Port
		for predID := range dag.adjacency {
			for _, succID := range dag.adjacency[predID] {
				if succID == node.ID {
					key := predID + "->" + node.ID
					if p, exists := ports[key]; exists {
						inputPorts = append(inputPorts, p)
					}
				}
			}
		}

		// --- Ports de sortie avec tee preview ---
		// CORRECTION : on crée un Port WRAPPER dont le Ch est le canal tee.
		// Le Port original dans ports[key] reste inchangé — le bloc suivant
		// lira toujours depuis ports[key].Ch (le canal brut du producteur).
		// La goroutine tee lit depuis ports[key].Ch et recopie vers tee.Ch.
		var outputPorts []*contracts.Port
		for _, succ := range dag.Successors(node.ID) {
			key := node.ID + "->" + succ.ID
			if originalPort, exists := ports[key]; exists {
				// Le producteur écrit dans originalPort.Ch.
				// La goroutine tee lit depuis originalPort.Ch → capture preview → recopie vers teeCh.
				// Mais le consommateur lit aussi depuis originalPort.Ch...
				// → On doit brancher le tee ENTRE producteur et consommateur.
				//
				// Architecture correcte :
				//   producteur → rawCh → goroutine tee → teeCh → consommateur
				//
				// On crée rawCh (le producteur y écrira via outputPorts),
				// tee lit rawCh et écrit dans originalPort.Ch (que le consommateur lit déjà).
				rawCh := make(chan contracts.DataRow, defaultChannelBuffer)

				// Lance la goroutine tee : rawCh → preview + originalPort.Ch
				go func(src chan contracts.DataRow, dst chan contracts.DataRow, bID string) {
					defer close(dst)
					for {
						select {
						case row, ok := <-src:
							if !ok {
								return
							}
							preview.Append(bID, row)
							select {
							case dst <- row:
							case <-ctx.Done():
								return
							}
						case <-ctx.Done():
							return
						}
					}
				}(rawCh, originalPort.Ch, node.ID)

				// Le producteur reçoit un Port pointant sur rawCh.
				outputPorts = append(outputPorts, &contracts.Port{
					ID: key,
					Ch: rawCh,
				})
			}
		}

		// Les blocs terminaux (targets) n'ont pas de successeurs :
		// on leur donne un Port de sortie factice pour la preview uniquement.
		if len(outputPorts) == 0 && len(inputPorts) > 0 {
			// Pas de successeur → pas de preview de sortie pour ce bloc (normal).
			// outputPorts reste vide : le bloc target doit gérer bctx.Outputs vide.
		}

		bctx := &contracts.BlockContext{
			Ctx:       ctx,
			Params:    node.ParamMap(),
			ConnRef:   node.ConnRef,
			ActiveEnv: e.ActiveEnv,
			BlockID:   node.ID,
			Preview:   preview,
			Inputs:    inputPorts,
			Outputs:   outputPorts,
		}

		e.log.Info().Str("node", node.ID).Str("type", node.Type).Msg("exécution bloc")

		if err := block.Run(bctx); err != nil {
			result.Err = err
			result.Duration = time.Since(start)
			report.Results = append(report.Results, result)
			report.EndedAt = time.Now()
			e.log.Error().Str("node", node.ID).Err(err).Msg("bloc en erreur")
			return report, fmt.Errorf("executor: bloc '%s': %w", node.ID, err)
		}

		result.Duration = time.Since(start)
		report.Results = append(report.Results, result)
		e.log.Info().Str("node", node.ID).Dur("durée", result.Duration).Msg("bloc terminé")
	}

	report.EndedAt = time.Now()
	report.Success = true
	report.Preview = preview.All()
	return report, nil
}
