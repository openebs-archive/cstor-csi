# list only csi source code directories
PACKAGES = $(shell go list ./... | grep -v 'vendor\|pkg/generated')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# Tools required for different make
# targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	gopkg.in/matm/v1/gocov-html \
	github.com/ugorji/go/codec/codecgen

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

# Specify the name for the binary
CSI_DRIVER=csi-driver

# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

.PHONY: all
all: test csi-driver-image

.PHONY: clean
clean:
	go clean -testcache
	rm -rf bin
	rm -rf ${GOPATH}/bin/${CSI_DRIVER}
	rm -rf ${GOPATH}/pkg/*

.PHONY: format
format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

.PHONY: test
test: format
	@echo "--> Running go test" ;
	@go test $(PACKAGES)

# Bootstrap downloads tools required
# during build
.PHONY: bootstrap
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "+ Installing $$tool" ; \
		go get -u $$tool; \
	done

# SRC_PKG is the path of code files
SRC_PKG := github.com/openebs/csi/pkg

# code generation for custom resources
.PHONY: kubegen
kubegen: kubegendelete deepcopy-install clientset-install lister-install informer-install
	@GEN_SRC=openebs.io/core/v1alpha1 GEN_DEST=core make deepcopy clientset lister informer
	@GEN_SRC=openebs.io/maya/v1alpha1 GEN_DEST=maya make deepcopy clientset lister informer

# deletes generated code by codegen
.PHONY: kubegendelete
kubegendelete:
	@rm -rf pkg/generated/clientset
	@rm -rf pkg/generated/lister
	@rm -rf pkg/generated/informer

.PHONY: deepcopy-install
deepcopy-install:
	@go install ./vendor/k8s.io/code-generator/cmd/deepcopy-gen

.PHONY: deepcopy
deepcopy:
	@echo "+ Generating deepcopy funcs for $(GEN_SRC)"
	@deepcopy-gen \
		--input-dirs $(SRC_PKG)/apis/$(GEN_SRC) \
		--output-file-base zz_generated.deepcopy \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

.PHONY: clientset-install
clientset-install:
	@go install ./vendor/k8s.io/code-generator/cmd/client-gen

.PHONY: clientset
clientset:
	@echo "+ Generating clientsets for $(GEN_SRC)"
	@client-gen \
		--fake-clientset=true \
		--input $(GEN_SRC) \
		--input-base $(SRC_PKG)/apis \
		--clientset-path $(SRC_PKG)/generated/clientset/$(GEN_DEST) \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

.PHONY: lister-install
lister-install:
	@go install ./vendor/k8s.io/code-generator/cmd/lister-gen

.PHONY: lister
lister:
	@echo "+ Generating lister for $(GEN_SRC)"
	@lister-gen \
		--input-dirs $(SRC_PKG)/apis/$(GEN_SRC) \
		--output-package $(SRC_PKG)/generated/lister \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

.PHONY: informer-install
informer-install:
	@go install ./vendor/k8s.io/code-generator/cmd/informer-gen

.PHONY: informer
informer:
	@echo "+ Generating informer for $(GEN_SRC)"
	@informer-gen \
		--input-dirs $(SRC_PKG)/apis/$(GEN_SRC) \
		--versioned-clientset-package $(SRC_PKG)/generated/clientset/$(GEN_DEST)/internalclientset \
		--listers-package $(SRC_PKG)/generated/lister/$(GEN_DEST) \
		--output-package $(SRC_PKG)/generated/informer/$(GEN_DEST) \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

.PHONY: csi-driver
csi-driver:
	@echo "--------------------------------"
	@echo "+ Building ${CSI_DRIVER}        "
	@echo "--------------------------------"
	@PNAME=${CSI_DRIVER} CTLNAME=${CSI_DRIVER} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: csi-driver-image
csi-driver-image: csi-driver
	@echo "--------------------------------"
	@echo "+ Generating ${CSI_DRIVER} image"
	@echo "--------------------------------"
	@cp bin/${CSI_DRIVER}/${CSI_DRIVER} buildscripts/${CSI_DRIVER}/
	cd buildscripts/${CSI_DRIVER} && sudo docker build -t openebs/${CSI_DRIVER}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${CSI_DRIVER}/${CSI_DRIVER}
