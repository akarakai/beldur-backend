-- Clean DB (drop in dependency order)
DROP TABLE IF EXISTS campaigns_players;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS accounts;

-- Accounts
CREATE TABLE accounts (
    account_id   SERIAL PRIMARY KEY,
    username     VARCHAR(20) UNIQUE NOT NULL,
    password     VARCHAR(255) NOT NULL,
    email        VARCHAR(255) UNIQUE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    last_access  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Players (1:1 with accounts)
CREATE TABLE players (
    player_id    SERIAL PRIMARY KEY,
    name  VARCHAR(20) UNIQUE NOT NULL,
    account_id   INTEGER NOT NULL UNIQUE,

    CONSTRAINT fk_players_account
        FOREIGN KEY (account_id)
        REFERENCES accounts(account_id)
        ON DELETE CASCADE
);

-- Campaigns
CREATE TABLE campaigns (
    campaign_id   SERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    description   VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at    TIMESTAMP,
    finished_at   TIMESTAMP,
    status        VARCHAR(30) NOT NULL,
    master_id     INTEGER NOT NULL,

    CONSTRAINT fk_campaigns_player_master
        FOREIGN KEY (master_id)
        REFERENCES players(player_id)
);

CREATE TABLE campaigns_players (
    campaign_id  INTEGER NOT NULL,
    player_id    INTEGER NOT NULL,
    joined_at    TIMESTAMP DEFAULT NOW(),
    is_master    BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT pk_campaigns_players
        PRIMARY KEY (campaign_id, player_id),

    CONSTRAINT fk_campaigns_players_campaign
        FOREIGN KEY (campaign_id)
        REFERENCES campaigns(campaign_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_campaigns_players_player
        FOREIGN KEY (player_id)
        REFERENCES players(player_id)
        ON DELETE CASCADE
);