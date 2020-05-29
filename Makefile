# Running "make" will show the available targets.

# commands
DOCKER := docker
MARK-IMAGE := mantzas/mark
CONFLUENCE_SECURITY=-u $(CONFLUENCE_USERNAME) -p $(CONFLUENCE_PASSWORD) -b $(CONFLUENCE_BASEURL)

# make help the default target
default: help

## confluence-docs-sync: Synchronize docs with confluence
confluence-docs-sync:
	$(DOCKER) run --rm -i -v $(CURDIR):/src/ -w /src $(MARK-IMAGE) $(CONFLUENCE_SECURITY) -f README.md
.PHONY: confluence-docs-sync

# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

## help: Show this help
help: Makefile
	@echo
	@echo "Available targets:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /' | LANG=C sort
	@echo

.PHONY: help