# Grafana dashboards sample instructions

Go definitions of Grafana dashboards built with grafana-foundation-sdk and
rendered as dashboardv2 JSON. `README.md` owns the dashboard conventions;
follow it instead of restating it here.

## Working agreements

- Communicate in Japanese. Write `README.md`, `docs/`, and code comments in
  Japanese; use English for identifiers and other documentation. Use ASCII
  filenames.
- Keep comments and documentation proportional to the change. Comment only a
  non-obvious reason, invariant, or hazard. Replace outdated text instead of
  appending history, and update an existing paragraph before adding a section.
- Update `README.md` and this file when the structure or commands change.

## Layout

- `cmd/generate/`: the single `main` package. `main.go` builds, validates,
  and writes the one dashboard defined in `dns_sample.go`.
  `helpers.go` and `validate.go` hold shared conventions and invariants.
  Builders call the SDK directly: `Panel(name, ...)` registers each element
  and `RowsLayout` places it by that name; queries use `QueryGroupBuilder`
  and `TargetBuilder`.
- `dashboards/`: generated JSON, mounted into local Grafana. It is gitignored;
  regenerate it instead of editing it.
- `provisioning/` and `docker-compose.yml`: local Grafana.
- `mise.toml`: pinned Go, uv, and podman-compose; `setup-linux.sh` installs
  only OS packages (Podman, make) and runs `mise install`.
- `docs/walkthrough.md`: step-by-step explanation of the code. Update it when the
  code it quotes changes.
- `generate` and `cmd/generate/generate` are local `go build` artifacts.

## Commands

- `go test ./...`: structural dashboard invariants.
- `make check`: diff a fresh generation against `dashboards/`. Run it with
  `go test` after every builder change.
- `make generate`: write JSON to `dashboards/`.
- `make dev` / `make dev-stop`: local Grafana at <http://localhost:3000> via
  Docker Compose or podman-compose (`COMPOSE=...` overrides the choice).
- Run `gofmt` on changed Go files.

## Safety

- `.env` holds the Prometheus URL. Do not print or commit it.
- Local Grafana grants anonymous Admin. Do not expose it beyond localhost.
