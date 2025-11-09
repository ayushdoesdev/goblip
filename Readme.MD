# GoBlip

GoBlip is a lightweight, dependency-free watcher for Go projects.  
It detects file changes (“blips”) and restarts your Go app automatically.

🧩 Gin support: GoBlip works seamlessly with the Gin framework.
Run your Gin server with goblip -- go run . and enjoy automatic reload on file changes.

---

## 🚀 Features

- No external dependencies — just Go’s standard library
- Cross-platform (macOS, Linux, Windows)
- Clean modular structure (`cmd/`, `internal/`)
- Configurable poll interval and file types
- Graceful shutdown via signals (Ctrl+C)

---

## 🧩 Installation

Clone or install:

```bash
go install github.com/ayushdoesdev/goblip/cmd/goblip@latest
```

Make sure your Go bin directory is in your PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## 🧠 Usage

```bash
# Default: go run your project automatically
goblip

# Run a specific command
goblip -- go run ./cmd/server

# Verbose output
goblip -v -- go run main.go

# Custom poll interval and extensions
goblip -interval 250ms -ext ".go,.tpl" -- go run .
```
