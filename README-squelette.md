# GoLoadIt — État du projet

## Vision
Plateforme ETL visuelle en Go, inspirée d'Alteryx et Talend, où chaque projet est modélisé
comme un graphe de blocs interconnectés (nodes), persisté en XML côté serveur, avec des
connexions réutilisables et switchables par environnement (Dev / Préprod / Prod).

## Fonctionnalités implémentées
- CRUD Pipeline complet avec persistance PostgreSQL
- Modèle Run : création, statuts, compteurs
- API runs : schedule, list, get
- Worker polling les runs `pending` toutes les 5 secondes
- Connecteur source CSV
- Connecteur target PostgreSQL (batch insert avec transaction)
- Moteur ETL Executor (extract → transform → load)

## Nouveaux concepts clés

### Projet ETL = graphe de blocs XML
Chaque projet créé dans GoLoadIt génère un fichier XML côté serveur (style Jenkins jobs).
Ce fichier décrit :
- Les blocs (nodes) : source, transformation, destination
- Les liens (edges) entre blocs
- Les paramètres de chaque bloc
- La version du projet et son historique

Exemple de structure XML d'un projet :
```xml
<project id="proj-001" name="Sales ETL" version="3">
  <nodes>
    <node id="n1" type="source.postgres" connectionRef="conn-crm-prod">
      <param name="query">SELECT * FROM orders WHERE date > :start</param>
    </node>
    <node id="n2" type="transform.filter">
      <param name="condition">amount > 100</param>
    </node>
    <node id="n3" type="transform.pivot">
      <param name="groupBy">region</param>
      <param name="aggregation">SUM(amount)</param>
    </node>
    <node id="n4" type="transform.add_column">
      <param name="name">tax</param>
      <param name="expression">amount * 0.20</param>
    </node>
    <node id="n5" type="target.csv">
      <param name="path">/exports/sales_pivot.csv</param>
    </node>
  </nodes>
  <edges>
    <edge from="n1" to="n2"/>
    <edge from="n2" to="n3"/>
    <edge from="n3" to="n4"/>
    <edge from="n4" to="n5"/>
  </edges>
</project>
```

### Connexions réutilisables et multi-environnements
Une connexion définie une fois est disponible dans tous les projets.
Chaque connexion possède plusieurs profils d'environnement (Dev, Préprod, Prod).
Le switch d'environnement se fait globalement sans modifier les projets.

```xml
<connection id="conn-crm" name="CRM Database" type="postgres">
  <environments>
    <env name="dev"     host="localhost"     port="5432" db="crm_dev"    user="dev_user"  secretRef="vault:crm-dev"/>
    <env name="preprod" host="preprod-db"    port="5432" db="crm_preprod" user="pp_user"  secretRef="vault:crm-pp"/>
    <env name="prod"    host="prod-db.internal" port="5432" db="crm_prod" user="prod_user" secretRef="vault:crm-prod"/>
  </environments>
</connection>
```

### Blocs de transformation disponibles (MVP)
| Bloc | Type interne | Description |
|---|---|---|
| Source CSV | `source.csv` | Lecture d'un fichier CSV |
| Source PostgreSQL | `source.postgres` | Requête SQL sur une connexion référencée |
| Source MySQL | `source.mysql` | Requête SQL MySQL |
| Source SQL Server | `source.mssql` | Requête SQL Server |
| Filtre | `transform.filter` | Filtrer les lignes selon une condition |
| Mapping colonnes | `transform.select` | Renommer / sélectionner des colonnes |
| Cast de type | `transform.cast` | Convertir le type d'une colonne |
| Colonne calculée | `transform.add_column` | Ajouter une colonne avec une expression |
| Split | `transform.split` | Diviser le flux en plusieurs sorties selon une condition |
| Pivot | `transform.pivot` | Pivoter des colonnes (GROUP BY + agrégation) |
| Dépivot | `transform.unpivot` | Transformer des colonnes en lignes |
| Jointure | `transform.join` | Joindre deux flux de données |
| Déduplication | `transform.dedup` | Supprimer les doublons |
| Tri | `transform.sort` | Trier les lignes |
| Agrégation | `transform.aggregate` | SUM, COUNT, AVG, MIN, MAX par groupe |
| Cible PostgreSQL | `target.postgres` | Insertion/upsert dans PostgreSQL |
| Cible CSV | `target.csv` | Écriture dans un fichier CSV |
| Cible API REST | `target.rest` | POST vers une API HTTP |

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
- `GET    /api/v1/connections`
- `POST   /api/v1/connections`
- `PUT    /api/v1/connections/{connID}`
- `DELETE /api/v1/connections/{connID}`
- `POST   /api/v1/connections/{connID}/test`
- `PUT    /api/v1/environment` — switch global d'environnement
- `GET    /api/v1/projects/{projectID}/xml` — export XML du projet
- `POST   /api/v1/projects/import` — import XML d'un projet

## Prochaines étapes
- Implémenter le parser XML → DAG de blocs exécutables
- Ajouter les blocs `split`, `pivot`, `add_column`, `join` dans le moteur
- Développer le gestionnaire de connexions multi-env
- Initier l'interface visuelle React Flow avec palette de blocs
- Brancher le switch d'environnement global sur l'API