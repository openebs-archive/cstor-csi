
GOFLAGS=-mod=vendor
export GOFLAGS
GINKGO=$(GOPATH)/bin/ginkgo
KUBECTL=/usr/local/bin/kubectl


# If you want to change the Kubernetes version for e2e test, specify this variable from command line.
# e.g. $ make TEST_KUBERNETES_VERSION=1.17 test
TEST_KUBERNETES_VERSION=1.17

ifeq ($(TEST_KUBERNETES_VERSION),1.17)
KUBERNETES_VERSION=1.17.2
endif


e2e-test:
	  env \
		PATH=$(PATH) \
		GOPATH=$(GOPATH) \
		GO111MODULE=on \
		GOFLAGS=$(GOFLAGS) \
		TEST_KUBERNETES_VERSION=$(TEST_KUBERNETES_VERSION) \
		$(GINKGO) --failFast -v .

