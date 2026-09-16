.PHONY: generate
generate:
	@sha=$$(grep 'elastic/beats/v7' go.mod | sed 's/.*-//') && \
	go get github.com/elastic/beats/v7@$$sha && \
	go mod tidy
