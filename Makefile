export GO111MODULE=on
export CGO_ENABLED=1
export GOPRIVATE=github.com/eroshiva

POC_NAME := product-reviews
POC_VERSION := 0.1.0 # $(shell git rev-parse --abbrev-ref HEAD)
DOCKER_REPOSITORY := eroshiva
GOLANGCI_LINTERS_VERSION := v2.8.0
GOFUMPT_VERSION := v0.9.2
BUF_VERSION := v1.62.1
GRPC_GATEWAY_VERSION := v2.27.4
PROTOC_GEN_ENT_VERSION := v0.7.0
KIND_VERSION := v0.31.0
DOCKER_POSTGRESQL_NAME := product-reviews-postgresql
DOCKER_POSTGRESQL_VERSION := 15
DOCKER_RABBITMQ_NAME := rabbitmq

KUBE_NAMESPACE := product-reviews

# Postgres DB configuration and credentials for testing. This mimics the Aurora
# production environment.
export PGHOST=localhost
export PGPORT=5432
export PGSSLMODE=disable
export PGDATABASE=postgres
export PGUSER=admin
export PGPASSWORD=pass

.PHONY: help
help: # Credits to https://gist.github.com/prwhite/8168133 for this handy oneliner
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

atlas-install: ## Installs Atlas tool for generating migrations
	curl -sSf https://atlasgo.sh | sh

buf-install: ## Installs buf to convert protobuf into Golang code
	go install github.com/bufbuild/buf/cmd/buf@${BUF_VERSION}

buf-generate: clean-vendor buf-install buf-update ## Generates Golang-driven bindings out of Protobuf
	mkdir -p internal/ent/schema
	buf generate --exclude-path api/v1/ent --path api/v1/product_reviews.proto

buf-update: ## Updates the buf dependencies
	buf dep update

buf-lint: ## Runs linters against Protobuf
	buf lint --path api/v1/product_reviews.proto

buf-breaking: ## Checks Protobuf schema on breaking changes
	buf breaking --against '.git#branch=main'

generate: buf-generate ## Generates all necessary code bindings
	go generate ./internal/ent

build: go-tidy build-product-reviews ## Builds all code

build-product-reviews: ## Build the Go binary for product reviews system
	go build -mod=vendor -o build/_output/${POC_NAME} ./cmd/product-reviews.go

deps: buf-install go-linters-install atlas-install kind-install ## Installs developer prerequisites for this project
	go get github.com/grpc-ecosystem/grpc-gateway/v2@${GRPC_GATEWAY_VERSION}
	go install entgo.io/contrib/entproto/cmd/protoc-gen-ent@${PROTOC_GEN_ENT_VERSION}
	go install mvdan.cc/gofumpt@${GOFUMPT_VERSION}

kind-install: ## Installs kind to the system
	go install sigs.k8s.io/kind@${KIND_VERSION}

atlas-inspect: ## Inspect connection with DB with atlas
	atlas schema inspect --url "postgresql://${PGUSER}:${PGPASSWORD}@localhost:${PGPORT}/${PGDATABASE}?search_path=public" --format "OK"

migration-apply: ## Uploads migration to the running DB instance
	$(MAKE) db-start
	sleep 5;
	atlas migrate apply --dir file://internal/ent/migrate/migrations \
      --url postgresql://${PGUSER}:${PGPASSWORD}@localhost:${PGPORT}/${PGDATABASE}?search_path=public

migration-hash: ## Hashes the atlas checksum to correspond to the migration
	atlas migrate hash --dir file://internal/ent/migrate/migrations

migration-generate: ## Generate DB migration "make migration-generate MIGRATION=<migration-name>"
	@if test -z $(MIGRATION); then echo "Please specify migration name" && exit 1; fi
	$(MAKE) db-start
	sleep 5;
	atlas migrate diff $(MIGRATION) \
  		--dir "file://internal/ent/migrate/migrations" \
  		--to "ent://internal/ent/schema" \
  		--dev-url "docker://postgres/15/${PGDATABASE}?search_path=public"
	$(MAKE) db-stop

