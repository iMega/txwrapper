REPO = github.com/imega/txwrapper
CWD = /go/src/$(REPO)
GO_IMG = golang:1.17.2-alpine3.14

test: lint unit

lint:
	@docker run --rm -t -v $(CURDIR):$(CWD) -w $(CWD) \
		golangci/golangci-lint golangci-lint run

unit:
	@docker run --rm -w $(CWD) -v $(CURDIR):$(CWD) \
		$(GO_IMG) sh -c "\
			apk add --upd alpine-sdk && \
			go list ./... | grep -v 'tests' | xargs go test -vet=off -coverprofile cover.out \
		"
