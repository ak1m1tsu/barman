CREATE TABLE IF NOT EXISTS reaction_stats (
    reaction_type TEXT PRIMARY KEY,
    count         INTEGER NOT NULL DEFAULT 0
);
