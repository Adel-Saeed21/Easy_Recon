<div align="center">

```
███████╗ █████╗ ███████╗██╗   ██╗    ██████╗ ███████╗ ██████╗ ██████╗ ███╗   ██╗
██╔════╝██╔══██╗██╔════╝╚██╗ ██╔╝    ██╔══██╗██╔════╝██╔════╝██╔═══██╗████╗  ██║
█████╗  ███████║███████╗ ╚████╔╝     ██████╔╝█████╗  ██║     ██║   ██║██╔██╗ ██║
██╔══╝  ██╔══██║╚════██║  ╚██╔╝      ██╔══██╗██╔══╝  ██║     ██║   ██║██║╚██╗██║
███████╗██║  ██║███████║   ██║       ██║  ██║███████╗╚██████╗╚██████╔╝██║ ╚████║
╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝       ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝
```

<img src="https://readme-typing-svg.demolab.com?font=Fira+Code&size=22&pause=1000&color=00FF41&center=true&vCenter=true&width=600&lines=Automated+Subdomain+Recon+Tool;Built+with+Go+%F0%9F%90%B9;Concurrent+enumeration+%7C+httpx+%7C+Wayback+%7C+Arjun" alt="Typing SVG" />

<br/>

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Status](https://img.shields.io/badge/Status-Active-brightgreen?style=for-the-badge)
![Tools](https://img.shields.io/badge/Tools-subfinder%20%7C%20httpx%20%7C%20waymore-orange?style=for-the-badge)

</div>

---

## 🔍 What is easyRecon?

**easyRecon** is a Go recon automation tool. It runs **subdomain enumeration** with every supported tool you have installed (in parallel), **deduplicates** results, pipes hosts through **httpx**, then optionally collects **historical URLs** (Wayback) and runs **parameter discovery** with Arjun. Output is written under a folder named after the target domain.

---

## ⚡ Quick Start

```bash
go build -o easyRecon .
./easyRecon -d example.com
```

Use `-no-tg` if you only want the terminal (no Telegram bot).

---

## 🔄 How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                   easyRecon -d domain.com [-t N] [-no-tg]        │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
              ┌──────────────────────────────┐
              │  Create folder: domain/      │
              └──────────────┬───────────────┘
                             │
                             ▼
              ┌──────────────────────────────┐
              │  httpx must be in PATH       │
              │  ≥1 enum tool or exit        │
              └──────────────┬───────────────┘
                             │
                             ▼
    ┌────────────────────────────────────────────────────────┐
    │  Subdomain tools (each optional, run concurrently):     │
    │  subfinder · assetfinder · findomain · amass           │
    │  → merge → dedup → subdomains.txt                     │
    └────────────────────────┬─────────────────────────────┘
                             │
                             ▼
              ┌──────────────────────────────┐
              │  httpx on all subdomains      │
              │  → alive.txt                  │
              └──────────────┬───────────────┘
                             │
              ┌──────────────┴───────────────┐
              │  Optional (prompts):          │
              │  waybackurls / waymore → urls │
              │  arjun → params               │
              │  cleanup intermediate files   │
              └──────────────────────────────┘
```

---

## 📁 Output Structure

Typical layout after a full run:

```
domain.com/
├── subdomains.txt      # Unique subdomains (merged from enum tools)
├── alive.txt           # Live hosts (httpx)
├── urls.txt            # Optional — Wayback / waymore URLs
├── params.txt          # Optional — parameters from arjun
├── subfinder.txt       # Per-tool raw output (if that tool ran)
├── assetfinder.txt
├── findomain.txt
├── amass.txt
├── waybackurls.txt
└── waymore.txt
```

If you choose **cleanup** at the end, intermediate per-tool files and `subdomains.txt` are removed; **`alive.txt`**, **`urls.txt`**, and **`params.txt`** are kept when present.

---

## 🧱 Project Structure

```
easyRecon/
├── main.go                 # CLI, prompts, orchestration
├── bot.go                  # Telegram bot & .easyrecon_config.json
├── subdomain.go            # Subdomain tool wiring
├── urls.go                 # Wayback URL collection
├── hiddenParamters.go      # Arjun parameter discovery
├── runner/
│   ├── tool.go             # Tool interface
│   ├── subfinder.go
│   ├── assetfinder.go
│   ├── findomain.go
│   ├── amass.go
│   ├── httpx.go
│   ├── waybackurl.go
│   ├── waymore.go
│   └── arjun.go
└── utils/
    ├── remove_duplicate.go
    ├── save_to_file.go
    └── is_tool_installed.go
```

---

## 🛠️ Prerequisites

| Role | Tools | Notes |
|------|--------|--------|
| **Required** | [httpx](https://github.com/projectdiscovery/httpx) | Always used for probing. |
| **Subdomains** (need ≥1) | [subfinder](https://github.com/projectdiscovery/subfinder), [assetfinder](https://github.com/tomnomnom/assetfinder), [findomain](https://github.com/Findomain/Findomain), [amass](https://github.com/owasp-amass/amass) | Installed tools run **in parallel**. |
| **URLs** (optional) | [waybackurls](https://github.com/tomnomnom/waybackurls), [waymore](https://github.com/xnl-h4ck3r/waymore) | Prompted after httpx. |
| **Parameters** (optional) | [arjun](https://github.com/s0md3v/Arjun) | Prompted if `urls.txt` was produced. |

Example installs:

```bash
go install github.com/projectdiscovery/httpx/cmd/httpx@latest
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
```

---

## 📦 Installation

```bash
git clone https://github.com/Adel-Saeed21/Easy_Recon.git
cd Easy_Recon
go build -o easyRecon .
./easyRecon -d example.com
```

---

## 🤖 Telegram (optional)

On startup, easyRecon can use a Telegram bot so you can answer **y/n** prompts from your phone and mirror output to chat.

1. Create a bot with [@BotFather](https://t.me/BotFather) and get a token.
2. Place credentials in **`.easyrecon_config.json`** in the project directory:

```json
{
  "bot_token": "YOUR_BOT_TOKEN",
  "chat_id": 123456789
}
```

3. Run as usual. Use **`-no-tg`** to skip Telegram and use stdin only.

---

## ⚙️ Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-d` | Target domain | required |
| `-t` | Concurrency for URL collection and Arjun | `10` |
| `-no-tg` | Disable Telegram; terminal only | off |

---

## 🖥️ Example Output

```
  [*] Target  : example.com
  [*] Threads : 10

  [*] Running 2 tools concurrently...

  [✓] Subfinder    found: 120
  [✓] Assetfinder  found: 45

  [+] Total unique subdomains : 142

  [*] Running httpx...
  [✓] Alive domains : 38
```

---

## 🤝 Contributing

Pull requests are welcome. To add a new pipeline stage, implement the `Tool` interface in `runner/` and wire it from `main` or the relevant `*.go` file:

```go
type Tool interface {
    Name() string
    Run(input []string) ([]string, error)
}
```

---

## 📄 License

MIT — use it, break it, improve it.

---

<div align="center">

<img src="https://readme-typing-svg.demolab.com?font=Fira+Code&size=14&pause=2000&color=00FF41&center=true&vCenter=true&width=400&lines=Happy+Hunting+%F0%9F%8E%AF;Stay+in+scope.+Always." alt="footer" />

</div>
