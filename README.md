# volumevault_exporter

Exporter Prometheus en Go pour l'API VolumeVault.

## Variables d'environnement

- `VOLUMEVAULT_API_URL` (obligatoire): URL de base de l'API, ex. `https://volumevault.example.com/api/v1`
- `VOLUMEVAULT_API_TOKEN` (obligatoire): token ****** capacité `read`
- `LISTEN_ADDRESS` (optionnel): adresse d'écoute, défaut `:9780`
- `METRICS_PATH` (optionnel): chemin des métriques, défaut `/metrics`
- `VOLUMEVAULT_TIMEOUT` (optionnel): timeout HTTP, défaut `10s`

## Exécution locale

```bash
go run ./cmd/volumevault_exporter \
  --volumevault.api-url="https://volumevault.example.com/api/v1" \
  --volumevault.api-token="<TOKEN>"
```

## Docker

```bash
docker build -t ghcr.io/nouchka/volumevault_exporter:local .
```

## Docker Compose (test)

```bash
VOLUMEVAULT_API_URL=https://volumevault.example.com/api/v1 \
VOLUMEVAULT_API_TOKEN=<TOKEN> \
  docker compose up --build
```

Puis accéder à `http://localhost:9780/metrics`.

## Métriques exposées

- `volumevault_up`
- `volumevault_scrape_duration_seconds`
- `volumevault_volumes_total`
- `volumevault_volume_backup_state{volume,state}`
- `volumevault_volume_last_backup_timestamp_seconds{volume}`
- `volumevault_volume_last_backup_size_bytes{volume}`
- `volumevault_backup_runs_total{status}`
- `volumevault_restore_runs_total{status}`
- `volumevault_backup_run_duration_seconds{run_id,status}`
