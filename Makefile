# list only csi source code directories
PACKAGES = $(shell go list ./... | grep -v 'vendor\|pkg/generated')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	gopkg.in/matm/v1/gocov-html \
	github.com/ugorji/go/codec/codecgen

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

# Specify the name for the binaries
CSI_DRIVER=csi-driver

# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

all: format test csi-driver-image

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

# SRC_PKG sets the path of code files
SRC_PKG := github.com/openebs/csi/pkg

# code generation for custom resources
kubegen: kubegendelete deepcopy-install clientset-install lister-install informer-install
	@GEN_SRC=openebs.io/core/v1alpha1 GEN_DEST=core make deepcopy clientset lister informer
	@GEN_SRC=openebs.io/maya/v1alpha1 GEN_DEST=maya make deepcopy clientset lister informer

# deletes generated code by codegen
kubegendelete:
	@rm -rf pkg/generated/clientset
	@rm -rf pkg/generated/lister
	@rm -rf pkg/generated/informer

deepcopy-install:
	@go install ./vendor/k8s.io/code-generator/cmd/deepcopy-gen

deepcopy:
	@echo "+ Generating deepcopy funcs for $(GEN_SRC)"
	@deepcopy-gen \
		--input-dirs $(SRC_PKG)/apis/$(GEN_SRC) \
		--output-file-base zz_generated.deepcopy \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

clientset-install:
	@go install ./vendor/k8s.io/code-generator/cmd/client-gen

# builds vendored version of client-gen tool
clientset:
	@echo "+ Generating clientsets for $(GEN_SRC)"
	@client-gen \
		--fake-clientset=true \
		--input $(GEN_SRC) \
		--input-base $(SRC_PKG)/apis \
		--clientset-path $(SRC_PKG)/generated/clientset/$(GEN_DEST) \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

lister-install:
	@go install ./vendor/k8s.io/code-generator/cmd/lister-gen

# builds vendored version via lister-gen tool
lister:
	@echo "+ Generating lister for $(GEN_SRC)"
	@lister-gen \
		--input-dirs $(SRC_PKG)/apis/$(GEN_SRC) \
		--output-package $(SRC_PKG)/generated/lister \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

informer-install:
	@go install ./vendor/k8s.io/code-generator/cmd/informer-gen

# builds vendored version via informer tool
informer:
	@echo "+ Generating informer for $(GEN_SRC)"
	@informer-gen \
		--input-dirs $(SRC_PKG)/apis/$(GEN_SRC) \
		--versioned-clientset-package $(SRC_PKG)/generated/clientset/$(GEN_DEST)/internalclientset \
		--listers-package $(SRC_PKG)/generated/lister/$(GEN_DEST) \
		--output-package $(SRC_PKG)/generated/informer/$(GEN_DEST) \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

#Use this to build csi-driver
csi-driver:
	@echo "-----------------------------"
	@echo "+ Building csi-driver        "
	@echo "-----------------------------"
	@PNAME="csi-driver" CTLNAME=${CSI_DRIVER} sh -c "'$(PWD)/buildscripts/build.sh'"


csi-driver-image: csi-driver
	@echo "-----------------------------"
	@echo "+ Generating csi-driver image"
	@echo "-----------------------------"
	@cp bin/csi-driver/${CSI_DRIVER} buildscripts/csi-driver/
	cd buildscripts/csi-driver && sudo docker build -t openebs/csi-driver:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/csi-driver/${CSI_DRIVER}
