# Squelette initial Go ETL Studio

## Contenu généré
- backend Go minimal avec `server`, `worker` et `migrate`
- routeur HTTP Chi
- configuration centralisée
- logger Zerolog
- contrats ETL de base
- moteur ETL minimal
- DTO et modèles initiaux
- Docker Compose PostgreSQL + Redis
- exemple de migration SQL

## Démarrage
1. Copier `configs/.env.example` vers `.env`
2. Lancer `docker compose -f deploy/docker/docker-compose.yml up -d`
3. Exécuter `go mod tidy`
4. Lancer `go run ./cmd/server`

## Prochaine étape
Implémenter les entités métier complètes, le repository PostgreSQL, puis le CRUD pipeline persistant.
