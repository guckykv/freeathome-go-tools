ifeq ($(OS),Windows_NT)
	TARGET_INFLUX = fahinflux.exe
    TARGET_CLI = fahcli.exe
	TARGET_SWITCH = fahvswitch.exe
else
	TARGET_INFLUX = fahinflux
    TARGET_CLI = fahcli
	TARGET_SWITCH = fahvswitch
endif

all: $(TARGET_INFLUX) $(TARGET_CLI) $(TARGET_SWITCH)
all-pi: fahinflux-pi fahcli-pi fahvswitch-pi

$(TARGET_INFLUX):
	cd cmd/fahinflux && go build -o $(TARGET_INFLUX) main.go Influx.go Influx4Unit.go

fahinflux-pi:
	cd cmd/fahinflux && env GOOS=linux GOARCH=arm GOARM=7 go build -o fahinflux-pi main.go Influx.go Influx4Unit.go

$(TARGET_CLI):
	cd cmd/fahcli && go build -o $(TARGET_CLI) main.go channel.go device.go getset.go virtual.go

fahcli-pi:
	cd cmd/fahcli && env GOOS=linux GOARCH=arm GOARM=7 go build -o fahcli-pi main.go channel.go device.go getset.go virtual.go

$(TARGET_SWITCH):
	cd cmd/fahvswitch && go build -o $(TARGET_SWITCH) main.go

fahvswitch-pi:
	cd cmd/fahvswitch && GOOS=linux GOARCH=arm GOARM=7 go build -o fahvswitch-pi main.go
