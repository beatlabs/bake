<!-- Space: DT -->
<!-- Title: Bake -->
<!-- Parent: Engineering -->
<!-- Parent: Dev Tools -->

# Bake

[![Coverage Status](https://coveralls.io/repos/github/beatlabs/bake/badge.svg?branch=master&t=yYHNCW)](https://coveralls.io/github/beatlabs/bake?branch=master)

Bake is a set of tools that aims to improve the developer experience for our Go projects.

The repository provides 3 things:

- [Mage](https://magefile.org/) powered make-like targets in the `targets` pkg.
- [DockerTest](https://github.com/ory/dockertest) powered state management for our component tests under the `docker` pkg.
- A [Docker image](https://github.com/beatlabs/bake/pkgs/container/bake) to ensure parity between CI and local environments.

## Setup Targets

Bake provides several Mage targets for tests, linting, documentation generation, etc.

In your project you can import these and use them to build your own CI target alongside any custom targets that you may have.

There is a simple example in `magefile.go` in this repo.

## Executing Targets

For a complete list of available targets run `mage`.

```console
$ mage
Targets:
  ci                    runs the Continuous Integration pipeline.
  diagram:generate      creates diagrams from python files.
  doc:confluenceSync    synchronized annotated docs to confluence.
  go:checkVendor        checks if vendor is in sync with go.mod.
  go:fmt                runs go fmt.
  go:fmtCheck           checks if all files are formatted.
  go:modSync            runs go module tidy and vendor.
  lint:docker           lints the docker file.
  lint:go               runs the golangci-lint linter.
  lint:goShowConfig     outputs the golangci-lint linter config.
  test:all              runs all tests.
  test:cleanup          removes any local resources created by `mage test:all`.
  test:component        runs unit and component tests.
  test:coverAll         runs all tests and produces a coverage report.
  test:coverUnit        runs unit tests and produces a coverage report.
  test:integration      runs unit and integration tests.
  test:unit             runs unit tests.

```

Go linting (using local golanci-lint if available):

```console
mage lint:go
```

Go Unit tests (using local Go installation and test cache):

```console
mage test:unit
```

### Component tests

Go unit+integration+component tests:

```console
mage test:all
```

_This will create docker containers according to your component test setup (usually in `TestMain` under `/tests`)._

Tear down Docker resources used for integration/component tests:

```console
mage test:cleanup
```

## Docker based isolated environment

This is a fully isolated approach to executing targets that provides parity between CI and local environments.

The trade-off is that it's slower since we must spin up a Docker container to execute the Mage targets so we don't make use of test caches or Mage caches.

The version of the Bake image and of the Bake Go module are kept in sync, and should be updated together in projects that use Bake.

Unlike the local version, containers used for component tests are torn down automatically after every run.

### 1. Generate a github personal access token

This is required in order to access private repos (including the go packages in the bake repo).

A token can be generated at [here](https://github.com/settings/tokens) and must have `repo` `read:packages` scope and be **SSO enabled**.

Export it in your shell

```console
export GITHUB_TOKEN=my-token
```

Login to `ghcr.io`

```console
$ echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR-USERNAME --password-stdin
> Login Succeeded
```

### 2. Import scripts package

Add this import in the `magefile.go` so `go mod vendor` will fetch the bake runner script.

```go
// generic bake script
import _ "github.com/beatlabs/bake/scripts"
```

### 3. Create a `bake.sh` script for your repo

It is a simple script that runs the bake runner script and can be copied from `go-matching-template`
or created from scratch:

```bash
#!/bin/bash
set -e
bash ./vendor/github.com/beatlabs/bake/scripts/run-bake.sh "$@"
```

If you need to pass any custom environment variables to Bake, you can do it
by adding one or more `--env` flags to the run-bake script.

```bash
bash ./vendor/github.com/beatlabs/bake/scripts/run-bake.sh --env SOME_ENV_VAR=some-value "$@"
```

### 4. Execute

Instead of executing `mage` we now execute the script, e.g:

```console
./bake.sh ci
```

This is the recommended way to run the CI target in Jenkins/Github Actions.

## Tools used by targets

- [hadolint](https://github.com/hadolint/hadolint) docker file linting
- [swaggo](https://github.com/swaggo/swag) for generating swagger files from annotations
- [mark](https://github.com/kovetskiy/mark) for syncing markdown documentation to Confluence
- [golangci-lint](https://github.com/golangci/golangci-lint) a multi-linter for Go
- [helm](https://helm.sh/) a k8s package manager
- [goveralls](https://github.com/mattn/goveralls) a Coveralls cli tool

## Repos using Bake

- [Github search](https://github.com/search?q=org%3Abeatlabs+filename%3A%2Fbake.sh&type=Code)
