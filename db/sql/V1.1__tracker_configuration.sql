CREATE TABLE tracker_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL
);

CREATE TABLE projects
(
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                      TEXT NOT NULL,
    team_id                   UUID NOT NULL,
    tracker_configuration_id  UUID REFERENCES tracker_configurations (id),
    project                   TEXT NOT NULL,
    issue_type                TEXT NOT NULL
)
