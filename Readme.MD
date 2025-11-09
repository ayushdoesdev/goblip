# GoBlip

GoBlip is a lightweight, dependency-free watcher for Go projects.  
It detects file changes (“blips”) and restarts your Go app automatically.

🧩 Gin support: GoBlip works seamlessly with the Gin framework.

Run your Gin server with goblip -- go run . and enjoy automatic reload on file changes.

---

💡 **Why GoBlip?**

The best part — **GoBlip uses zero external libraries.**  
It’s built entirely with the **Go standard library**, which makes it:

- ⚡ **Fast:** minimal overhead and instant startup  
- 🔒 **Safe:** no dependency vulnerabilities or supply-chain risks  
- 🧩 **Stable:** powered only by Go’s well-supported core packages  
- 🛠️ **Long-term reliable:** works across all future Go versions without maintenance headaches  

This means GoBlip will keep working for years — fast, clean, and dependency-free.


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

## 🤝 Contributing

Contributions are welcome!

1. **Fork** the repository  
2. **Create a feature branch:**
```bash
   git checkout -b feature/your-feature
```
3. Make your changes
4. Run formatting and vet:
```bash
  go fmt ./...
  go vet ./...
```
5. Commit and open a Pull Request
