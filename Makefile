
proj = binserver
binary = $(proj)

default: build
build: main.go
	@go-bindata -prefix dist/ dist/...
	@go build

deps:
	@glide up -v

clean:
	@rm $(binary)

.PHONY: build
