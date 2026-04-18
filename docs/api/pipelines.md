# API Pipelines

## Create pipeline
```http
POST /api/v1/pipelines
Content-Type: application/json

{
  "name": "import_customers",
  "description": "Import CSV vers PostgreSQL",
  "source_type": "csv",
  "target_type": "postgres"
}
```

## Update pipeline
```http
PUT /api/v1/pipelines/{pipelineID}
Content-Type: application/json

{
  "name": "import_customers_v2",
  "description": "Import nightly CSV vers PostgreSQL",
  "status": "ready",
  "source_type": "csv",
  "target_type": "postgres"
}
```
