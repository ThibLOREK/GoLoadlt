# Exemple : CSV → PostgreSQL avec transformations

## Créer le pipeline
```http
POST /api/v1/pipelines
Content-Type: application/json

{
  "name": "import_clients",
  "description": "Import CSV clients vers PostgreSQL avec cast et filtre",
  "source_type": "csv",
  "target_type": "postgres",
  "source_config": {
    "file_path": "/data/clients.csv",
    "delimiter": ",",
    "has_header": true
  },
  "target_config": {
    "schema": "public",
    "table_name": "clients",
    "batch_size": 500
  },
  "steps": [
    {
      "type": "mapper",
      "config": {
        "mapping": {
          "prenom": "first_name",
          "nom": "last_name",
          "age": "age"
        }
      }
    },
    {
      "type": "filter",
      "config": {
        "rules": [
          {"column": "age", "operator": "gt", "value": "0"}
        ]
      }
    },
    {
      "type": "caster",
      "config": {
        "rules": [
          {"column": "age", "cast_to": "int"}
        ]
      }
    }
  ]
}
```

## Lancer le pipeline
```http
POST /api/v1/pipelines/{pipelineID}/runs
```

## Suivre l'exécution
```http
GET /api/v1/pipelines/{pipelineID}/runs/{runID}
```
