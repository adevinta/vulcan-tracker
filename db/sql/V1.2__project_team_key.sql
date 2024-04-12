ALTER TABLE projects
ADD CONSTRAINT projects_team_id_key UNIQUE (team_id);
