package engine

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// InjectResolvedConnections parcourt les nodes du projet et injecte les DSN
// dans les params des blocs qui référencent une connexion partagée.
func InjectResolvedConnections(
	project *contracts.Project,
	resolve func(connID string) (*resolver.ResolvedConn, error),
) error {
	for i := range project.Nodes {
		node := &project.Nodes[i]
		if node.ConnectionRef == "" {
			continue
		}
		rc, err := resolve(node.ConnectionRef)
		if err != nil {
			return fmt.Errorf("node '%s': résolution connexion '%s': %w", node.ID, node.ConnectionRef, err)
		}
		ensureParam(&node.Params, "dsn", rc.DSN)
		ensureParam(&node.Params, "db_type", rc.Type)
		ensureParam(&node.Params, "host", rc.Host)
		ensureParam(&node.Params, "database", rc.Database)
	}
	return nil
}

func ensureParam(params *[]contracts.Param, key, value string) {
	for i := range *params {
		if (*params)[i].Name == key {
			(*params)[i].Value = value
			return
		}
	}
	*params = append(*params, contracts.Param{Name: key, Value: value})
}
