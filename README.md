<!-- Space: DT -->
<!-- Title: Bake -->
<!-- Parent: Engineering -->
<!-- Parent: Dev Tools -->

# bake

[![Coverage Status](https://coveralls.io/repos/github/taxibeat/bake/badge.svg?branch=master&t=yYHNCW)](https://coveralls.io/github/taxibeat/bake?branch=master)

Bake contains build tools to help make us elevate the developer experience of our Go projects.

The repository contains two major components:

- Mage targets and helpers, to support a more structured and common make experience.
- Dockerfile, which is used to create an image for our CI/CD process

## Integration

In order to incorporate Bake into a project please follow the following steps:

### 1. Generate a github [personal access token](https://github.com/settings/tokens) with 'repo' scope and SSO enabled

This is required in order to access private repos (including go packages in this repo).

You can export it in your shell (`export GITHUB_TOKEN=my-token`) or set it inline before executing bake (`GITHUB_TOKEN=my-token ./bake.sh`).

### 2. Create a `bake.sh` script for running Bake

In order to generate an initial `bake.sh` you can run

```console
$ docker run --rm -it -e GITHUB_TOKEN=$GITHUB_TOKEN taxibeat/bake:<version> --gen-script > bake.sh
```

And modify to your needs, e.g. add any env vars that your targets may require.

### 3. Optional - Speed up bake by prebuilding a mage binary

```console
$ docker run --rm -it -v $PWD:/src -w /src -e GITHUB_TOKEN=$GITHUB_TOKEN -u $(id -u):$(id -g) taxibeat/bake:<version> --gen-bin
```

And add `bake-build` to your `.gitignore`.
Note that it's your responsibility to keep this local binary up to date whenever your local mage targets change.

## Executing targets

Bake uses [mage](https://magefile.org/) which is an alternative to make. Targets are written in Go and can be found in `magefile.go` in the root directory of the repository.

For a complete list of available targets run `mage` or `./bake.sh` (explained later).

Some examples:

Run unit tests:

```bash
./bake.sh test:unit
```

And all tests (unit+integration+component):

```bash
./bake.sh test:all
```

Run the CI target (advisable to run in docker to match the jenkins)

```bash
./bake.sh ci:run
```

Take a look at the `magefile.go` of this project to see how it works.

## Bake Dockerfile

The bake image is used for local and CI/CD and contains all the tools we need in our local and CI/CD environments:

- [mage](https://magefile.org/) a make file replacement with Go code
- [hadolint](https://github.com/hadolint/hadolint) docker file linting
- [swaggo](https://github.com/swaggo/swag) for generating swagger files from annotations
- [mark](https://github.com/mantzas/mark) for syncing markdown documentation to Confluence
- [golangci-lint](https://github.com/golangci/golangci-lint) a multi-linter for Go
- [helm](https://helm.sh/) a k8s package manager

Along with the Dockerfile we have also a script `bake.sh` which is responsible for starting up the above container and run our mage targets.

The image is built and uploaded to the repository's package storage on every new release.
