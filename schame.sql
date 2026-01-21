CREATE TABLE IF NOT EXISTS accounts (
    account_id SERIAL PRIMARY KEY,
    username VARCHAR(20) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE, 
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_access TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS players (
    player_id SERIAL PRIMARY KEY,
    player_name VARCHAR(20) UNIQUE NOT NULL,
    account_id INTEGER NOT NULL UNIQUE,
    CONSTRAINT fk_players_account
        FOREIGN KEY (account_id)
        REFERENCES accounts(account_id)
);