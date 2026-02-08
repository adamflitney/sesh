# sesh

A simple CLI tool to browse your Git projects and open them in a tmux session with three pre-configured windows: neovim, opencode, and zsh.

## What does it do?

`sesh` scans your project directories for Git repositories, lets you fuzzy search and select one, then automatically creates (or attaches to) a tmux session with:

- **Window 1**: neovim opened to the project
- **Window 2**: opencode opened to the project  
- **Window 3**: a regular terminal in the project directory

If a session for that project already exists, it attaches to it instead of creating a new one.

## Installation

```bash
go install github.com/adamflitney/sesh@latest
```

Make sure `~/go/bin` is in your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

## Configuration

On first run, sesh creates `~/.config/sesh/config.yaml` with a default configuration:

```yaml
project_directories:
  - ~/dev
```

Edit this file to add more directories where your Git projects live:

```yaml
project_directories:
  - ~/dev
  - ~/work
  - ~/personal/projects
```

## Usage

```bash
sesh
```

This opens an interactive fuzzy search interface. Use:
- **↑/k** or **↓/j**: Navigate
- **Enter**: Select project
- **Esc/Ctrl+C**: Quit
- Type to fuzzy search

## Prerequisites

- Go 1.21+
- tmux
- neovim
- opencode

## License

MIT
