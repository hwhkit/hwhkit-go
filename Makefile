MODULES := \
  buildinfo config core observability hwhkit \
  tenant jwt ratelimit idempotency circuitbreaker scheduler \
  integration/postgres integration/redis integration/s3 \
  integration/mongodb integration/nats integration/qdrant integration/neo4j integration/oss \
  cmd/hwhkit examples/minimal examples/full-stack

.PHONY: tidy build test vet lint clean smoke

tidy:
	@for m in $(MODULES); do echo "--- tidy $$m ---"; (cd $$m && GOWORK=off go mod tidy) || exit 1; done

build:
	@for m in $(MODULES); do echo "--- build $$m ---"; (cd $$m && GOWORK=off go build ./...) || exit 1; done

test:
	@for m in $(MODULES); do echo "--- test $$m ---"; (cd $$m && GOWORK=off go test ./...) || exit 1; done

vet:
	@for m in $(MODULES); do echo "--- vet $$m ---"; (cd $$m && GOWORK=off go vet ./...) || exit 1; done

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

smoke: build
	go build -o bin/minimal ./examples/minimal
	@echo "run ./bin/minimal in one terminal; curl /health /health/ready /metrics /version /info in another"
