API Cache
=========

> Taken from: SEDMAX

Description
-----------

Write server application that has an HTTP API for working with string key-value storage in RAM.

Each key must have its own lifetime, which is extended when the key value changes.
If the time has expired, then the key with the value is deleted from the storage 
regardless of user requests.

Need implement the WEB API that allows to work with the key-value storage:

 - set the value for the given key (if the key already exists, then it is updated, if not, it is added)
 - return the value for the given key (if there is no value, an error is returned)
 - the delete function, if there is no value for the given key, should return an error

Constraints
-----------

There can be any number of network connections, but a limited number 
of goroutines accessing the storage inside the application.

Solution
--------

Taste it!

#### Dependencies

1. `redis (any)`
2. `memcached (any)`

#### Testing

```bash
# Note: memcached and redis servers will up with `$ docker run` command
$ make
```

#### Running

```bash
# Note: redis server will up with `$ docker run` command
$ make dev
```

#### Example

```bash
curl -X GET http://127.0.0.1:8080/1
# {"error":"key (1) not exist"}

curl -X POST -d '{"key":"1","val":"2","ttl":2}' http://127.0.0.1:8080
# no body
curl -X POST -d '{"key":"2","val":"1","ttl":10}' http://127.0.0.1:8080
# no body

curl -X GET http://127.0.0.1:8080/1
# {"value":"2"}

curl -X GET http://127.0.0.1:8080/2
# {"value":"1"}

sleep 2

curl -X DELETE http://127.0.0.1:8080/1
# {"error":"key (1) not exist"}

curl -X DELETE http://127.0.0.1:8080/2 
# no body
```

#### How it works

There are the main entities used:

1. `Storage` - real key-value storage that implements `internal/fs/Driver` interface.
2. `FileSystem` - implements `internal/fs/Driver` interface and aggregates the needed constraint:
    ```
        There can be any number of network connections, but a limited number 
        of goroutines accessing the storage inside the application.
    ```
3. `Server` - APICache HTTP server to interrupt with any entity that implements `internal/fs/Driver` interface.

The solution was inspired by different physical `Storage` and how Operation System 
works with their drivers through `FileSystem` with own logical storage driver.

So, `Server` doesn't know about inner `Storage` and how it works. 
`internal/fs` package provides for `Server` some API - `internal/fs/Driver`. 
We can use Specific `Driver` directly from `internal/drivers`, but `internal/fs` package
provides the same interface and adds to Specific `Driver` validation and concurrent, 
like Operation System high-level `FileSystem` manipulates with different hardware.
