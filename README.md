# vulcan-tracker

Service to register tickets associated to vulnerabilities in a tracker tool.

## ⚠️ Alpha status

This service is under active development and for sure will break compatibility until it gets a stable release.



## Running the service in local mode

Firt of all, you will need to build the service locally and run it from your local folder. To do this, run the following command from the root folder of the project:

```bash
go build ./cmd/vulcan-tracker
```


```sh
# run the service. This script uses config.toml as configuration file.
./run.sh
```

If you want to use a toml file as storage for servers and teams configuration, you will find an example in `./_resources/config/local.toml`.



## Installing

For running the component locally, clone and run at the root of the repo the following:

```sh
go install ./...
```

## Test

Execute the tests:

```
_script/test -cover ./...
```

`_script/test` makes sure the testing infrastructure is up and running and then
runs `go test` with the provided arguments. It also disables test caching and
avoids running multiple test programs in parallel.

Stop the testing infrastructure:

```
_script/clean
```