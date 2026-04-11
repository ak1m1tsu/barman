# barman

Discord-бот на Go с Clean Architecture. Управляет авто-ролью при вступлении участников на сервер и предоставляет базовые slash-команды.

## Команды

| Команда | Описание | Права |
|---------|----------|-------|
| `/ping` | Проверить задержку бота | все |
| `/help` | Список доступных команд | все |
| `/userinfo [пользователь]` | Информация о пользователе | все |
| `/autorole set <роль>` | Установить авто-роль для новых участников | Manage Roles |
| `/autorole remove` | Удалить авто-роль | Manage Roles |
| `/autorole info` | Показать текущую авто-роль | Manage Roles |

## Быстрый старт

```bash
# Скопировать шаблон конфига
cp configs/config.example.yaml configs/config.yaml
# Заполнить token, app_id в configs/config.yaml

# Запустить через Docker Compose
make up

# Или собрать и запустить напрямую
make build
./bin/bot --config configs/config.yaml
```

## Конфигурация

```yaml
# configs/config.yaml
discord:
  token: "YOUR_BOT_TOKEN"
  app_id: "YOUR_APP_ID"
  guild_id: ""        # оставить пустым для глобальных команд

database:
  path: "barman.db"   # путь к SQLite-файлу
```

`configs/config.yaml` добавлен в `.gitignore` и не попадает в репозиторий.

## Разработка

```bash
make test          # запустить все тесты
make lint          # golangci-lint
make mock          # пересгенерировать моки (после изменения интерфейсов)
make build         # собрать бинарь в bin/bot
make docker-build  # собрать Docker-образ
```

## Архитектура

Clean Architecture — зависимости направлены строго внутрь:

```
infrastructure → adapter → usecase → domain
```

```
internal/
├── domain/guild/          # сущность Guild, интерфейс Repository
├── usecase/
│   ├── guild/             # SetAutoRole, GetAutoRole, RemoveAutoRole
│   └── member/            # AssignAutoRole, интерфейс RoleAssigner
├── adapter/
│   ├── command/           # slash-команды (discordgo)
│   ├── handler/           # обработчик GuildMemberAdd
│   └── repository/sqlite/ # реализация Repository через SQLite
└── infrastructure/
    ├── config/            # загрузка YAML-конфига
    ├── database/          # открытие SQLite, миграции
    └── discord/           # discordgo session, RoleAssigner
```

Моки генерируются через [mockery](https://github.com/vektra/mockery) (`make mock`) и закоммичены в репозиторий.

## CI/CD

GitHub Actions pipeline при каждом push:

```
build → lint ┐
             ├─ параллельно
       test  ┤
             ├─ параллельно
  dep_check ─┘
       └── deploy  (только master → VPS по SSH)
```

- **build** — собирает Docker-образ, пушит в GHCR с тегом `{sha7}-{YYYYMMDD}` (master) или `{branch}-{sha7}` (другие ветки)
- **lint** — `golangci-lint`
- **test** — `go test ./...`
- **dependency_check** — `govulncheck`
- **deploy** — `docker compose pull && up -d` на VPS

### Необходимые секреты репозитория

| Секрет | Описание |
|--------|----------|
| `GITHUB_TOKEN` | Встроенный, создавать не нужно |
| `VPS_HOST` | IP или домен VPS |
| `VPS_USER` | SSH-пользователь |
| `SSH_PRIVATE_KEY` | Приватный SSH-ключ |

## Стек

- [discordgo](https://github.com/bwmarrin/discordgo) — Discord API
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) — pure Go SQLite (без CGO)
- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) — YAML конфиг
- [testify](https://github.com/stretchr/testify) + [mockery](https://github.com/vektra/mockery) — тесты и моки
- [golangci-lint](https://github.com/golangci/golangci-lint) — линтер