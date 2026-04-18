package connections

// Connection représente une connexion réutilisable entre projets.
type Connection struct {
	ID   string                 `xml:"id,attr" json:"id"`
	Name string                 `xml:"name,attr" json:"name"`
	Type string                 `xml:"type,attr" json:"type"` // postgres, mysql, mssql, rest
	Envs map[string]ConnEnv     `xml:"-" json:"envs"`
	EnvList []ConnEnv           `xml:"environments>env" json:"-"`
}

// ConnEnv est un profil d'environnement d'une connexion.
type ConnEnv struct {
	Name      string `xml:"name,attr" json:"name"`
	Host      string `xml:"host,attr" json:"host"`
	Port      int    `xml:"port,attr" json:"port"`
	Database  string `xml:"db,attr" json:"database"`
	User      string `xml:"user,attr" json:"user"`
	SecretRef string `xml:"secretRef,attr" json:"secretRef"` // ex: ${DB_PASSWORD} ou vault:secret/crm
}

// DSN retourne la chaîne de connexion pour PostgreSQL.
func (e *ConnEnv) DSN(password string) string {
	return "host=" + e.Host +
		" port=" + itoa(e.Port) +
		" dbname=" + e.Database +
		" user=" + e.User +
		" password=" + password +
		" sslmode=disable"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 10)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
