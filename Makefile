.PHONY: help build snapshot release test lint lint-fix clean generate

help:
	@echo "Available targets:"
	@echo "  build     - Build the binary for current platform"
	@echo "  snapshot  - Build snapshot release with GoReleaser"
	@echo "  release   - Create a release with GoReleaser"
	@echo "  test      - Run tests"
	@echo "  lint      - Run golangci-lint"
	@echo "  lint-fix  - Run golangci-lint with auto-fix"
	@echo "  clean     - Remove build artifacts"

generate:
	go tool mockery

build:
	goreleaser build --single-target --snapshot --clean

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean

test:
	go test -v -race -coverprofile=coverage.out ./... 2>&1 | tee test-output.log
	go tool cover -func=coverage.out | grep -E '^(total|.*/(parser|repository|checker|config)/)'
	@grep -E 'ok\s+.*/(checker|config|parser|repository)\s' test-output.log | \
		grep -v '/mocks' | \
		awk '{ for(i=1;i<=NF;i++) if($$i == "coverage:") { cov=$$(i+1)+0; pkg=$$2; sub(/.*\//, "", pkg); printf "%-12s statement coverage: %.1f%%\n", pkg, cov; if(cov < 80) { printf "FAIL: %s coverage %.1f%% is below 80%%\n", pkg, cov; fail=1 } } } END { if(fail) exit 1 }'
	@rm -f test-output.log

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

clean:
	rm -rf dist/
	rm -f coverage.out
