# Go ETL Studio — État du projet

## Fonctionnalités implémentées
- CRUD Pipeline complet avec persistance PostgreSQL
- Modèle Run : création, statuts, compteurs
- API runs : schedule, list, get
- Worker polling les runs `pending` toutes les 5 secondes
- Connecteur source CSV
- Connecteur target PostgreSQL (batch insert avec transaction)
- Moteur ETL Executor (extract → transform → load)

## Endpoints disponibles
- `GET    /health`
- `GET    /api/v1/pipelines`
- `POST   /api/v1/pipelines`
- `GET    /api/v1/pipelines/{pipelineID}`
- `PUT    /api/v1/pipelines/{pipelineID}`
- `DELETE /api/v1/pipelines/{pipelineID}`
- `POST   /api/v1/pipelines/{pipelineID}/runs`
- `GET    /api/v1/pipelines/{pipelineID}/runs`
- `GET    /api/v1/pipelines/{pipelineID}/runs/{runID}`

## Prochaine étape suggérée
- Brancher un vrai Executor depuis la définition de pipeline dans le worker
- Ajouter un connecteur source PostgreSQL
- Implémenter les transformations de base (map, filter, cast)
- Ajouter scheduling cron
- Initier l'interface visuelle React + React Flow
