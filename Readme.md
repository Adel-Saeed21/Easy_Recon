<div align="center">

```
███████╗ █████╗ ███████╗██╗   ██╗    ██████╗ ███████╗ ██████╗ ██████╗ ███╗   ██╗
██╔════╝██╔══██╗██╔════╝╚██╗ ██╔╝    ██╔══██╗██╔════╝██╔════╝██╔═══██╗████╗  ██║
█████╗  ███████║███████╗ ╚████╔╝     ██████╔╝█████╗  ██║     ██║   ██║██╔██╗ ██║
██╔══╝  ██╔══██║╚════██║  ╚██╔╝      ██╔══██╗██╔══╝  ██║     ██║   ██║██║╚██╗██║
███████╗██║  ██║███████║   ██║       ██║  ██║███████╗╚██████╗╚██████╔╝██║ ╚████║
╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝       ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝
```

<img src="https://readme-typing-svg.demolab.com?font=Fira+Code&size=22&pause=1000&color=00FF41&center=true&vCenter=true&width=600&lines=Automated+Subdomain+Recon+Tool;Built+with+Go+%F0%9F%90%B9;Fast+%7C+Clean+%7C+Modular;subfinder+%2B+httpx+Pipeline" alt="Typing SVG" />

<br/>

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Status](https://img.shields.io/badge/Status-Active-brightgreen?style=for-the-badge)
![Tools](https://img.shields.io/badge/Tools-subfinder%20%7C%20httpx-orange?style=for-the-badge)

</div>

---

## 🔍 What is easyRecon?

**easyRecon** is a lightweight, modular recon automation tool written in Go.  
It automates the classic subdomain enumeration pipeline:

```
Domain Input  ──►  subfinder  ──►  Dedup  ──►  httpx  ──►  Results
```

One command. Clean output. Everything saved automatically.

---

## ⚡ Quick Start

```bash
easyRecon -d domain.com
```

That's it. Everything else is handled automatically.

---

## 🔄 How It Works

```
┌─────────────────────────────────────────────────────────┐
│                      easyRecon -d domain.com             │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
          ┌──────────────────────────┐
          │  Create folder: domain/  │
          └──────────────┬───────────┘
                         │
                         ▼
          ┌──────────────────────────┐
          │  Check: subfinder & httpx│
          │  installed? ✅            │
          └──────────────┬───────────┘
                         │
                         ▼
          ┌──────────────────────────┐
          │  Run subfinder -d domain │
          │  → collect subdomains    │
          └──────────────┬───────────┘
                         │
                         ▼
          ┌──────────────────────────┐
          │  Remove duplicates       │
          │  Save → subdomains.txt   │
          └──────────────┬───────────┘
                         │
                         ▼
          ┌──────────────────────────┐
          │  Pipe to httpx -silent   │
          │  → filter alive hosts    │
          └──────────────┬───────────┘
                         │
                         ▼
          ┌──────────────────────────┐
          │  Save → alive.txt  🚀    │
          └──────────────────────────┘
```

---

## 📁 Output Structure

After running, you'll find:

```
domain.com/
├── subdomains.txt    # All unique subdomains from subfinder
└── alive.txt         # Live hosts confirmed by httpx
```

---

## 🧱 Project Structure

```
easyRecon/
├── main.go                  # Entry point & orchestration
├── runner/
│   ├── tool.go              # Tool interface definition
│   ├── subfinder.go         # Subfinder wrapper
│   └── httpx.go             # Httpx wrapper
└── utils/
    ├── dedup.go             # Remove duplicate entries
    ├── file.go              # Save results to disk
    └── tools.go             # Check tool availability
```

---

## 🛠️ Prerequisites

Make sure these tools are installed and in your `$PATH`:

| Tool | Install |
|------|---------|
| [subfinder](https://github.com/projectdiscovery/subfinder) | `go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest` |
| [httpx](https://github.com/projectdiscovery/httpx) | `go install github.com/projectdiscovery/httpx/cmd/httpx@latest` |

---

## 📦 Installation

```bash
# Clone the repo
git clone https://github.com/youruser/easyRecon.git
cd easyRecon

# Build
go build -o easyRecon .

# Run
./easyRecon -d example.com
```

---

## 🖥️ Example Output

```
[*] Starting recon on: example.com
[*] Running subfinder...
[+] Found 42 unique subdomains
[*] Running httpx...
[+] Found 18 alive domains
[✓] Recon completed! Results saved in: example.com/
```

---

## ⚙️ Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-d` | Target domain | required |

---

## 🤝 Contributing

Pull requests are welcome!  
If you want to add a new tool (e.g. `nmap`, `nuclei`), just implement the `Tool` interface:

```go
type Tool interface {
    Name() string
    Run(input []string) ([]string, error)
}
```

Drop your file in `runner/` and plug it into `main.go`. Done.

---

## 📄 License

MIT — use it, break it, improve it.

---

<div align="center">


<img src="https://readme-typing-svg.demolab.com?font=Fira+Code&size=14&pause=2000&color=00FF41&center=true&vCenter=true&width=400&lines=Happy+Hunting+%F0%9F%8E%AF;Stay+in+scope.+Always." alt="footer" />

</div>