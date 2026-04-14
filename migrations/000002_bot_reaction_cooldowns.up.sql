CREATE TABLE IF NOT EXISTS bot_reaction_cooldowns (
    user_id    TEXT PRIMARY KEY,
    used_at    DATETIME NOT NULL
);
