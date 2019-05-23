# list only csi source code directories
PACKAGES = $(shell go list ./... | grep -v 'vendor\|pkg/client/generated\|integration-tests')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# API_PKG sets namespace where the API resources are defined
API_PKG := github.com/openebs/csi/pkg

# ALL_API_GROUPS has the list of all API resources from various groups
ALL_API_GROUPS=\
	openebs.io/runtask/v1beta1 \
	openebs.io/openebscluster/v1alpha1 \
	openebs.io/catalog/v1alpha1 \
	openebs.io/kubeassert/v1alpha1

# API_GROUPS sets api version of the resources exposed by csi
ifeq (${API_GROUPS}, )
  API_GROUPS = openebs.io/v1alpha1
  export API_GROUPS
endif

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	gopkg.in/matm/v1/gocov-html \
	github.com/ugorji/go/codec/codecgen \
	gopkg.in/alecthomas/gometalinter.v1 \
	github.com/golang/protobuf/protoc-gen-go

# list only our .go files i.e. exlcudes any .go files from the vendor directory
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = ci
  export IMAGE_TAG
endif

ifeq (${TRAVIS_TAG}, )
  BASE_TAG = ci
  export BASE_TAG
else
  BASE_TAG = ${TRAVIS_TAG}
  export BASE_TAG
endif

CSTOR_BASE_IMAGE= openebs/cstor-base:${BASE_TAG}

# Specify the name for the binaries
WEBHOOK=admission-server
CSI_DRIVER=csi-driver

# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

all: csi-driver-image

initialize: bootstrap

deps:
	dep ensure

clean:
	go clean -testcache
	rm -rf bin
	rm -rf ${GOPATH}/bin/${CSI_DRIVER}
	rm -rf ${GOPATH}/pkg/*

release:
	@$(MAKE) bin

# Run the bootstrap target once before trying cov
cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	@cat /tmp/coverage.html

test: format
	@echo "--> Running go test" ;
	@go test $(PACKAGES)

cover:
	go list ./... | grep -v vendor | xargs -n1 go test --cover

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

# Target to run gometalinter in Travis (deadcode, golint, errcheck, unconvert, goconst)
golint-travis:
	@gometalinter.v1 --install
	-gometalinter.v1 --config=metalinter.config ./...

# Run the bootstrap target once before trying gometalinter in Develop environment
golint:
	@gometalinter.v1 --install
	@gometalinter.v1 --vendor --deadline=600s ./...

vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@echo "--> Running go tool vet ..."
	@go tool vet $(VETARGS) ${GOFILES_NOVENDOR} ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "[LINT] Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
	fi

	@git grep -n `echo "log"".Print"` | grep -v 'vendor/' ; if [ $$? -eq 0 ]; then \
		echo "[LINT] Found "log"".Printf" calls. These should use Maya's logger instead."; \
	fi

# Bootstrap the build by downloading additional tools
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "+ Installing $$tool" ; \
		go get -u $$tool; \
	done

# code generation for custom resources
kubegen2: deepcopy2 clientset2 lister2 informer2

# code generation for custom resources
kubegen: deepcopy clientset lister informer kubegen2

# code generation for custom resources and protobuf
generated_files: kubegen protobuf

# builds vendored version of deepcopy-gen tool
# deprecate once the old pkg/apis/ folder structure is removed
deepcopy:
	@go install ./vendor/k8s.io/code-generator/cmd/deepcopy-gen
	@echo "+ Generating deepcopy funcs for $(API_GROUPS)"
	@deepcopy-gen \
		--input-dirs $(API_PKG)/apis/$(API_GROUPS) \
		--output-file-base zz_generated.deepcopy \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

# builds vendored version of deepcopy-gen tool
deepcopy2:
	@go install ./vendor/k8s.io/code-generator/cmd/deepcopy-gen
	@for apigrp in  $(ALL_API_GROUPS) ; do \
		echo "+ Generating deepcopy funcs for $$apigrp" ; \
		deepcopy-gen \
			--input-dirs $(API_PKG)/apis/$$apigrp \
			--output-file-base zz_generated.deepcopy \
			--go-header-file ./buildscripts/custom-boilerplate.go.txt; \
	done

# builds vendored version of client-gen tool
# deprecate once the old pkg/apis/ folder structure is removed
clientset:
	@go install ./vendor/k8s.io/code-generator/cmd/client-gen
	@echo "+ Generating clientsets for $(API_GROUPS)"
	@client-gen \
		--fake-clientset=true \
		--input $(API_GROUPS) \
		--input-base $(API_PKG)/apis \
		--clientset-path $(API_PKG)/client/generated/clientset \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

# builds vendored version of client-gen tool
clientset2:
	@go install ./vendor/k8s.io/code-generator/cmd/client-gen
	@for apigrp in  $(ALL_API_GROUPS) ; do \
		echo "+ Generating clientsets for $$apigrp" ; \
		client-gen \
			--fake-clientset=true \
			--input $$apigrp \
			--input-base $(API_PKG)/apis \
			--clientset-path $(API_PKG)/client/generated/$$apigrp/clientset \
			--go-header-file ./buildscripts/custom-boilerplate.go.txt; \
	done

# builds vendored version of lister-gen tool
# deprecate once the old pkg/apis/ folder structure is removed
lister:
	@go install ./vendor/k8s.io/code-generator/cmd/lister-gen
	@echo "+ Generating lister for $(API_GROUPS)"
	@lister-gen \
		--input-dirs $(API_PKG)/apis/$(API_GROUPS) \
		--output-package $(API_PKG)/client/generated/lister \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

# builds vendored version of lister-gen tool
lister2:
	@go install ./vendor/k8s.io/code-generator/cmd/lister-gen
	@for apigrp in  $(ALL_API_GROUPS) ; do \
		echo "+ Generating lister for $$apigrp" ; \
		lister-gen \
			--input-dirs $(API_PKG)/apis/$$apigrp \
			--output-package $(API_PKG)/client/generated/$$apigrp/lister \
			--go-header-file ./buildscripts/custom-boilerplate.go.txt; \
	done

# builds vendored version of informer-gen tool
# deprecate once the old pkg/apis/ folder structure is removed
informer:
	@go install ./vendor/k8s.io/code-generator/cmd/informer-gen
	@echo "+ Generating informer for $(API_GROUPS)"
	@informer-gen \
		--input-dirs $(API_PKG)/apis/$(API_GROUPS) \
		--output-package $(API_PKG)/client/generated/informer \

#Use this to build csi-driver
csi-driver:
	@echo "----------------------------"
	@echo "--> csi-driver         "
	@echo "----------------------------"
	@PNAME="csi-driver" CTLNAME=${CSI_DRIVER} sh -c "'$(PWD)/buildscripts/build.sh'"


csi-driver-image: csi-driver
	@echo "----------------------------"
	@echo "--> csi-driver image         "
	@echo "----------------------------"
	@cp bin/csi-driver/${CSI_DRIVER} buildscripts/csi-driver/
	cd buildscripts/csi-driver && sudo docker build -t payes/csi-driver:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/csi-driver/${CSI_DRIVER}

