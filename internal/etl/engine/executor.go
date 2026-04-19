package engine

import (
	"context"
	"fmt"
	"sync"
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

// previewPort wrape un Port de sortie pour capturer les lignes dans le PreviewStore
// sans modifier la topologie des canaux.
// Architecture : producteur → port.Ch (lu par previewPort.drain en goroutine)
//                           → capture preview
//                           → mirrorCh (lu par le consommateur)
//
// Le consommateur reçoit mirrorCh dans ses Inputs — il ne voit jamais rawCh.
type previewPort struct {
	raw    *contracts.Port // port "brut" que le producteur écrit
	mirror *contracts.Port // port "miroir" que le consommateur lit
}

func newPreviewPort(edgeKey string, ps *contracts.PreviewStore, blockID string, ctx context.Context) *previewPort {
	rawCh := make(chan contracts.DataRow, defaultChannelBuffer)
	mirrorCh := make(chan contracts.DataRow, defaultChannelBuffer)

	go func() {
		defer close(mirrorCh)
		for {
			select {
			case row, ok := <-rawCh:
				if !ok {
					return
				}
				ps.Append(blockID, row)
				select {
				case mirrorCh <- row:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return &previewPort{
		raw:    &contracts.Port{ID: edgeKey, Ch: rawCh},
		mirror: &contracts.Port{ID: edgeKey, Ch: mirrorCh},
	}
}

// Execute exécute un projet ETL complet.
//
// Phases :
//  1. BuildDAG + TopologicalSort
//  2. Câblage : créer UN canal par edge actif, indexer incoming/outgoing par nodeID
//  3. Exécution séquentielle (ordre topologique)
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

	// ── Phase 1 : câblage ────────────────────────────────────────────────────
	// Pour chaque edge actif, on crée un previewPort :
	//   - raw.Ch    → le producteur y écrit
	//   - mirror.Ch → le consommateur y lit (après passage dans la goroutine preview)
	//
	// incoming[nodeID] = liste des mirror ports à passer en Inputs
	// outgoing[nodeID] = liste des raw    ports à passer en Outputs
	incoming := make(map[string][]*contracts.Port)
	outgoing := make(map[string][]*contracts.Port)

	for _, edge := range project.Edges {
		if edge.Disabled {
			continue
		}
		if _, ok := dag.Nodes[edge.From]; !ok {
			continue
		}
		if _, ok := dag.Nodes[edge.To]; !ok {
			continue
		}

		key := edge.From + "->" + edge.To
		pp := newPreviewPort(key, preview, edge.From, ctx)

		outgoing[edge.From] = append(outgoing[edge.From], pp.raw)
		incoming[edge.To] = append(incoming[edge.To], pp.mirror)
	}

	// ── Phase 2 : exécution ──────────────────────────────────────────────────
	var mu sync.Mutex

	for _, node := range ordered {
		start := time.Now()
		result := RunResult{NodeID: node.ID}

		factory, ok := blocks.Registry[node.Type]
		if !ok {
			err := fmt.Errorf("bloc de type '%s' non enregistré", node.Type)
			mu.Lock()
			result.Err = err
			report.Results = append(report.Results, result)
			report.EndedAt = time.Now()
			mu.Unlock()
			return report, err
		}
		block := factory()

		bctx := &contracts.BlockContext{
			Ctx:       ctx,
			Params:    node.ParamMap(),
			ConnRef:   node.ConnRef,
			ActiveEnv: e.ActiveEnv,
			BlockID:   node.ID,
			Preview:   preview,
			Inputs:    incoming[node.ID],  // nil-safe : slice vide si source
			Outputs:   outgoing[node.ID],  // nil-safe : slice vide si target
		}

		e.log.Info().
			Str("node", node.ID).
			Str("type", node.Type).
			Int("inputs", len(bctx.Inputs)).
			Int("outputs", len(bctx.Outputs)).
			Msg("exécution bloc")

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
		e.log.Info().
			Str("node", node.ID).
			Dur("durée", result.Duration).
			Msg("bloc terminé")
	}

	report.EndedAt = time.Now()
	report.Success = true
	report.Preview = preview.All()
	return report, nil
}
