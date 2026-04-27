# barman

[English](README.md)

Discord-бот на Go с Clean Architecture. Управляет авто-ролью при вступлении участников на сервер, предоставляет slash-команды и отправляет аниме-реакции в виде GIF (25 типов) — источником служит nekos.best с параллельным fallback через otakugifs.xyz.

## Команды

| Команда | Описание | Права |
|---------|----------|-------|
| `/ping` | Проверить задержку бота | все |
| `/help` | Список доступных команд | все |
| `/userinfo [пользователь]` | Информация о пользователе | все |
| `/autorole` | Управление авто-ролью для новых участников (интерактивно) | Manage Roles |
| `/react <тип> [пользователь]` | Отправить аниме-реакцию в виде GIF | все |
| `/reactions` | Список доступных реакций с описанием и счётчиком использования | все |
| `/prefix` | Просмотр и смена префикса команд сервера (интерактивно) | Manage Server |
| `<префикс><тип> [@пользователь]` | Отправить реакцию через префикс (reply автоматически определяет цель) | все |

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
  prefix: "!"         # префикс по умолчанию для команд через сообщения

database:
  path: "barman.db"   # путь к SQLite-файлу

notifications:
  webhook_url: "" # Discord webhook URL для уведомлений бота (оставить пустым для отключения)
```

`configs/config.yaml` добавлен в `.gitignore` и не попадает в репозиторий.

### Webhook-уведомления

Укажите `notifications.webhook_url` — Discord webhook URL — чтобы получать уведомления бота в канале.
Ошибки и действия пользователей отправляются в один и тот же вебхук:

| Событие | Цвет embed |
|---|---|
| `Error` / `Fatal` / `Panic` | 🔴 Красный |
| Слэш-команды, префиксные реакции, входы участников, смена префикса/авто-роли | 🟢 Зелёный |

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
│   ├── guild/             # SetAutoRole, GetAutoRole, RemoveAutoRole, SetPrefix, GetPrefix, RemovePrefix
│   ├── member/            # AssignAutoRole, интерфейс RoleAssigner
│   └── reaction/          # FetchGIFUseCase, FetchGIFWithFallbackUseCase, интерфейсы GIFFetcher/GIFExecutor
├── adapter/
│   ├── command/           # slash-команды (discordgo)
│   ├── handler/           # обработчики GuildMemberAdd, MessageCreate, интерактивных компонентов
│   └── repository/sqlite/ # реализация Repository через SQLite
└── infrastructure/
    ├── config/            # загрузка YAML-конфига
    ├── database/          # открытие SQLite
    ├── discord/           # discordgo session, RoleAssigner
    ├── nekos/             # HTTP клиент nekos.best (основной источник GIF)
    └── otakugifs/         # HTTP клиент otakugifs.xyz (fallback источник GIF)
```

Моки генерируются через [mockery](https://github.com/vektra/mockery) (`make mock`) и закоммичены в репозиторий.

### Миграции базы данных

Миграции находятся в `migrations/` и применяются вручную на сервере:

```bash
sqlite3 barman.db < migrations/000001_init_guild_settings.up.sql
```

## CI/CD

GitHub Actions pipeline при каждом push:

```
build (go build → artifact)
  ├── lint
  ├── test
  └── dependency_check
        └── docker (сборка и публикация образа в GHCR)
                └── deploy  (только main → VPS по SSH)
```

- **build** — компилирует бинарь с `CGO_ENABLED=0`, сохраняет как artifact
- **lint** — `golangci-lint`
- **test** — `go test ./...`
- **dependency_check** — `govulncheck`
- **docker** — скачивает artifact, собирает минимальный Docker-образ и пушит в GHCR с тегом `{sha7}-{YYYYMMDD}` (main) или `{branch}-{sha7}` (другие ветки)
- **deploy** — `docker compose pull && down && up -d` на VPS

### Необходимые секреты репозитория

| Секрет | Описание |
|--------|----------|
| `GITHUB_TOKEN` | Встроенный, создавать не нужно |
| `VPS_HOST` | IP или домен VPS |
| `VPS_USER` | SSH-пользователь |
| `VPS_PASSWORD` | SSH-пароль |
| `BOT_TOKEN` | Discord bot token |
| `BOT_APP_ID` | Discord application ID |

## Стек

- [discordgo](https://github.com/bwmarrin/discordgo) — Discord API
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) — pure Go SQLite (без CGO)
- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) — YAML конфиг
- [testify](https://github.com/stretchr/testify) + [mockery](https://github.com/vektra/mockery) — тесты и моки
- [golangci-lint](https://github.com/golangci/golangci-lint) — линтер
- [nekos.best](https://nekos.best) — основной API аниме-реакций в формате GIF
- [otakugifs.xyz](https://otakugifs.xyz) — fallback API аниме-реакций в формате GIF