all: build-proxy

build-proxy:
	@mkdir -p bin
	go build -o bin/rcproxy ./main

clean:
	@rm -rf bin

gotest:
	go test ./proxy/... -race