# Taskfile - Task Runner

This project uses [Taskfile](https://taskfile.dev) to manage common development tasks.

## Installation

```bash
# Install Taskfile (one-time setup)
go install github.com/go-task/task/v3/cmd/task@latest

# Verify installation
task --version
```

## Usage

```bash
# See all available tasks
task --list

# Or just run (shows help)
task

# Run a specific task
task test
task lint
task build
task run
```

## Available Tasks

- `task lint` - Run golangci-lint
- `task lint-fix` - Run golangci-lint with auto-fix
- `task test` - Run tests with coverage (fast)
- `task test-race` - Run tests with race detector (slower, catches concurrency bugs)
- `task test-coverage` - Generate HTML coverage report
- `task build` - Build the application
- `task run` - Run the application

## Why Taskfile?

- ✅ **Easier to read** - YAML syntax is cleaner than Makefile
- ✅ **Cross-platform** - Works on Windows, macOS, Linux
- ✅ **Better errors** - Clear error messages
- ✅ **Built-in help** - `task --list` shows all tasks
- ✅ **Go-native** - Written in Go, fits your stack

## Editing Tasks

The `Taskfile.yml` is easy to modify. Just edit the YAML file:

```yaml
tasks:
  my-task:
    desc: Description of what this task does
    cmds:
      - command to run
      - another command
```

That's it! Much simpler than Makefile syntax.

