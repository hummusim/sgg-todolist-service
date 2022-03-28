BUF_VERSION:=1.1.0
MIGRATION_DIR=scripts/db/migrations

generate:
	docker run -v $$(pwd):/src -w /src --rm bufbuild/buf:$(BUF_VERSION) generate