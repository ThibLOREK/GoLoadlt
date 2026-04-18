package security

import (
	"errors"
	"os"
	"strings"
)

// SecretStore résout les secrets depuis l'environnement.
// Supporte les préfixes :
//   - "env:VAR_NAME"   → os.Getenv(VAR_NAME)
//   - "file:/path"     → contenu du fichier
//   - valeur brute     → utilisée directement (dev uniquement)
type SecretStore struct{}

func NewSecretStore() *SecretStore { return &SecretStore{} }

func (s *SecretStore) Resolve(ref string) (string, error) {
	if strings.HasPrefix(ref, "env:") {
		key := strings.TrimPrefix(ref, "env:")
		val := os.Getenv(key)
		if val == "" {
			return "", errors.New("secret: env var " + key + " is empty")
		}
		return val, nil
	}

	if strings.HasPrefix(ref, "file:") {
		path := strings.TrimPrefix(ref, "file:")
		data, err := os.ReadFile(path)
		if err != nil {
			return "", errors.New("secret: cannot read file " + path)
		}
		return strings.TrimSpace(string(data)), nil
	}

	// Valeur brute — acceptable en dev, à interdire en prod via lint
	return ref, nil
}
