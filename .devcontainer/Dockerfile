FROM mcr.microsoft.com/devcontainers/go:1.26-bookworm

# Use `zsh` as the default shell.
RUN chsh -s $(which zsh) vscode

# `vscode` is a built-in user in the base image.
USER vscode

# Homebrew installer must be run as non-root user, as instructed by the maintainer.
# https://docs.brew.sh/FAQ#why-does-homebrew-say-sudo-is-bad
RUN NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" && \
    # Add `brew` command to the PATH of zsh (default shell).
    echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv zsh)"' >> /home/vscode/.zshrc && \
    # Enable access to `brew` command and install packages.
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)" && \
    brew install httpie pgcli gh go-task && \
    # Remove all cache files to reduce image size.
    brew cleanup --prune=all
