# List the available justfile recipes.
@default:
  just --list

# Format, vet, and test Go code.
check:
	go fmt ./...
	go vet ./...
	GOEXPERIMENT=loopvar go test ./... -cover

# Verbosely format, vet, and test Go code.
checkv:
	go fmt ./...
	go vet ./...
	GOEXPERIMENT=loopvar go test -v ./... -cover

# Lint code using staticcheck.
lint:
	staticcheck -f stylish ./...

# Test and provide HTML coverage report.
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# List the outdated go modules.
outdated:
  go list -u -m all
