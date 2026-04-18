package handlers

import (
	"net/http"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
)

type BlockInfo struct {
	Type        string `json:"type"`
	Category    string `json:"category"`
	Label       string `json:"label"`
	Description string `json:"description"`
	MinInputs   int    `json:"minInputs"`
	MaxInputs   int    `json:"maxInputs"`
	MinOutputs  int    `json:"minOutputs"`
	MaxOutputs  int    `json:"maxOutputs"`
}

// ListBlocks retourne le catalogue complet des blocs disponibles pour l'UI React Flow.
func ListBlocks() http.HandlerFunc {
	catalogue := blocks.Catalogue()
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, catalogue, http.StatusOK)
	}
}