TOOLS := fahinflux fahcli fahvswitch sysapprobe

# The library is developed alongside these tools, joined by a go.work in the
# parent directory. check-noworkspace proves the build also works without it.
FAHAPI     := github.com/guckykv/freeathome-go-fahapi
FAHAPI_DIR := ../freeathome-go-fahapi

.PHONY: check build test vet fmt fmt-check check-noworkspace smoke \
        all all-pi all-pi64 clean $(TOOLS) \
        $(addsuffix -pi,$(TOOLS)) $(addsuffix -pi64,$(TOOLS))

# The one command to run before committing.
check: fmt-check vet build test

build:
	go build ./...

test:
	go test -race ./...

# go vet caches results and can report a stale pass. Clear it first.
vet:
	@go clean -cache
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt'ed:"; echo "$$unformatted"; exit 1; \
	fi

# Builds the way a consumer would: no workspace, the library as a dependency.
# go.mod carries no require on the library because go.work supplies it here, so
# this adds one temporarily, pointed at the neighbouring checkout, and puts
# go.mod and go.sum back afterwards -- including when the build fails.
check-noworkspace:
	@echo "==> building without the workspace"
	@cp go.mod go.mod.nows.bak && cp go.sum go.sum.nows.bak
	@trap 'mv go.mod.nows.bak go.mod; mv go.sum.nows.bak go.sum' EXIT; \
	go mod edit -require=$(FAHAPI)@v0.0.0 -replace=$(FAHAPI)=$(FAHAPI_DIR) && \
	GOWORK=off go mod tidy && \
	GOWORK=off go build ./... && \
	GOWORK=off go vet ./...
	@echo "==> ok, go.mod and go.sum restored"

# Reads the real System Access Point once and dumps every unit. Needs a
# reachable SysAP and ~/.fahapi-config.json. Writes nothing anywhere.
smoke:
	go run ./cmd/fahinflux -n -d

all: $(TOOLS)
all-pi: $(addsuffix -pi,$(TOOLS))
all-pi64: $(addsuffix -pi64,$(TOOLS))

$(TOOLS):
	go build -o cmd/$@/$@ ./cmd/$@

# 32-bit Raspberry Pi OS
$(addsuffix -pi,$(TOOLS)):
	env GOOS=linux GOARCH=arm GOARM=7 go build -o cmd/$(patsubst %-pi,%,$@)/$@ ./cmd/$(patsubst %-pi,%,$@)

# 64-bit Raspberry Pi OS
$(addsuffix -pi64,$(TOOLS)):
	env GOOS=linux GOARCH=arm64 go build -o cmd/$(patsubst %-pi64,%,$@)/$@ ./cmd/$(patsubst %-pi64,%,$@)

clean:
	rm -f $(foreach t,$(TOOLS),cmd/$(t)/$(t) cmd/$(t)/$(t)-pi cmd/$(t)/$(t)-pi64)
