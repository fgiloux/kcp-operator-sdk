package templates

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Makefile{}

// Makefile scaffolds a file that defines project management CLI commands
type Makefile struct {
	machinery.TemplateMixin
	machinery.ComponentConfigMixin
	machinery.ProjectNameMixin
	machinery.DomainMixin

	// Registry is the container registry where the image gets stored
	Registry string
	// Image is controller manager image name
	Image string
	// BoilerplatePath is the path to the boilerplate file
	BoilerplatePath string
	// Controller tools version to use in the project
	ControllerToolsVersion string
	// Kustomize version to use in the project
	KustomizeVersion string
	// ControllerRuntimeVersion version to be used to download the envtest setup script
	ControllerRuntimeVersion string
	// kcp version used for testing
	KCPVersion string
	// yq version used for parsing yaml files
	YQVersion string
	// version of the Kubebuilder assets used by EnvTest
	EnvTestK8s string
}

// SetTemplateDefaults implements file.Template
func (f *Makefile) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "Makefile"
	}

	f.TemplateBody = makefileTemplate

	f.IfExistsAction = machinery.Error

	if f.Image == "" {
		f.Image = "controller:latest"
	}

	return nil
}

//nolint:lll
const makefileTemplate = `
##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# Image registry and name used by all targets building/pushing images
REGISTRY ?= {{ .Registry }}
IMG ?= {{ .Image }}
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = {{ .EnvTestK8s }}

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# kcp specific
APIEXPORT_PREFIX ?= today

.PHONY: all
all: build

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: apiresourceschemas
apiresourceschemas: $(KUSTOMIZE) ## Convert CRDs from config/crds to APIResourceSchemas. Specify APIEXPORT_PREFIX as needed.
	$(KUSTOMIZE) build config/crd | kubectl kcp crd snapshot -f - --prefix $(APIEXPORT_PREFIX) > config/kcp/$(APIEXPORT_PREFIX).apiresourceschemas.yaml

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile={{printf "%q" .BoilerplatePath}} paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet $(ENVTEST) ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

ARTIFACT_DIR ?= .test

.PHONY: test-e2e
test-e2e: $(ARTIFACT_DIR)/kind.kubeconfig kcp-synctarget ready-deployment run-test-e2e ## Set up prerequisites and run end-to-end tests on a cluster.

.PHONY: run-test-e2e
run-test-e2e: ## Run end-to-end tests on a cluster.
	go test ./test/e2e/... --kubeconfig $(abspath $(ARTIFACT_DIR)/kcp.kubeconfig) --workspace $(shell $(KCP_KUBECTL) kcp workspace . --short)

.PHONY: ready-deployment
ready-deployment: KUBECONFIG = $(ARTIFACT_DIR)/kcp.kubeconfig
ready-deployment: kind-image install deploy apibinding ## Deploy the controller-manager and wait for it to be ready.
	$(KCP_KUBECTL) --namespace "controller-runtime-example-system" rollout status deployment/controller-runtime-example-controller-manager
# TODO(skuznets|ncdc): this APIBinding is not needed, but here only to work around https://github.com/kcp-dev/kcp/issues/1183 - remove it once that is fixed

.PHONY: apibinding
apibinding:
	$( eval WORKSPACE = $(shell $(KCP_KUBECTL) kcp workspace . --short) )
	sed 's/WORKSPACE/$(WORKSPACE)/' ./test/e2e/apibinding.yaml | $(KCP_KUBECTL) apply -f -
	$(KCP_KUBECTL) wait --for=condition=Ready apibinding/data.my.domain

.PHONY: kind-image
kind-image: docker-build ## Load the controller-manager image into the kind cluster.
	kind load docker-image $(REGISTRY)/$(IMG) --name controller-runtime-example

$(ARTIFACT_DIR)/kind.kubeconfig: $(ARTIFACT_DIR) ## Run a kind cluster and generate a $KUBECONFIG for it.
	@if ! kind get clusters --quiet | grep --quiet controller-runtime-example; then kind create cluster --name controller-runtime-example; fi
	kind get kubeconfig --name controller-runtime-example > $(ARTIFACT_DIR)/kind.kubeconfig

$(ARTIFACT_DIR): ## Create a directory for test artifacts.
	mkdir -p $(ARTIFACT_DIR)

KCP_KUBECTL ?= PATH=$(LOCALBIN):$(PATH) KUBECONFIG=$(ARTIFACT_DIR)/kcp.kubeconfig kubectl
KIND_KUBECTL ?= kubectl --kubeconfig $(ARTIFACT_DIR)/kind.kubeconfig

.PHONY: kcp-synctarget
kcp-synctarget: kcp-workspace $(ARTIFACT_DIR)/syncer.yaml $(YQ) ## Add the kind cluster to kcp as a target for workloads.
	$(KIND_KUBECTL) apply -f $(ARTIFACT_DIR)/syncer.yaml
	$(eval DEPLOYMENT_NAME = $(shell $(YQ) 'select(.kind=="Deployment") | .metadata.name' < $(ARTIFACT_DIR)/syncer.yaml ))
	$(eval DEPLOYMENT_NAMESPACE = $(shell $(YQ) 'select(.kind=="Deployment") | .metadata.namespace' < $(ARTIFACT_DIR)/syncer.yaml ))
	$(KIND_KUBECTL) --namespace $(DEPLOYMENT_NAMESPACE) rollout status deployment/$(DEPLOYMENT_NAME)
	@if [[ ! -s $(ARTIFACT_DIR)/syncer.log ]]; then ( $(KIND_KUBECTL) --namespace $(DEPLOYMENT_NAMESPACE) logs deployment/$(DEPLOYMENT_NAME) -f >$(ARTIFACT_DIR)/syncer.log 2>&1 & ); fi
	$(KCP_KUBECTL) wait --for=condition=Ready synctarget/controller-runtime

$(ARTIFACT_DIR)/syncer.yaml: ## Create the SyncTarget and generate the manifests necessary to register the kind cluster with kcp.
	$(KCP_KUBECTL) kcp workload sync controller-runtime --resources services --syncer-image ghcr.io/kcp-dev/kcp/syncer:v$(KCP_VERSION) --output-file $(ARTIFACT_DIR)/syncer.yaml

.PHONY: kcp-workspace
kcp-workspace: $(KUBECTL_KCP) kcp-server ## Create a workspace in kcp for the controller-manager.
	$(KCP_KUBECTL) kcp workspace use '~'
	@if ! $(KCP_KUBECTL) kcp workspace use controller-runtime-example; then $(KCP_KUBECTL) kcp workspace create controller-runtime-example --type universal --enter; fi

.PHONY: kcp-server
kcp-server: $(KCP) $(ARTIFACT_DIR)/kcp ## Run the kcp server.
	@if [[ ! -s $(ARTIFACT_DIR)/kcp.log ]]; then ( $(KCP) start -v 5 --root-directory $(ARTIFACT_DIR)/kcp --kubeconfig-path $(ARTIFACT_DIR)/kcp.kubeconfig --audit-log-maxsize 1024 --audit-log-mode=batch --audit-log-batch-max-wait=1s --audit-log-batch-max-size=1000 --audit-log-batch-buffer-size=10000 --audit-log-batch-throttle-burst=15 --audit-log-batch-throttle-enable=true --audit-log-batch-throttle-qps=10 --audit-policy-file ./test/e2e/audit-policy.yaml --audit-log-path $(ARTIFACT_DIR)/audit.log >$(ARTIFACT_DIR)/kcp.log 2>&1 & ); fi
	@while true; do if [[ ! -s $(ARTIFACT_DIR)/kcp.kubeconfig ]]; then sleep 0.2; else break; fi; done
	@while true; do if ! kubectl --kubeconfig $(ARTIFACT_DIR)/kcp.kubeconfig get --raw /readyz >$(ARTIFACT_DIR)/kcp.probe.log 2>&1; then sleep 0.2; else break; fi; done

$(ARTIFACT_DIR)/kcp: ## Create a directory for the kcp server data.
	mkdir -p $(ARTIFACT_DIR)/kcp

.PHONY: test-e2e-cleanup
test-e2e-cleanup: ## Clean up processes and directories from an end-to-end test run.
	kind delete cluster --name controller-runtime-example || true
	rm -rf $(ARTIFACT_DIR) || true
	pkill -sigterm kcp || true
	pkill -sigterm kubectl || true

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

NAME_PREFIX ?= {{ .ProjectName }}
APIEXPORT_NAME ?= {{ .Domain }}

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go --api-export-name $(NAME_PREFIX)$(APIEXPORT_NAME)

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	docker build -t ${REGISTRY}/${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${REGISTRY}/${IMG}

# PLATFORMS defines the target platforms for the manager image being build to support multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to the registry configured
# To properly supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${REGISTRY}/${IMG} -f Dockerfile.cross
	- docker buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests $(KUSTOMIZE) ## Install APIResourceSchemas and APIExport into kcp (using $KUBECONFIG or ~/.kube/config).
	$(KUSTOMIZE) build config/kcp | kubectl --kubeconfig $(KUBECONFIG) apply -f -

.PHONY: uninstall
uninstall: manifests $(KUSTOMIZE) ## Uninstall APIResourceSchemas and APIExport from kcp (using $KUBECONFIG or ~/.kube/config). Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/kcp | kubectl --kubeconfig $(KUBECONFIG) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests $(KUSTOMIZE) ## Deploy controller 
	cd config/manager && $(KUSTOMIZE) edit set image controller=${REGISTRY}/${IMG}
	$(KUSTOMIZE) build config/default | kubectl --kubeconfig $(KUBECONFIG) apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl --kubeconfig $(KUBECONFIG) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-crd
deploy-crd: manifests $(KUSTOMIZE) ## Deploy controller
	cd config/manager && $(KUSTOMIZE) edit set image controller=${REGISTRY}/${IMG}
	$(KUSTOMIZE) build config/default-crd | kubectl --kubeconfig $(KUBECONFIG) apply -f - || true

.PHONY: undeploy-crd
undeploy-crd: ## Undeploy controller. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default-crd | kubectl --kubeconfig $(KUBECONFIG) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
KCP ?= $(LOCALBIN)/kcp
KUBECTL_KCP ?= $(LOCALBIN)/kubectl-kcp
YQ ?= $(LOCALBIN)/yq

## Tool Versions
KUSTOMIZE_VERSION ?= {{ .KustomizeVersion }}
CONTROLLER_TOOLS_VERSION ?= {{ .ControllerToolsVersion }}
KCP_VERSION ?= {{ .KCPVersion }}
YQ_VERSION ?= {{ .YQVersion }}

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: yq
yq: $(YQ) ## Download yq locally if necessary.
$(YQ): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/v4@$(YQ_VERSION)

OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

.PHONY: kcp
kcp: $(KCP) ## Download kcp locally if necessary.
$(KCP): $(LOCALBIN)
	curl -L -s -o - https://github.com/kcp-dev/kcp/releases/download/v$(KCP_VERSION)/kcp_$(KCP_VERSION)_$(OS)_$(ARCH).tar.gz | tar --directory $(LOCALBIN)/../ -xvzf - bin/kcp
	touch $(KCP) # we download an "old" file, so make will re-download to refresh it unless we make it newer than the owning dir

.PHONY: kubectl_kcp
kubectl_kcp: $(KUBECTL_KCP) ## Download kcp kubectl plugins locally if necessary.
$(KUBECTL_KCP): $(LOCALBIN)
	curl -L -s -o - https://github.com/kcp-dev/kcp/releases/download/v$(KCP_VERSION)/kubectl-kcp-plugin_$(KCP_VERSION)_$(OS)_$(ARCH).tar.gz | tar --directory $(LOCALBIN)/../ -xvzf - bin
	touch $(KUBECTL_KCP) # we download an "old" file, so make will re-download to refresh it unless we make it newer than the owning dir
`
