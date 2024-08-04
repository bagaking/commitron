BUILD_DIR := .build

.PHONY: run test build check bundle cz

run:
	source ./set_env.sh && go run .

test:
	go test ./...

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/ .

check: test build

cz:
	source ./set_env.sh && git cz

bundle:
	file_bundle -v
