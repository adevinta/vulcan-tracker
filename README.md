# vulcan-tracker

Service to register tickets associated to vulnerabilities in a tracker tool.

## ⚠️ Alpha status

This service is under active development and for sure will break compatibility until it gets a stable release.


## Running the service in local mode

For running the component locally, clone and run at the root of the repo the following:

```bash
go install ./...
cd db && source postgres-start.sh && cd -
cd db && source flyway-migrate.sh && cd -
vulcan-tracker -c _resources/config/local.toml
```

To stop the dependencies, run:
```bash
docker-compose down --remove-orphans
```

## Docker execute

Those are the variables you have to use:

|Variable|Description|Sample|
|---|---|---|
|PORT||8080|
|LOG_LEVEL||error|
|PG_HOST|Database host|localhost|
|PG_NAME|Database name|vulnerabilitydb|
|PG_USER|Database user|vulnerabilitydb|
|PG_PASSWORD|Database password|vulnerabilitydb|
|PG_PORT|Database port|5432|
|PG_SSLMODE|One of these (disable,allow,prefer,require,verify-ca,verify-full)|disable|


```bash
docker build . -t vulcantracker

# Use the default config.toml customized with env variables.
docker run --env-file ./local.env vulcantracker

# Use custom config.toml
docker run -v `pwd`/custom.toml:/app/config.toml vdba
```
