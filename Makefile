.PHONY: generate
generate:
	@sha=$$(grep -o 'beats-sha=[0-9a-f]*' go.mod | cut -d= -f2); \
	go get github.com/elastic/beats/v7@$$sha; \
	go mod tidy
