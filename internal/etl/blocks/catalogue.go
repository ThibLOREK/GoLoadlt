package blocks

import "sort"

// Catalogue retourne la liste descriptive des blocs disponibles pour l'UI.
func Catalogue() []map[string]any {
	items := []map[string]any{
		meta("source.csv", "input", "CSV Input", "Lit un fichier CSV", 0, 0, 1, 1),
		meta("source.postgres", "input", "PostgreSQL Input", "Exécute une requête SQL sur PostgreSQL", 0, 0, 1, 1),
		meta("source.mysql", "input", "MySQL Input", "Exécute une requête SQL sur MySQL", 0, 0, 1, 1),
		meta("source.mssql", "input", "MSSQL Input", "Exécute une requête SQL sur SQL Server", 0, 0, 1, 1),

		meta("transform.select", "transform", "Select Columns", "Sélectionne un sous-ensemble de colonnes", 1, 1, 1, 1),
		meta("transform.filter", "transform", "Filter Rows", "Filtre les lignes selon une condition", 1, 1, 1, 1),
		meta("transform.cast", "transform", "Cast Type", "Convertit le type d'une colonne", 1, 1, 1, 1),
		meta("transform.add_column", "transform", "Add Column", "Ajoute une colonne calculée", 1, 1, 1, 1),
		meta("transform.join", "transform", "Join", "Joint deux flux sur une clé", 2, 2, 1, 1),
		meta("transform.split", "transform", "Split", "Sépare un flux en plusieurs sorties", 1, 1, 2, 10),
		meta("transform.aggregate", "transform", "Aggregate", "Agrège des lignes par groupe", 1, 1, 1, 1),
		meta("transform.sort", "transform", "Sort", "Trie un flux", 1, 1, 1, 1),
		meta("transform.dedup", "transform", "Deduplicate", "Supprime les doublons", 1, 1, 1, 1),
		meta("transform.pivot", "transform", "Pivot", "Pivote des lignes en colonnes", 1, 1, 1, 1),
		meta("transform.unpivot", "transform", "Unpivot", "Transforme des colonnes en lignes", 1, 1, 1, 1),

		meta("target.csv", "output", "CSV Output", "Écrit le flux dans un fichier CSV", 1, 1, 0, 0),
		meta("target.postgres", "output", "PostgreSQL Output", "Insère le flux dans PostgreSQL", 1, 1, 0, 0),
		meta("target.browse", "output", "Browse", "Affiche un aperçu du flux", 1, 1, 0, 0),
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i]["type"].(string) < items[j]["type"].(string)
	})
	return items
}

func meta(t, c, l, d string, minIn, maxIn, minOut, maxOut int) map[string]any {
	return map[string]any{
		"type":        t,
		"category":    c,
		"label":       l,
		"description": d,
		"minInputs":   minIn,
		"maxInputs":   maxIn,
		"minOutputs":  minOut,
		"maxOutputs":  maxOut,
	}
}
