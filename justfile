# Golang build tags
BUILDTAGS := "osusergo netgo"

# Golang build LDFLAGS
LDFLAGS := "-extldflags '-static'"

# Enabled Golang experiments
GOEXPERIMENT := "rangefunc"

@_default:
    just --list

# Build server
build:
	GOEXPERIMENT="{{ GOEXPERIMENT }}" GOOS=linux GOARCH=amd64 go build -tags="{{ BUILDTAGS }}" -ldflags="{{ LDFLAGS }}" -o build/x86_64/forzatelemetry ./cmds/forzatelemetry

# Clean build directories
clean:
	rm -rf build
	go clean -testcache

# Run linters
lint:
	go fmt ./...

	go mod tidy -v 2>/dev/stdout | tee /dev/stderr | grep -c 'unused' | exit 0
	git diff --quiet --exit-code

	GOEXPERIMENT={{ GOEXPERIMENT }} go vet ./...

# Format files
format:
	gofmt -w -s cmds/* .

# Run tests
test *FLAGS: 
	GOEXPERIMENT={{ GOEXPERIMENT }} go test -count=1 -coverprofile=coverage.out {{ FLAGS }} ./...
	GOEXPERIMENT={{ GOEXPERIMENT }} go tool cover -func=coverage.out

# Run go code generation
generate:
	protoc -I=models/ --go_out=. models/*.proto
	go generate ./models

# Start local postgres instance
postgres:
	scripts/postgres.sh

# Start local grafana instance
grafana:
	scripts/grafana.sh

# Start local server instance 
dev: build
	build/x86_64/forzatelemetry
