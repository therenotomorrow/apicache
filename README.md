API Cache
=========

> Test task taken somewhere from the internet.

Goal
----

Write server application that has an HTTP API for working with string key-value storage in RAM.

Each key must have its own lifetime, which is extended when the key value changes.
If the time has expired, then the key with the value is deleted from the storage 
regardless of user requests.

Need implement the WEB API that allows to work with the key-value storage:

 - set the value for the given key (if the key already exists, then it is updated, if not, it is added)
 - return the value for the given key (if there is no value, an error is returned)
 - delete the value for the given key (if there is no value, do nothing)

**Constraints**: there can be any number of network connections, but a limited number 
of goroutines accessing the storage inside the application. That's why was selected another approach than 
simple rate limiter because rate limiter must limit network connections, not internal.

System Requirements
-------------------

```shell
go version
# go version go1.22.4 ...

redis-server --version
# Redis server v=7.2.5 ...

memcached --version
# memcached 1.6.29 ...
```

Environment
-----------

**Note**: use prefix `APICACHE_` for enable variable to be caught in runtime.

| Variable              | Type                                | Description                                           |
|:----------------------|:------------------------------------|:------------------------------------------------------|
| `DEBUG`               | `bool`                              | Enable debug mode or not                              |
| `DRIVER_NAME`         | `["machine", "memcached", "redis"]` | Driver type (supported)                               |
| `DRIVER_ADDRESS`      | `string`                            | Driver DSN address                                    |
| `DRIVER_MAX_CONN`     | `int`                               | Maximum number of simultaneous connections to the API |
| `DRIVER_CONN_TIMEOUT` | `time.Duration`                     | Connection timeout for application                    |

Development
-----------

Download sources

```shell
export PROJECT_ROOT=apicache
git clone https://github.com/therenotomorrow/apicache.git ${PROJECT_ROOT}
cd ${PROJECT_ROOT}
```

Setup project requirements

```shell
# setup requirements
go mod download
go mod verify

# lint GO code
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.1

# create documentation on API
go install github.com/swaggo/swag/cmd/swag@v1.16.3
```

Setup project environment

```shell
cp ./configs/.env.example .env
vim .env
```

Taste it :heart:

```shell
# check code integrity
make docs code test/smoke

# run application
go run cmd/app/main.go

# check everything works
open 'http://127.0.0.1:8080/api/docs/'
```

Setup safe development

```shell
./scripts/pre-commit.sh
```

Testing
-------

Controls by [test.sh](./scripts/test.sh) or [Makefile](./Makefile) and contains:

```shell
# fast unit tests to be sure that no regression was 
make test/smoke

# same as test/smoke but with -race condition check
make test/unit

# integration (driver, etc.) tests that needed external resources, also with -race condition
# ATTENTION: before run make sure that redis and memcached are available
make driver/redis &
make driver/memcached &
make test/integration

# combines both (test/unit and test/integration) to create local coverage report in HTML
make test/coverage
```
