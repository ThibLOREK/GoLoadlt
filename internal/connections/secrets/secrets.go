package secrets

import (
	"fmt"
	"os"
	"strings"
)

// Resolve résout une référence de secret vers sa valeur.
// Formats supportés :
//   ${VAR_NAME}       → variable d'environnement
//   env:VAR_NAME      → variable d'environnement (syntaxe explicite)
//   plain:valeur      → valeur en clair (déconseillé, usage dev uniquement)
//   vault:path/secret → réservé pour une intégration Vault future
func Resolve(ref string) (string, error) {
	if ref == "" {
		return "", nil
	}

	// ${VAR_NAME}
	if strings.HasPrefix(ref, "${") && strings.HasSuffix(ref, "}") {
		varName := ref[2 : len(ref)-1]
		return resolveEnv(varName)
	}

	// env:VAR_NAME
	if strings.HasPrefix(ref, "env:") {
		return resolveEnv(strings.TrimPrefix(ref, "env:"))
	}

	// plain:valeur (dév uniquement)
	if strings.HasPrefix(ref, "plain:") {
		return strings.TrimPrefix(ref, "plain:"), nil
	}

	// vault:path (placeholder pour intégration future)
	if strings.HasPrefix(ref, "vault:") {
		return "", fmt.Errorf("secrets: intégration Vault non encore implémentée (ref: %s)", ref)
	}

	return "", fmt.Errorf("secrets: format de référence non reconnu: '%s'", ref)
}

func resolveEnv(varName string) (string, error) {
	val := os.Getenv(varName)
	if val == "" {
		return "", fmt.Errorf("secrets: variable d'environnement '%s' non définie ou vide", varName)
	}
	return val, nil
}
