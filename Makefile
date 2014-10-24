all: gen

test: gen
	go test ./...

fmt: gen
	go fmt ./...

gen:
	(cd core && make gen)
