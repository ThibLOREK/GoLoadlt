# Squelette initial Go ETL Studio

## Contenu généré
- backend Go minimal avec `server`, `worker` et `migrate`
- routeur HTTP Chi
- configuration centralisée
- logger Zerolog
- contrats ETL de base
- moteur ETL minimal
- Docker Compose PostgreSQL + Redis
- migration SQL enrichie pour `pipelines`
- repository PostgreSQL pour les pipelines
- service applicatif `PipelineService`
- CRUD HTTP complet pour les pipelines

## Endpoints disponibles
- `GET /health`
- `GET /api/v1/pipelines`
- `POST /api/v1/pipelines`
- `GET /api/v1/pipelines/{pipelineID}`
- `PUT /api/v1/pipelines/{pipelineID}`
- `DELETE /api/v1/pipelines/{pipelineID}`
- `POST /api/v1/pipelines/{pipelineID}/runs`

## Prochaine étape
Implémenter les `runs`, brancher un vrai worker, puis ajouter un premier connecteur source CSV et un loader PostgreSQL.
