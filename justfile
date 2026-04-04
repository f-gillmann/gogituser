build:
    mkdir -p bin
    go build -o bin/gogituser .

install: build
    cp bin/gogituser ~/.local/bin/gogituser

uninstall:
    rm ~/.local/bin/gogituser                       || true
    rm ~/.zsh/completions/_gogituser                || true
    rm ~/.bash_completion.d/gogituser               || true
    rm ~/.config/fish/completions/gogituser.fish    || true

completions:
    [ -d ~/.zsh/completions ]                  && gogituser completion zsh  > ~/.zsh/completions/_gogituser                || true
    [ -d ~/.bash_completion.d ]                && gogituser completion bash > ~/.bash_completion.d/gogituser               || true
    [ -d ~/.config/fish/completions ]          && gogituser completion fish > ~/.config/fish/completions/gogituser.fish    || true
