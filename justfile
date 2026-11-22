bin := "bin/ggu"
version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

[private]
default:
    @just --list

[group('build')]
[doc('Build the ggu binary into bin/')]
build:
    mkdir -p bin
    go build -trimpath -ldflags "-s -w -X github.com/f-gillmann/gogituser/cmd.version={{version}}" -o {{bin}} ./cmd/ggu

[group('build')]
[doc('Remove build output, including goreleaser snapshots')]
clean:
    rm -rf bin dist completions

[group('dev')]
[doc('Run ggu with the given arguments')]
run *args:
    go run ./cmd/ggu {{args}}

[group('dev')]
[doc('Format all Go files')]
fmt:
    gofmt -w .

[group('dev')]
[doc('Run go vet and check formatting')]
lint:
    go vet ./...
    @unformatted="$(gofmt -l .)"; [ -z "$unformatted" ] || { printf "not formatted:\n%s\n" "$unformatted"; exit 1; }

[group('dev')]
[doc('Run all tests')]
test:
    go test ./...
