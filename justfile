# Able v12

go_dir := "v12/interpreters/go"
cmd_able_dir := go_dir / "cmd/able"

# Copy kernel source into the embedded/ directory for go:embed
embed:
    mkdir -p {{cmd_able_dir}}/embedded/kernel/src
    cp v12/kernel/package.yml {{cmd_able_dir}}/embedded/kernel/
    cp v12/kernel/src/kernel.able {{cmd_able_dir}}/embedded/kernel/src/

# Build the able binary
build: embed
    cd {{go_dir}} && go build -o able ./cmd/able
    cp {{go_dir}}/able ./able

# Build with optimizations stripped for smaller binary
build-small: embed
    cd {{go_dir}} && go build -ldflags="-s -w" -o able ./cmd/able
    cp {{go_dir}}/able ./able

# Run all tests
test: embed
    cd {{go_dir}} && go test ./pkg/runtime/... ./pkg/interpreter/... ./pkg/compiler/...

# Run CLI tests
test-cli: embed
    cd {{go_dir}} && go test ./cmd/able/...

# Clean build artifacts
clean:
    rm -f able {{go_dir}}/able
    rm -rf {{cmd_able_dir}}/embedded/
