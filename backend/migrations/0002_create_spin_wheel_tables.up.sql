CREATE TABLE tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('spin_wheel','auction')),
    status VARCHAR(30) NOT NULL DEFAULT 'registration_open'
        CHECK (status IN ('registration_open','registration_closed','in_progress','completed')),
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    registration_link_token UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE spin_wheel_settings (
    tournament_id UUID PRIMARY KEY REFERENCES tournaments(id) ON DELETE CASCADE,
    bracket_size INT NOT NULL CHECK (bracket_size >= 2 AND bracket_size <= 64 AND bracket_size % 2 = 0)
);

CREATE TABLE spin_wheel_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    player_name VARCHAR(100) NOT NULL,
    phone VARCHAR(15),
    source VARCHAR(10) NOT NULL CHECK (source IN ('form','manual')),
    added_by UUID REFERENCES users(id),
    color_hex VARCHAR(7),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tournament_id, color_hex)
);

CREATE TABLE spin_wheel_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    round INT NOT NULL,
    match_number INT NOT NULL,
    player1_entry_id UUID REFERENCES spin_wheel_entries(id),
    player2_entry_id UUID REFERENCES spin_wheel_entries(id),
    player1_score INT,
    player2_score INT,
    winner_entry_id UUID REFERENCES spin_wheel_entries(id),
    status VARCHAR(10) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','completed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tournament_id, round, match_number)
);
