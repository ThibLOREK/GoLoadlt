# Scheduling cron

## Activer un schedule sur un pipeline
```http
PUT /api/v1/pipelines/{pipelineID}/schedule
Content-Type: application/json

{
  "cron_expr": "0 2 * * *",
  "enabled": true
}
```

## Expressions cron supportées (5 champs)
| Expr          | Signification           |
|---------------|-------------------------|
| `0 2 * * *`   | Tous les jours à 02h00  |
| `*/15 * * * *`| Toutes les 15 minutes   |
| `0 8 * * 1`   | Lundi à 08h00           |
| `0 0 1 * *`   | 1er du mois à minuit    |

## Consulter le schedule
```http
GET /api/v1/pipelines/{pipelineID}/schedule
```

## Supprimer le schedule
```http
DELETE /api/v1/pipelines/{pipelineID}/schedule
```
