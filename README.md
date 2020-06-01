# bake

Bake contains build tools to help make us elevate the developer experience of our Go projects.

The repository contains two major components:

- Mage targets and helpers, to support a more structured and common make experience.
- Bake Dockerfile, which is used to create an image for our CI/CD process

## Mage

[mage](https://magefile.org/) is an alternative to make. Targets are written in Go and can be found in `magefile.go` in the root directory of the repository.

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

As an opt-in speed improvement we can pre-generate project specific mage binary named `bake-build`.
To build one first install mage and then run `mage -goos linux -goarch amd64 -compile bake-build`.
Note that it's your responsibility to keep this local binary up to date if any mage related code changes.

The image is build and uploaded to the repository's package storage on every new release.