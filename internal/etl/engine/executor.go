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
	Preview   map[string][]contracts.DataRow `json:"preview"` // N premières lignes par bloc
}

// Executor exécute un projet ETL à partir de son DAG.
type Executor struct {
	log       zerolog.Logger
	ActiveEnv string
}

// NewExecutor crée un Executor.
func NewExecutor(log zerolog.Logger, activeEnv string) *Executor {
	return &Executor{log: log, ActiveEnv: activeEnv}
}

// teeChannel intercepte chaque ligne passant dans src :
// elle est capturée dans le PreviewStore puis ré-émise sur le canal retourné.
// Non-bloquant : n'interrompt pas le flux même si la preview est pleine.
func teeChannel(ctx context.Context, src chan contracts.DataRow, blockID string, ps *contracts.PreviewStore) chan contracts.DataRow {
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

	// Construire le DAG.
	dag, err := BuildDAG(project)
	if err != nil {
		return nil, fmt.Errorf("executor: %w", err)
	}

	// Tri topologique.
	ordered, err := dag.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("executor: %w", err)
	}

	// Indexer les edges désactivés pour les sauter.
	disabledEdges := make(map[string]bool, len(project.Edges))
	for _, edge := range project.Edges {
		if edge.Disabled {
			disabledEdges[edge.From+"->"+edge.To] = true
		}
	}

	// Créer les ports (canaux) entre les blocs (sauf edges désactivés).
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

	// Exécuter chaque bloc dans l'ordre topologique.
	for _, node := range ordered {
		start := time.Now()
		result := RunResult{NodeID: node.ID}

		// Récupérer la factory du bloc depuis le registre.
		factory, ok := blocks.Registry[node.Type]
		if !ok {
			err := fmt.Errorf("bloc de type '%s' non enregistré", node.Type)
			result.Err = err
			report.Results = append(report.Results, result)
			report.EndedAt = time.Now()
			return report, err
		}
		block := factory()

		// Assembler les ports d'entrée.
		var inputPorts []*contracts.Port
		for predID, succs := range dag.adjacency {
			for _, succID := range succs {
				if succID == node.ID {
					key := predID + "->" + node.ID
					if p, ok := ports[key]; ok {
						inputPorts = append(inputPorts, p)
					}
				}
			}
		}

		// Assembler les ports de sortie avec tee pour la preview.
		var outputPorts []*contracts.Port
		for _, succ := range dag.Successors(node.ID) {
			key := node.ID + "->" + succ.ID
			if p, ok := ports[key]; ok {
				// Wrapper le canal avec tee : chaque ligne est capturée dans preview
				p.Ch = teeChannel(ctx, p.Ch, node.ID, preview)
				outputPorts = append(outputPorts, p)
			}
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
