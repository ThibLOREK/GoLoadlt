package secrets

import (
	"fmt"
	"os"
	"strings"
)

// Resolve résout une référence de secret.
// Formats supportés :
//   - ${ENV_VAR}        → variable d'environnement
//   - vault:secret/path → non implémenté (extensible)
//   - texte brut        → retourné tel quel (déconseillé en prod)
func Resolve(ref string) (string, error) {
	if strings.HasPrefix(ref, "${") && strings.HasSuffix(ref, "}") {
		envKey := ref[2 : len(ref)-1]
		val := os.Getenv(envKey)
		if val == "" {
			return "", fmt.Errorf("secrets: variable d'environnement '%s' non définie", envKey)
		}
		return val, nil
	}
	if strings.HasPrefix(ref, "vault:") {
		return "", fmt.Errorf("secrets: intégration Vault non encore implémentée (ref: %s)", ref)
	}
	// Valeur brute (dev local uniquement)
	return ref, nil
}
