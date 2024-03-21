#!/usr/bin/env bash

# Copyright 2023 Adevinta

docker run --net=host flyway/flyway:"${FLYWAY_VERSION:-10}-alpine" \
    -user=vultrackerdb -password=vultrackerdb -url=jdbc:postgresql://localhost:5439/vultrackerdb -baselineOnMigrate=true -cleanDisabled=false clean
