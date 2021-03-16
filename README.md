<!-- Space: DT -->
<!-- Title: Bake -->
<!-- Parent: Engineering -->
<!-- Parent: Dev Tools -->

# Bake

[![Coverage Status](https://coveralls.io/repos/github/taxibeat/bake/badge.svg?branch=master&t=yYHNCW)](https://coveralls.io/github/taxibeat/bake?branch=master)

Bake contains build tools to help make us elevate the developer experience of our Go projects.

The repository provides two major components:

- Mage targets and helpers, to support a more structured and common make-like experience.
- A Bake Docker image containing pinned versions of Go and several CI tools (linters/code generators/etc), to guarantee a consistent experience.

## Working with Mage targets

Bake uses [Mage](https://magefile.org/) which is an alternative to make.

Mage targets are written in Go and can be found in `magefile.go` in the root directory of the repository.

This repository provides several pre-built Mage targets for tests, linting, documentation generation, and others.

You can see the `magefile.go` in this repo for reference.

## Executing targets directly

In order to run Mage targets in your local environment, you should install `mage` as well as any dependencies the targets may need (e.g. `go`, `golangci-lint`).

For a complete list of available targets run `mage`.

Some examples:

Run unit tests (using local Go installation, test caching).

```shell
mage test:unit
```

And all tests (unit+integration+component).

```shell
mage test:all
```

Clear Docker resources used for integration/component tests.

```shell
mage test:cleanup
```

Take a look at the `magefile.go` of this project to see how it works.

## Executing targets with the Bake Docker image

In order to run Mage targets in a controlled environment with no external dependencies we can use the Bake image.

This is a fully isolated approach (requiring nothing but Docker to be installed) that behaves the same on any linux/mac execution environment. This has the aim of providing parity between CI and local environments.

The trade-off is that it's slower since we must spin up a Bake Docker container to execute the Mage targets and we don't make use of test caches or Mage caches.

### 1. Generate a github personal access token

This is required in order to access private repos (including the go packages in the bake repo).

A token can be generated at [here](https://github.com/settings/tokens) and must have 'repo' scope and be SSO enabled.

You can export it in your shell 

```bash
export GITHUB_TOKEN=my-token
```

or set it inline before executing bake:

```bash
GITHUB_TOKEN=my-token ./bake.sh
```

### 2. Create a `bake.sh` script

In order to generate an initial `bake.sh` you can run copy the one from this repo or from a project where bake has already been setup and modify to your needs, e.g. add any env vars that your targets may require.

### 3. Executing targets

Instead of executing `mage` we now execute the script, e.g:

```bash
./bake.sh ci:run
```

This is the recommended way to run the CI target in Jenkins/Github Actions.

Note: You can set a `SKIP_CLEANUP=1` env var to keep Docker resources available after finishing the run, which can be helpful to debug failed runs.

## Tools included in the Bake image

The bake Docker image contains the tools required to execute all Mage targets.

- [mage](https://magefile.org/) a make file replacement with Go code
- [hadolint](https://github.com/hadolint/hadolint) docker file linting
- [swaggo](https://github.com/swaggo/swag) for generating swagger files from annotations
- [mark](https://github.com/mantzas/mark) for syncing markdown documentation to Confluence
- [golangci-lint](https://github.com/golangci/golangci-lint) a multi-linter for Go
- [helm](https://helm.sh/) a k8s package manager
- [goveralls](https://github.com/mattn/goveralls) a Coveralls cli tool

The image is built and uploaded to the repository's package storage on every new release.

The version of the Bake image and of the Bake Go module are kept in sync, and should be updated together in projects that use Bake.

## Repos using Bake

- https://github.com/taxibeat/direction
- https://github.com/taxibeat/route
- https://github.com/taxibeat/sonar
- https://github.com/taxibeat/dispatch
- https://github.com/taxibeat/eta

## Recipes

For a list of useful recipes please see [doc/recipes.md](doc/recipes.md).
