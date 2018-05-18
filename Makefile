.PHONY: debug
debug: generate
	go install -gcflags "-N -l" ./...

.PHONY: release
release: generate
	go install -tags "release full" ./...

.PHONY: clean
clean:
	go clean ./...

.PHONY: test
test:
	go test ./...

.PHONY: generate
generate:
	go generate -x ./...

.PHONY: run-gallery
run-gallery:
	go install ./gallery
	$(GOPATH)/bin/gallery
