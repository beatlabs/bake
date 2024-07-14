default: test

test: fmtcheck
	go test ./... -cover -race -timeout 60s

testint: fmtcheck
	go test ./... -race -cover -tags=integration -timeout 300s

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

build-docker:
	docker build -f Dockerfile -t ghcr.io/beatlabs/bake:latest . --build-arg GH_TOKEN=${GITHUB_TOKEN}

# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

.PHONY: default test testint fmtcheck build-docker
