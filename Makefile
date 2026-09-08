TOOLS := fahinflux fahcli fahvswitch

.PHONY: all all-pi test vet fmt clean $(TOOLS) $(addsuffix -pi,$(TOOLS))

all: $(TOOLS)
all-pi: $(addsuffix -pi,$(TOOLS))

$(TOOLS):
	go build -o cmd/$@/$@ ./cmd/$@

$(addsuffix -pi,$(TOOLS)):
	env GOOS=linux GOARCH=arm GOARM=7 go build -o cmd/$(patsubst %-pi,%,$@)/$@ ./cmd/$(patsubst %-pi,%,$@)

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

clean:
	rm -f $(foreach t,$(TOOLS),cmd/$(t)/$(t) cmd/$(t)/$(t)-pi)
