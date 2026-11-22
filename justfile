bin := "bin/ggu"
version := `git describe --tags --always --dirty 2>/dev/null || echo dev`
completions_dir := "bin/completions"

[private]
default:
    @just --list

[group('build')]
[doc('Build the ggu binary into bin/')]
build:
    mkdir -p bin
    go build -trimpath -ldflags "-s -w -X github.com/f-gillmann/gogituser/cmd.version={{version}}" -o {{bin}} ./cmd/ggu

[group('build')]
[doc('Generate shell completions into bin/completions/')]
completions: build
    mkdir -p {{completions_dir}}
    {{bin}} completion bash > {{completions_dir}}/ggu.bash
    {{bin}} completion zsh  > {{completions_dir}}/_ggu
    {{bin}} completion fish > {{completions_dir}}/ggu.fish

[group('build')]
[doc('Remove build output')]
clean:
    rm -rf bin

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
