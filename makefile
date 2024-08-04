.PHONY: run test build check bundle cz

run:
	source ./set_env.sh && go run .

test:
	go test ./...

build:
	go build ./...

check: test build

cz:
	source ./set_env.sh && git cz

bundle:
	file_bundle -v
