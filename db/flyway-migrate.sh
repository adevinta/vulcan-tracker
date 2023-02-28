#!/usr/bin/env bash

# Copyright 2023 Adevinta

docker run --net=host --rm -v "$PWD"/sql:/flyway/sql flyway/flyway:"${FLYWAY_VERSION:-9.15.1}-alpine" -community \
    -user=vultrackerdb -password=vultrackerdb -url=jdbc:postgresql://localhost:5439/vultrackerdb -baselineOnMigrate=true migrate
