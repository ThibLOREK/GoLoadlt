package handlers

import (
	"net/http"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
)

// ListBlocks retourne le catalogue complet des blocs disponibles pour l'UI React Flow.
func ListBlocks() http.HandlerFunc {
	catalogue := blocks.Catalogue()
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, catalogue)
	}
}
