CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE finding_tickets (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     team_id UUID NOT NULL,
     finding_id UUID NOT NULL,
     url_tracker TEXT NOT NULL,

     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
     updated_at TIMESTAMP WITH TIME ZONE,
     UNIQUE (team_id, finding_id)
);
