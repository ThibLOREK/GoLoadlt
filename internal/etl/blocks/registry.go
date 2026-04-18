package blocks

import "github.com/rinjold/go-etl-studio/internal/etl/contracts"

// Registry est le catalogue global de tous les blocs disponibles.
// Clé : type du bloc (ex: "source.csv"), Valeur : factory.
var Registry = map[string]contracts.BlockFactory{}

// Register enregistre un bloc dans le catalogue.
// À appeler depuis les fonctions init() de chaque package de bloc.
func Register(blockType string, factory contracts.BlockFactory) {
	Registry[blockType] = factory
}
