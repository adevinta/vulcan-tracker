#!/bin/sh

# Copyright 2022 Adevinta

# Apply env variables
envsubst < config.toml > run.toml

exec ./vulcan-jira-api -c run.toml