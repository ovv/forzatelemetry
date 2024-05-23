# Golang build tags
BUILDTAGS := "osusergo netgo"

# Golang build LDFLAGS
LDFLAGS := "-extldflags '-static'"

# Enabled Golang experiments
GOEXPERIMENT := "rangefunc"

# Developement configuration
LOCAL_POSTGRES_DNS := "postgres://postgres:owncgbwpwmyyiq@127.0.0.1:5432/postgres?sslmode=disable"
LOCAL_GRAFANA_BASE_URL := "http://localhost:3000/d/bdwp0tg8b21oge/race?orgId=1&kiosk=1"

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
	dev/postgres.sh

# Connect to the running postgres instance with pgcli
pgcli:
	pgcli --pgclirc dev/pgclirc -D forzatelemetry

# Start local grafana instance
grafana:
	dev/grafana.sh

# Start local server instance 
dev $POSTGRES_DSN=LOCAL_POSTGRES_DNS $GRAFANA_BASE_URL=LOCAL_GRAFANA_BASE_URL: build
	build/x86_64/forzatelemetry
