all: build-proxy

build-proxy:
	@mkdir -p bin
	go build -o bin/rcproxy ./main

clean:
	@rm -rf bin

test:
	go test -v ./proxy/... -race
