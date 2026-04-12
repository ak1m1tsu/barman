CREATE TABLE IF NOT EXISTS guild_settings (
    guild_id     TEXT PRIMARY KEY,
    auto_role_id TEXT NOT NULL DEFAULT '',
    prefix       TEXT NOT NULL DEFAULT ''
);