db-start: ## Starts PostgreSQL Docker instance with uploaded migration
	- $(MAKE) db-stop
	docker run --name ${DOCKER_POSTGRESQL_NAME} --rm -p ${PGPORT}:${PGPORT} -e POSTGRES_PASSWORD=${PGPASSWORD} -e POSTGRES_DB=${PGDATABASE} -e POSTGRES_USER=${PGUSER} -d postgres:$(DOCKER_POSTGRESQL_VERSION)

db-stop: ## Stops PostgreSQL Docker instance
	docker stop ${DOCKER_POSTGRESQL_NAME}

rabbitmq-start: ## Starts RabbitMQ Docker instance
	- $(MAKE) rabbitmq-stop
	docker run --name ${DOCKER_RABBITMQ_NAME} -p 5672:5672 -p 15672:15672 --rm -d rabbitmq:latest

rabbitmq-stop: ## Stops RabbitMQ Docker instance
	docker stop ${DOCKER_RABBITMQ_NAME}

go-linters-install: ## Install linters locally for verification
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin ${GOLANGCI_LINTERS_VERSION}

go-linters: go-linters-install ## Perform linting to verify codebase
	golangci-lint run --timeout 5m

govulncheck-install: ## Installs latest govulncheck tool
	go install golang.org/x/vuln/cmd/govulncheck@latest

govulncheck: govulncheck-install ## Runs govulncheck on the current codebase
	govulncheck ./...

go-vet: ## Searching for suspicious constructs in Go code
	go vet ./...

go-test: rabbitmq-start bring-up-db ## Run unit tests present in the codebase
	mkdir -p tmp
	go test -coverprofile=./tmp/test-cover.out -race ./...
	$(MAKE) db-stop
	$(MAKE) rabbitmq-stop

test-ci: generate buf-lint build go-vet govulncheck go-linters go-test ## Test the whole codebase (mimics CI/CD)

run: go-tidy build-product-reviews bring-up-db ## Runs compiled network device monitoring service
	sleep 5;
	./build/_output/${POC_NAME}

consume: ## Runs listener on RabbitMQ channel
	go run cmd/helpers/consumer.go

run-rest-list-products: ## Runs CURL command and lists all products
	curl -v http://localhost:50052/v1/product/all

bring-up-db: migration-apply ## Start DB and upload migrations to it

image: ## Builds a Docker image for Network Device monitoring service
	docker build . -f build/Dockerfile \
		-t ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}

images: image ## Builds Docker images for monitoring service and for device simulator

docker-run: image bring-up-db ## Runs compiled binary in a Docker container
	docker run --net=host --rm ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}

kind: images ## Builds Docker image for API Gateway and loads it to the currently configured kind cluster
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}

create-cluster: delete-cluster ## Creates cluster with KinD
	kind create cluster

delete-cluster: ## Removes KinD cluster
	kind delete cluster

kubectl-delete-namespace: ## Deletes namespace with kubectl
	kubectl delete namespace ${KUBE_NAMESPACE}

deploy-product-reviews: ## Deploys Product Reviews service Helm charts
	helm upgrade --install product-reviews ./helm-charts/product-reviews --namespace ${KUBE_NAMESPACE} --create-namespace --wait

update-product-reviews-charts: ## Updates dependencies for Product Reviews service charts (i.e., pull PostgreSQL dependency)
	helm dependency update ./helm-charts/product-reviews

helm-test: ## Runs helm testo for a Product Reviews service
	helm test product-reviews --namespace ${KUBE_NAMESPACE}

poc: kind-install create-cluster go-tidy kind update-product-reviews-charts deploy-product-reviews ## Runs PoC in Kubernetes cluster

poc-test: ## Runs PoC in Kubernetes cluster
poc-test: kind-install create-cluster go-tidy kind update-product-reviews-charts deploy-product-reviews deploy-product-reviews helm-test delete-cluster

go-tidy: ## Runs go mod related commands
	go mod tidy
	go mod vendor

clean-vendor: ## Cleans only vendor folder
	rm -rf ./vendor

clean: ## Remove all the build artifacts
	rm -rf ./build/_output ./vendor ./tmp
	go clean -testcache
