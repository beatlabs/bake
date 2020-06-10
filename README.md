# bake

Bake contains build tools to help make us elevate the developer experience of our Go projects.

The repository contains two major components:

- Mage targets and helpers, to support a more structured and common make experience.
- Dockerfile, which is used to create an image for our CI/CD process

## Integration

In order to incorporate Bake into a project please follow the following steps:

### 1. Generate and export a GITHUB_TOKEN with 'repo' scope

This is required if you have private repos as dependencies.

### 2. Create a `bake.sh` script for running Bake

In order to generate an initial `bake.sh` you can run

```console
$ docker run --rm -it taxibeat/bake:<version> --gen-script > bake.sh
```

And modify to your needs, e.g. add any env vars that your targets may require.

### 3. Optional - Speed up bake by prebuilding a mage binary

```console
$ docker run --rm -it taxibeat/bake:<version> --gen-bin
```

Note that it's your responsibility to keep this local binary up to date if any mage related code changes.

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

Along with the Dockerfile we have also a script `bake.sh` which is responsible for starting up the above container and run our mage targets.

The image is built and uploaded to the repository's package storage on every new release.
