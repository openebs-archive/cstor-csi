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
	github.com/ugorji/go/codec/codecgen \
	github.com/onsi/ginkgo/ginkgo \
	github.com/onsi/gomega/...

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
CSI_DRIVER=cstor-csi-driver

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

# deps ensures fresh go.mod and go.sum.
.PHONY: deps
deps:
	@go mod tidy
	@go mod verify

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
SRC_PKG := github.com/openebs/cstor-csi/pkg

# code generation for custom resources
.PHONY: kubegen
kubegen:
	./buildscripts/update-codegen.sh

# deletes generated code by codegen
.PHONY: kubegendelete
kubegendelete:
	@rm -rf pkg/client/clientset
	@rm -rf pkg/client/lister
	@rm -rf pkg/client/informer

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
	cd buildscripts/${CSI_DRIVER} && sudo docker build -t openebs/${CSI_DRIVER}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} . && docker tag openebs/${CSI_DRIVER}:${IMAGE_TAG} quay.io/openebs/${CSI_DRIVER}:${IMAGE_TAG}
	@rm buildscripts/${CSI_DRIVER}/${CSI_DRIVER}

# Push images
deploy-images:
	@DIMAGE="openebs/cstor-csi-driver" ./buildscripts/push

