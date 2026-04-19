package contracts

import "sync"

// PreviewStore capture les N premières lignes de sortie par bloc.
// Thread-safe : peut être écrit depuis plusieurs goroutines simultanément.
type PreviewStore struct {
	mu      sync.RWMutex
	data    map[string][]DataRow
	maxRows int
}

// NewPreviewStore crée un PreviewStore avec une limite de maxRows lignes par bloc.
func NewPreviewStore(maxRows int) *PreviewStore {
	return &PreviewStore{
		data:    make(map[string][]DataRow),
		maxRows: maxRows,
	}
}

// Append ajoute une ligne au preview d'un bloc (ignorée si le max est atteint).
func (ps *PreviewStore) Append(blockID string, row DataRow) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if len(ps.data[blockID]) >= ps.maxRows {
		return
	}
	// Deep copy pour éviter les data races si la row est réutilisée
	cp := make(DataRow, len(row))
	for k, v := range row {
		cp[k] = v
	}
	ps.data[blockID] = append(ps.data[blockID], cp)
}

// All retourne une copie de toutes les previews indexées par blockID.
func (ps *PreviewStore) All() map[string][]DataRow {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	out := make(map[string][]DataRow, len(ps.data))
	for k, v := range ps.data {
		out[k] = v
	}
	return out
}
