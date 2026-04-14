# barman

Discord bot written in Go using Clean Architecture. Manages auto-role assignment for new server members, provides slash commands, and sends anime reaction GIFs (25 reaction types) sourced from nekos.best with otakugifs.xyz as a parallel fallback.

[Русский](README.ru.md)

## Commands

| Command | Description | Permissions |
|---------|-------------|-------------|
| `/ping` | Check bot latency | everyone |
| `/help` | List available commands | everyone |
| `/userinfo [user]` | Show user information | everyone |
| `/autorole` | Manage auto-role for new members (interactive) | Manage Roles |
| `/react <type> [user]` | Send an anime reaction GIF | everyone |
| `/reactions` | List all available reaction types | everyone |
| `/prefix` | View and change the server command prefix (interactive) | Manage Server |
| `<prefix><type> [@user]` | Send a reaction via prefix (reply auto-targets the replied-to user) | everyone |

## Quick Start

```bash
# Copy config template
cp configs/config.example.yaml configs/config.yaml
# Fill in token and app_id in configs/config.yaml

# Run via Docker Compose
make up

# Or build and run directly
make build
./bin/bot --config configs/config.yaml
```

## Configuration

```yaml
# configs/config.yaml
discord:
  token: "YOUR_BOT_TOKEN"
  app_id: "YOUR_APP_ID"
  guild_id: ""        # leave empty for global commands
  prefix: "!"         # default prefix for message commands

database:
  path: "barman.db"   # path to SQLite file
```

`configs/config.yaml` is listed in `.gitignore` and is never committed to the repository.

## Development

```bash
make test          # run all tests
make lint          # golangci-lint
make mock          # regenerate mocks (after changing interfaces)
make build         # build binary to bin/bot
make docker-build  # build Docker image
```

## Architecture

Clean Architecture — dependencies point strictly inward:

```
infrastructure → adapter → usecase → domain
```

```
internal/
├── domain/guild/          # Guild entity, Repository interface
├── usecase/
│   ├── guild/             # SetAutoRole, GetAutoRole, RemoveAutoRole, SetPrefix, GetPrefix, RemovePrefix
│   ├── member/            # AssignAutoRole, RoleAssigner interface
│   └── reaction/          # FetchGIFUseCase, FetchGIFWithFallbackUseCase, GIFFetcher/GIFExecutor interfaces
├── adapter/
│   ├── command/           # slash commands (discordgo)
│   ├── handler/           # GuildMemberAdd, MessageCreate, interaction handlers
│   └── repository/sqlite/ # Repository implementation via SQLite
└── infrastructure/
    ├── config/            # YAML config loading
    ├── database/          # SQLite open
    ├── discord/           # discordgo session, RoleAssigner
    ├── nekos/             # nekos.best HTTP client (primary GIF source)
    └── otakugifs/         # otakugifs.xyz HTTP client (fallback GIF source)
```

### Database migrations

Migrations live in `migrations/` and are applied manually on the server:

```bash
sqlite3 barman.db < migrations/000001_init_guild_settings.up.sql
```

Mocks are generated via [mockery](https://github.com/vektra/mockery) (`make mock`) and committed to the repository.

## CI/CD

GitHub Actions pipeline on every push:

```
build (go build → artifact)
  ├── lint
  ├── test
  └── dependency_check
        └── docker (builds & pushes image to GHCR)
                └── deploy  (main branch only → VPS via SSH)
```

- **build** — compiles the binary with `CGO_ENABLED=0`, uploads as a workflow artifact
- **lint** — `golangci-lint`
- **test** — `go test ./...`
- **dependency_check** — `govulncheck`
- **docker** — downloads the artifact, builds a minimal Docker image and pushes to GHCR tagged `{sha7}-{YYYYMMDD}` (main) or `{branch}-{sha7}` (other branches)
- **deploy** — `docker compose pull && down && up -d` on VPS

### Required Repository Secrets

| Secret | Description |
|--------|-------------|
| `GITHUB_TOKEN` | Built-in, no setup needed |
| `VPS_HOST` | VPS IP or domain |
| `VPS_USER` | SSH user |
| `VPS_PASSWORD` | SSH password |
| `BOT_TOKEN` | Discord bot token |
| `BOT_APP_ID` | Discord application ID |

## Stack

- [discordgo](https://github.com/bwmarrin/discordgo) — Discord API
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) — pure Go SQLite (no CGO)
- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) — YAML config
- [testify](https://github.com/stretchr/testify) + [mockery](https://github.com/vektra/mockery) — tests and mocks
- [golangci-lint](https://github.com/golangci/golangci-lint) — linter
- [logrus](https://github.com/sirupsen/logrus) — structured JSON logging
- [nekos.best](https://nekos.best) — primary anime reaction GIFs API
- [otakugifs.xyz](https://otakugifs.xyz) — fallback anime reaction GIFs API
