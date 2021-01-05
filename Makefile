all: fahinflux fahcli
all-pi: fahinflux-pi fahcli-pi

fahinflux:
	cd cmd/fahinflux && go build -o fahinflux main.go Influx.go Influx4Unit.go

fahinflux-pi:
	cd cmd/fahinflux && env GOOS=linux GOARCH=arm GOARM=7 go build -o fahinflux-pi main.go Influx.go Influx4Unit.go

fahcli:
	cd cmd/fahcli && go build -o fahcli main.go channel.go device.go getset.go virtual.go

fahcli-pi:
	cd cmd/fahcli && env GOOS=linux GOARCH=arm GOARM=7 go build -o fahcli-pi main.go channel.go device.go getset.go virtual.go

