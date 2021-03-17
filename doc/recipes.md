<!-- Space: DT -->
<!-- Title: Bake Recipes -->
<!-- Parent: Engineering -->
<!-- Parent: Dev Tools -->
<!-- Parent: Bake -->

# Bake recipes

## Debugging component tests

These instructions use [Route service](https://github.com/taxibeat/route) as an example with [Delve](https://github.com/go-delve/delve) as the debugger.

Start by running component tests locally in one of the usual ways.

```shell
$ mage test:all
$ go test -mod=vendor -tags=component./...
$ go test -mod=vendor -tags=component -failfast -run=TestGetRoute ./... 
$ dlv test ./test --build-flags '--tags=component' -- -test.run TestGetRoute
```

This will start several Docker containers, one for the service itself and a two for dependent services consul and mockserver.

```shell
$ docker ps --format {{.Names}}
000-route-comptest-route
000-route-comptest-consul
000-route-comptest-mockserver
```

We can kill the service container:

```shell
$ docker rm -f 000-route-comptest-route
```

Inspect the published ports in `test/.bakesession`

```shell
$ cat test/.bakesession
```

```json
{
    "ID": "000-route-comptest",
    "NetworkID": "2b56c3591e47ac876180625cbcfbc87dd76edc6042cea34f7ec0968c3dd4ee14",
    "ServiceAddresses": {
        "consul": "000-route-comptest-consul:8500",
        "mockserver": "000-route-comptest-mockserver:1080",
        "route": "000-route-comptest-route:8080"
    },
    "HostMappedServiceAddresses": {
        "consul": "localhost:39439",
        "mockserver": "localhost:38047",
        "route": "localhost:43075"
    }
}
```

Construct and activate the appropriate env vars in your shell or create a `.env` file and source it.

```shell
PATRON_HTTP_DEFAULT_PORT=43075
PATRON_LOG_LEVEL=debug
CONSUL_HTTP_ADDR="localhost:39439"
ROUTE_GOOGLE_URL="http://localhost:38047"
ROUTE_GOOGLE_CLIENT_ID=random
ROUTE_GOOGLE_CLIENT_SIGNATURE=Ym9ubmVtYW1hbg==
ROUTE_OSRM_URL="http://localhost:38047"
```

Start the service.

```shell
$ dlv debug ./cmd/route/main.go
b internal/infra/http/http.go:106
c
...
```

Now in a different terminal we can execute component tests in one of the usual ways.

```shell
$ mage test:all
$ go test -mod=vendor -tags=component./...
$ go test -mod=vendor -tags=component -failfast -run=TestGetRoute ./... 
$ dlv test ./test --build-flags '--tags=component' -- -test.run TestGetRoute
```

To remove containers.

```shell
$ mage test:cleanup
```

## Speed up local Bake executions

One of the most time consuming steps when running a mage target via the Bake image is waiting for mage to compile an ad-hoc binary.

This can be circumvented by manually creating that binary, which saves a few seconds. The downside is that if the `magefile.go` changes then this manually created binary will be out of date, and must be manually updated/deleted.

```shell
$ docker run --rm -it -v $PWD:/src -w /src -e GITHUB_TOKEN=$GITHUB_TOKEN -u $(id -u):$(id -g) taxibeat/bake:<version> --gen-bin
```

And add `bake-build` to your `.gitignore`.
