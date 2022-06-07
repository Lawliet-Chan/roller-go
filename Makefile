.PHONY: lint docker clean roller

IMAGE_NAME=roller
IMAGE_VERSION=latest

roller: ## Builds the Roller instance.
	GOBIN=$(PWD)/build/bin go install ./cmd/roller

lint: ## Lint the files - used for CI
	GOBIN=$(PWD)/build/bin go run build/lint.go

clean: ## Empty out the bin folder
	@rm -rf build/bin

docker:
	docker build -t scrolltech/${IMAGE_NAME}:${IMAGE_VERSION} ./