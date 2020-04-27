help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ".:*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

run:
run: ## Run dev mode
	go run cmd/main.go -servername=bastion

run-error:
run-error: ## Run error mode
	go run cmd/main.go -servername=bastionA

run-help:
run-help: ## show app help
	go run cmd/main.go

release:
release: ## Build new release version
	AUTHOR=$(Author) goreleaser --skip-validate --skip-publish

clean:
clean: ## clean dist
	rm -rf dist
