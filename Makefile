default:
	sh -c "'scripts/build.sh'"

deps:
	go get ./...

fmt:
	gofmt -w .

.PHONY: default
