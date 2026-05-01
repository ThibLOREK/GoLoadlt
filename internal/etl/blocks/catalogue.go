package blocks

import "sort"

// ParamDef décrit un paramètre d'un bloc pour l'UI.
type ParamDef struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`        // text | select | column-select | column-multiselect | checkbox
	Default     string   `json:"default,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Options     []string `json:"options,omitempty"` // pour type=select
	Description string   `json:"description,omitempty"`
	// Pour column-select et column-multiselect, le frontend peuple
	// les options depuis les colonnes du port d'entrée du bloc.
}

// Catalogue retourne la liste descriptive des blocs disponibles pour l'UI.
func Catalogue() []map[string]any {
	items := []map[string]any{
		// --- Sources ---
		meta("source.csv", "input", "CSV Input", "Lit un fichier CSV", 0, 0, 1, 1, nil),
		meta("source.postgres", "input", "PostgreSQL Input", "Exécute une requête SQL sur PostgreSQL", 0, 0, 1, 1, nil),
		meta("source.mysql", "input", "MySQL Input", "Exécute une requête SQL sur MySQL", 0, 0, 1, 1, nil),
		meta("source.mssql", "input", "MSSQL Input", "Exécute une requête SQL sur SQL Server", 0, 0, 1, 1, nil),
		meta("source.data_grid", "input", "Data Grid", "Table statique définie en dur dans le pipeline — parité complète Pentaho PDI Data Grid (columns + rows JSON)", 0, 0, 1, 1, nil),

		// --- Transforms ---
		meta("transform.dummy", "transform", "Dummy", "Laisse passer les données sans modification — utile pour observer le flux (like Pentaho PDI Dummy)", 1, 1, 1, 1, nil),
		meta("transform.filter_advanced", "transform", "Filter (If/Else)", "Filtre les lignes avec des conditions if/else sur texte, nombre ou booléen — 2 sorties: true + false", 1, 1, 1, 2, nil),
		meta("transform.select", "transform", "Select Columns", "Sélectionne un sous-ensemble de colonnes", 1, 1, 1, 1, nil),
		meta("transform.filter", "transform", "Filter Rows", "Filtre les lignes selon une condition simple", 1, 1, 1, 1, nil),
		meta("transform.cast", "transform", "Cast Type", "Convertit le type d'une colonne", 1, 1, 1, 1, nil),
		meta("transform.add_column", "transform", "Add Column", "Ajoute une colonne calculée", 1, 1, 1, 1, nil),
		meta("transform.join", "transform", "Join", "Joint deux flux sur une clé", 2, 2, 1, 1, nil),
		meta("transform.merge", "transform", "Merge", "Fusionne deux flux — parité complète pd.merge (how, on, left_on, right_on, suffixes, validate)", 2, 2, 1, 1, nil),
		meta("transform.groupby", "transform", "GroupBy", "Agrège un flux par groupe — parité complète df.groupby (by, sort, as_index, dropna, SUM/COUNT/AVG/MIN/MAX/MEDIAN/NUNIQUE/STD/VAR)", 1, 1, 1, 1, nil),
		meta("transform.fillna", "transform", "Fill NA", "Remplace les valeurs nulles — parité complète df.fillna (value, method ffill/bfill, columns, limit)", 1, 1, 1, 1, nil),
		meta("transform.rename", "transform", "Rename Columns", "Renomme des colonnes — parité complète df.rename (columns mapping ancien:nouveau, errors ignore/raise)", 1, 1, 1, 1, nil),
		meta("transform.drop_duplicates", "transform", "Drop Duplicates", "Supprime les doublons — parité complète df.drop_duplicates (subset, keep first/last/false, ignore_index)", 1, 1, 1, 1, nil),
		meta("transform.split", "transform", "Split", "Sépare un flux en plusieurs sorties", 1, 1, 2, 10, nil),
		meta("transform.aggregate", "transform", "Aggregate", "Agrège des lignes par groupe", 1, 1, 1, 1, nil),
		meta("transform.sort", "transform", "Sort", "Trie un flux", 1, 1, 1, 1, nil),
		meta("transform.dedup", "transform", "Deduplicate",
			"Supprime les doublons sur une ou plusieurs clés — liste déroulante alimentée par les colonnes du flux d'entrée",
			1, 1, 1, 1,
			[]ParamDef{
				{
					Name:        "keys",
					Label:       "Colonnes clés de dédoublonnement",
					Type:        "column-multiselect",
					Required:    false,
					Description: "Sélectionnez une ou plusieurs colonnes. Laissez vide pour dédoublonner sur TOUTES les colonnes.",
				},
			},
		),
		meta("transform.pivot", "transform", "Pivot", "Pivote des lignes en colonnes", 1, 1, 1, 1, nil),
		meta("transform.unpivot", "transform", "Unpivot", "Transforme des colonnes en lignes", 1, 1, 1, 1, nil),

		// --- Blocs bonus Sprint E ---
		meta("transform.union", "transform", "Union", "Fusionne deux flux ou plus en un seul (UNION ALL)", 2, 10, 1, 1, nil),
		meta("transform.regex", "transform", "Regex Extract", "Extrait des groupes via une regex sur une colonne (modes: extract, replace, match)", 1, 1, 1, 1, nil),
		meta("transform.find_replace", "transform", "Find & Replace", "Remplace des valeurs dans une colonne (modes: exact, contains, regex)", 1, 1, 1, 1, nil),
		meta("transform.sampling", "transform", "Sampling", "Échantillonne le flux : N premières lignes, % aléatoire, ou 1 ligne sur N", 1, 1, 1, 1, nil),
		meta("transform.text_to_columns", "transform", "Text to Columns", "Découpe une colonne texte en plusieurs colonnes via un délimiteur", 1, 1, 1, 1, nil),
		meta("transform.auto_field", "transform", "Auto Field", "Détecte et convertit automatiquement les types de colonnes (int, float, bool, string)", 1, 1, 1, 1, nil),
		meta("transform.append_fields", "transform", "Append Fields", "Fusionne horizontalement deux flux ligne par ligne (colonnes en conflit préfixées 'right_')", 2, 2, 1, 1, nil),
		meta("transform.data_cleansing", "transform", "Data Cleansing", "Nettoie les données : trim, casse, suppression caractères spéciaux, nullification des vides", 1, 1, 1, 1, nil),
		meta("transform.datetime", "transform", "DateTime Transform", "Parse, formate, ajoute une durée ou extrait une composante d'une colonne date/heure", 1, 1, 1, 1, nil),

		// --- Targets ---
		meta("target.csv", "output", "CSV Output", "Écrit le flux dans un fichier CSV", 1, 1, 0, 0, nil),
		meta("target.postgres", "output", "PostgreSQL Output", "Insère le flux dans PostgreSQL", 1, 1, 0, 0, nil),
		meta("target.browse", "output", "Browse", "Affiche un aperçu du flux", 1, 1, 0, 0, nil),
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i]["type"].(string) < items[j]["type"].(string)
	})
	return items
}

func meta(t, c, l, d string, minIn, maxIn, minOut, maxOut int, params []ParamDef) map[string]any {
	m := map[string]any{
		"type":        t,
		"category":    c,
		"label":       l,
		"description": d,
		"minInputs":   minIn,
		"maxInputs":   maxIn,
		"minOutputs":  minOut,
		"maxOutputs":  maxOut,
	}
	if len(params) > 0 {
		m["paramSchema"] = params
	}
	return m
}
