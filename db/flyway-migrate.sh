#!/usr/bin/env bash

# Copyright 2023 Adevinta

docker exec vultrackerdb psql -c "DROP DATABASE IF EXISTS vultrackerdb_test;" -U vultrackerdb
docker exec vultrackerdb psql -c "DROP USER IF EXISTS vultrackerdb_test;" -U vultrackerdb
docker exec vultrackerdb psql -c "CREATE USER vultrackerdb_test WITH PASSWORD 'vultrackerdb_test';" -U vultrackerdb
docker exec vultrackerdb psql -c "ALTER USER vultrackerdb_test WITH SUPERUSER;" -U vultrackerdb
docker exec vultrackerdb psql -c "CREATE DATABASE vultrackerdb_test;" -U vultrackerdb


docker run --net=host --rm -v "$PWD"/sql:/flyway/sql flyway/flyway:"${FLYWAY_VERSION:-9.15.1}-alpine" -community \
    -user=vultrackerdb -password=vultrackerdb -url=jdbc:postgresql://localhost:5439/vultrackerdb -baselineOnMigrate=true migrate
