all: gen

test: gen
	go test ./...

gen:
	(cd core && make gen)
