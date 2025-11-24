go run ./cmd/daemon

go run ./cmd/bastion

To persist data in PostgreSQL, set `BASTION_DB_DSN` (e.g. `postgres://user:pass@localhost:5432/bastion?sslmode=disable`) before starting `cmd/bastion`. Environment variables are also loaded from a `.env` file in the repo root when present:

```
BASTION_DB_DSN=postgres://user:pass@localhost:5432/bastion?sslmode=disable
DAEMON_URL=http://localhost:9090
```

