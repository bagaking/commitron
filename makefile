.PHONY: run bundle cz

run:
	source ./set_env.sh && go run .

cz:
	source ./set_env.sh && git cz

bundle:
	file_bundle -v