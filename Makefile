VERSION?=$(shell git tag | grep ^v | sort -V | tail -n 1)
GOFLAGS?=-ldflags '-X main.VERSION=${VERSION}'

dungeonbot: dungeonbot.go go.mod go.sum
	@echo
	@echo Building dungeonbot. This may take a minute or two.
	go build $(GOFLAGS) -o $@
	@echo
	@echo ...Done\!

.PHONY: clean
clean:
	@echo
	@echo Cleaning build and module caches...
	go clean
	@echo
	@echo ...Done\!

.PHONY: update
update:
	@echo
	@echo Updating from upstream repository...
	@echo
	git pull --rebase origin master
	@echo
	@echo ...Done\!
