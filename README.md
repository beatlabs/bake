<!-- Space: DT -->
<!-- Title: Bake -->
<!-- Parent: Engineering -->
<!-- Parent: Dev Tools -->

# Bake

[![Coverage Status](https://coveralls.io/repos/github/taxibeat/bake/badge.svg?branch=master&t=yYHNCW)](https://coveralls.io/github/taxibeat/bake?branch=master)

Bake contains build tools to help make us elevate the developer experience of our Go projects.

The repository provides two major components:

- Mage targets and helpers, to support a more structured and common make experience.
- A Bake Docker image containing pinned versions of Go and several CI tools (linters/code generators/etc), to guarantee a consistent experience.

## Working with Mage

Bake uses [mage](https://magefile.org/) which is an alternative to make.

Mage targets are written in Go and can be found in `magefile.go` in the root directory of the repository.

This repository provides several pre-built mage targets for tests, linting, documentation generation, and others.

You can see the `magefile.go` in this repo for reference.

## Executing mage targets locally

In order to run mage targets in your local environment, you can install mage as well as any dependencies you may need (e.g. `golangci-lint`).

For a complete list of available targets run `mage`.

Some examples:

Run unit tests:

```bash
mage test:unit
```

And all tests (unit+integration+component):

```bash
mage test:all
```

Take a look at the `magefile.go` of this project to see how it works.

## Using the Bake Docker image

In order to run mage targets in a controlled environment with external dependencies we can use the Bake image.

This is a clean slate approach (requiring nothing but Docker to be installed) which behaves the same both locally and in Jenkins thus providing strong reproducibility guarantees. The trade-off is that it's slower since we must spin up a container to run the mage targets, and can't make use of test caches or mage caches for example.

In order to use the Bake image please follow these instructions:

#### 1. Generate a github personal access token

This is required in order to access private repos (including the go packages in the bake repo).

A token can be generated at [https://github.com/settings/tokens](https://github.com/settings/tokens) and must have 'repo' scope and be SSO enabled

You can export it in your shell 

```bash
export GITHUB_TOKEN=my-token
```

or set it inline before executing bake:

```bash
GITHUB_TOKEN=my-token ./bake.sh
```

#### 2. Create a `bake.sh` script

In order to generate an initial `bake.sh` you can run copy the one from this repo or run:

```bash
$ docker run --rm -it -e GITHUB_TOKEN=$GITHUB_TOKEN taxibeat/bake:<version> --gen-script > bake.sh
```

And modify to your needs, e.g. add any env vars that your targets may require.

#### 3. Optional - Speed up bake

One of the most time consuming steps when running a mage target via the Bake image is waiting for Mage to compile it's ad-hoc binary.
This can be circumvented by manually creating that binary, which saves a few seconds. The downside is that if the `magefile.go` changes then this manually created binary will be out of date, and must be manually updated/deleted.

```bash
$ docker run --rm -it -v $PWD:/src -w /src -e GITHUB_TOKEN=$GITHUB_TOKEN -u $(id -u):$(id -g) taxibeat/bake:<version> --gen-bin
```

And add `bake-build` to your `.gitignore`.

#### 4. Executing targets

Instead of executing `mage` we now execute the script, e.g:

```bash
./bake.sh ci:run
```

This is the recommended way to run the CI target in Jenkins.

## Tools included in the Bake image

The bake image is used for local and CI/CD and contains all the tools we need in our local and CI/CD environments:

- [mage](https://magefile.org/) a make file replacement with Go code
- [hadolint](https://github.com/hadolint/hadolint) docker file linting
- [swaggo](https://github.com/swaggo/swag) for generating swagger files from annotations
- [mark](https://github.com/mantzas/mark) for syncing markdown documentation to Confluence
- [golangci-lint](https://github.com/golangci/golangci-lint) a multi-linter for Go
- [helm](https://helm.sh/) a k8s package manager
- [goveralls](https://github.com/mattn/goveralls) a Coveralls cli tool

The image is built and uploaded to the repository's package storage on every new release.

The version of the Bake image and of the Bake Go module are kept in sync, and should be updated together in projects that use Bake.

## Example projects

- https://github.com/taxibeat/direction
