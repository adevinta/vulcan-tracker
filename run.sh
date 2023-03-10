#!/bin/sh

# Copyright 2022 Adevinta

# Apply env variables
envsubst < config.toml > run.toml

if [ -n "$PG_CA_B64" ]; then
  mkdir /root/.postgresql
  echo "$PG_CA_B64" | base64 -d > /root/.postgresql/root.crt # for flyway
  echo "$PG_CA_B64" | base64 -d > /etc/ssl/certs/pg.crt # for vulcan-api
fi

flyway -user="$PG_USER" -password="$PG_PASSWORD" \
  -url="jdbc:postgresql://$PG_HOST:$PG_PORT/$PG_NAME?sslmode=$PG_SSLMODE" \
  -community -baselineOnMigrate=true -locations=filesystem:/app/sql migrate

exec ./vulcan-tracker -c run.toml
