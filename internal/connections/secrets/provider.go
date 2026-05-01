package secrets

import "fmt"

// Provider résout un secret à partir d'une référence arbitraire.
// Les implémentations concrètes permettent d'étendre le système
// sans modifier le resolver existant.
type Provider interface {
	Resolve(ref string) (string, error)
}

// EnvProvider résout les secrets via os.Getenv et les valeurs brutes.
// C'est l'implémentation par défaut (délègue à Resolve).
type EnvProvider struct{}

func (EnvProvider) Resolve(ref string) (string, error) {
	return Resolve(ref)
}

// VaultProvider est le stub pour une future intégration HashiCorp Vault.
// TODO: implémenter via GET {Address}/v1/{path} avec header X-Vault-Token.
type VaultProvider struct {
	Address string // ex: "https://vault.example.com"
	Token   string // X-Vault-Token
}

func (v VaultProvider) Resolve(ref string) (string, error) {
	// Implémentation future :
	//   path := strings.TrimPrefix(ref, "vault:")
	//   resp, err := http.Get(v.Address + "/v1/" + path) ...
	return "", fmt.Errorf("vault: non encore implémenté (ref: %s)", ref)
}
