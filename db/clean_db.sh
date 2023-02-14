#!/usr/bin/env bash

# Copyright 2020 Adevinta

source postgres-stop.sh
source postgres-start.sh
sleep 2
source flyway-migrate.sh
