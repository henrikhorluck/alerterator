
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: fmt vet
	go test ./... -coverprofile cover.out

# Build manager binary
manager: fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run ./main.go

# Install CRDs into a cluster
install:
	kubectl apply -f config/local-test/configmaps.yaml
	kubectl apply -f ../liberator/config/crd/bases/nais.io_alerts.yaml

# Uninstall CRDs from a cluster
uninstall:
	kubectl delete -f config/local-test/configmaps.yaml
	kubectl delete -f ../liberator/config/crd/bases/nais.io_alerts.yaml

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

