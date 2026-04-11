# barman

Discord bot written in Go using Clean Architecture. Manages auto-role assignment for new server members and provides basic slash commands.

[Русский](README.ru.md)

## Commands

| Command | Description | Permissions |
|---------|-------------|-------------|
| `/ping` | Check bot latency | everyone |
| `/help` | List available commands | everyone |
| `/userinfo [user]` | Show user information | everyone |
| `/autorole set <role>` | Set auto-role for new members | Manage Roles |
| `/autorole remove` | Remove auto-role | Manage Roles |
| `/autorole info` | Show current auto-role | Manage Roles |

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
│   ├── guild/             # SetAutoRole, GetAutoRole, RemoveAutoRole
│   └── member/            # AssignAutoRole, RoleAssigner interface
├── adapter/
│   ├── command/           # slash commands (discordgo)
│   ├── handler/           # GuildMemberAdd event handler
│   └── repository/sqlite/ # Repository implementation via SQLite
└── infrastructure/
    ├── config/            # YAML config loading
    ├── database/          # SQLite open & migrations
    └── discord/           # discordgo session, RoleAssigner
```

Mocks are generated via [mockery](https://github.com/vektra/mockery) (`make mock`) and committed to the repository.

## CI/CD

GitHub Actions pipeline on every push:

```
build → lint ┐
             ├─ parallel
       test  ┤
             ├─ parallel
  dep_check ─┘
       └── deploy  (main branch only → VPS via SSH)
```

- **build** — builds Docker image, pushes to GHCR tagged `{sha7}-{YYYYMMDD}` (main) or `{branch}-{sha7}` (other branches)
- **lint** — `golangci-lint`
- **test** — `go test ./...`
- **dependency_check** — `govulncheck`
- **deploy** — `docker compose pull && up -d` on VPS

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
