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
cd db && ./postgres-stop.sh
```


## How to enroll teams manually

To register teams manually in vulcan-tracker, it is necessary to create the corresponding records in the `project` and `tracker_configurations` tables.

Afterwards, the corresponding secrets must be created in the AWS Secret Manager. Under the stored key `AWSSERVERCREDENTIALS_KEY`.
For every register in `tracker_configuration` we create a secret with the information below:
- Type of secret: "Other type of secret":
- Secret name: `AWSSERVERCREDENTIALS_KEY/<id_tracker_configuration>`.
- Secret value: create key/value pair with the key "token" and the Personal Access Token of the Jira account.


Using aws cli;
```shell
aws secretsmanager create-secret \
    --name /path/to/credential/key/f49b0a11-6cb6-47da-9739-21a92d84f4db \
    --description "Credentials for the account example" \
    --secret-string "{\"token\":\"7wSIKx=zV6J66E5ng4-Cqj7i-bwk-aGHumyjOkf/4LTeN6RNVT?5ZRdzBYFYNPwx\"}"
```

At this point, the access with Personal Access Token (PAT) is the only one supported.


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
|AWSSERVERCREDENTIALS_KEY|Parent key in the AWS Secret Manager to store server secrets|/vulcan/k8s/tracker/jira/|
|AWS_REGION||eu-west-1|



```bash
docker build . -t vulcantracker

# Use the default config.toml customized with env variables.
docker run --env-file ./local.env vulcantracker

# Use custom config.toml
docker run -v `pwd`/custom.toml:/app/config.toml vdba
```
