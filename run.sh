#!/bin/sh

# Copyright 2022 Adevinta

# Apply env variables
envsubst < config.toml > run.toml

exec ./vulcan-tracker -c run.toml