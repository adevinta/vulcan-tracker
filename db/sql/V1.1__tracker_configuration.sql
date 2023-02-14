CREATE TABLE ticket_tracker_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    kind TEXT NOT NULL
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    team_id UUID NOT NULL,
    ticket_tracker_servers_id UUID REFERENCES ticket_tracker_servers(id),
    project TEXT NOT NULL,
    issue_type TEXT NOT NULL,
    fix_workflow TEXT[],
    wont_fix_workflow TEXT[],
    auto_create BOOL NOT NULL
)