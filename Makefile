.PHONY: generate
# Renovate runs this as a post-upgrade task. The runner's Go can be older than the
# version in go.mod and has GOTOOLCHAIN=local, so let Go fetch the matching toolchain.
generate: export GOTOOLCHAIN = auto
generate:
	@tmp=$$(mktemp -d) && cp go.mod go.sum $$tmp && \
	sha=$$(grep 'elastic/beats/v7' go.mod | sed 's/.*-//') && \
	if go mod edit -droprequire=github.com/elastic/beats/v7 && \
		go get github.com/elastic/beats/v7@$$sha && \
		go mod tidy; then \
		rm -rf $$tmp; \
	else \
		cp $$tmp/go.mod $$tmp/go.sum . && rm -rf $$tmp; \
		exit 1; \
	fi
