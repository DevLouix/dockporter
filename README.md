# 🧳🐳 DockPorter 
**by DevLouix**

**A zero-dependency, peer-to-peer Docker migration tool and web control panel.**

DockPorter allows you to seamlessly teleport live Docker containers (including their volumes, environment variables, and port bindings) directly from one server to another—without using a Docker Registry. 

It comes packed as a **single, standalone binary** with an embedded React Web UI. Just drop it on your server, run it, and start shifting.

![DockPorter UI](https://via.placeholder.com/1000x500.png?text=Replace+with+a+screenshot+of+your+beautiful+React+UI)

## ✨ Features

- **Peer-to-Peer Migration**: Stream containers and their volume data directly between servers over the network. Zero intermediate tarballs. Zero disk I/O bottlenecks.
- **Single Binary**: The Go backend and React frontend are compiled into one executable. No Node.js, NPM, or Go installation required on the host.
- **Real-Time Web UI**: A beautiful dashboard to Start, Stop, Delete, Rename, and Migrate containers with live WebSocket progress bars.
- **Batch Operations**: Multi-select containers in the UI or CLI and migrate them all concurrently.
- **Enterprise Security**: Built-in Preflight CORS protection, strict `X-Auth-Token` authentication, and TLS support.
- **Headless CLI Mode**: Fully scriptable via terminal for CI/CD pipelines or headless servers.

---

## 🚀 Installation

You don't need to install any dependencies (other than Docker). 

1. Go to the [Releases Page](../../releases) of this repository.
2. Download the binary for your operating system (Linux, macOS, or Windows).
3. Extract and make it executable (Linux/macOS):

```bash
chmod +x dockporter-linux-amd64
💻 Quick Start (Server & Web UI)
To manage containers and receive incoming migrations, start DockPorter in server mode.
code
Bash
./dockporter-linux-amd64 -mode server -port 8080
1. Get your Auth Token
Security is enabled by default. To log into the Web UI or send containers to this server, you need its secret Auth Token. Open a new terminal and run:
code
Bash
./dockporter-linux-amd64 -show-key

# Output: 
# 🔑 YOUR AGENT AUTH TOKEN: a1b2c3d4e5f6...
2. Open the Dashboard
Open your web browser and navigate to:
👉 http://localhost:8080 (or your server's IP).
Paste your Auth Token into the top navigation bar to unlock the Control Panel.
🚚 Headless Migration (CLI Mode)
If you don't want to use the Web UI, you can migrate containers purely from the terminal.
Scenario: You want to move my-database and my-web-app from Server A to Server B.
Ensure DockPorter is running in server mode on Server B (Target).
Get the Auth Token from Server B.
On Server A (Source), run:
code
Bash
./dockporter-linux-amd64 -mode ship \
  -id my-database,my-web-app \
  -to 192.168.1.50:8080 \
  -token <SERVER_B_AUTH_TOKEN>
The CLI will display real-time progress bars as the containers stream across the network.
🛠️ Building from Source
If you want to contribute or build the binary yourself:
Prerequisites: Go 1.21+ and Node.js 18+
Build the React UI:
code
Bash
cd ui
npm install
npm run build
cd ..
Compile the Go Binary (which embeds the UI):
code
Bash
go build -o dockporter ./cmd/agent/main.go
To compile for all platforms at once, run the included script: ./build_all.sh.
🔒 Security Architecture
DockPorter is designed to safely operate on internal networks and behind cloud proxies:
X-Auth-Token: Every HTTP request and WebSocket connection must pass a 64-character hex token.
Tar-Slip Protection: The volume extraction engine verifies file paths before writing to disk, preventing malicious archives from overwriting host OS files.
Context-Aware Connections: If a user cancels a migration midway, the TCP stream is instantly killed, preventing memory leaks and orphaned files.
Smart Protocol Resolution: The UI automatically upgrades to https:// and wss:// when deployed behind secure cloud tunnels (like GitHub Codespaces or Cloudflare Tunnels).
📄 License
This project is licensed under the MIT License - see the LICENSE file for details.
code
Code
### Tips for your Repo:
1. Replace the `https://via.placeholder.com/...` link at the top with an actual screenshot of your React app. Save an image into your repo (e.g., `docs/screenshot.png`) and link it: `![UI](docs/screenshot.png)`.
2. When you upload your binaries to GitHub, go to the "Releases" tab on the right side of the repo, click "Draft a new release", and drag/drop the files from your `bin/` folder into the attachments box.