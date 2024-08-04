.PHONY: run test bundle cz

run:
	source ./set_env.sh && go run .

test:
	go test ./...

cz:
	source ./set_env.sh && git cz

bundle:
	file_bundle -v
