# 🧳🐳 DockPorter 
**by DevLouix**

**A zero-dependency, peer-to-peer Docker migration tool and web control panel.**

DockPorter allows you to seamlessly teleport live Docker containers (including their volumes, environment variables, and port bindings) directly from one server to another—without using a Docker Registry. 

It comes packed as a **single, standalone binary** with an embedded React Web UI. Just drop it on your server, run it, and start managing your fleet.

![DockPorter UI](https://via.placeholder.com/1000x500.png?text=Replace+with+a+screenshot+of+your+beautiful+React+UI)

## ✨ Features

- **Peer-to-Peer Migration**: Stream containers and their volume data directly between servers over the network. Zero intermediate tarballs. Zero disk I/O bottlenecks.
- **Docker Explorer (Control Panel)**: A full-featured web dashboard to Start, Stop, Delete, Rename, and Manage containers with a modern, responsive interface.
- **Single Binary Architecture**: The Go backend and React frontend are compiled into one executable. No Node.js, NPM, or Go installation required on the host.
- **Batch Operations**: Multi-select containers in the UI or CLI and migrate, stop, or delete them all concurrently.
- **Magic Link Startup**: Get up and running instantly with auto-populating "Magic Links" that handle authentication for you.
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
```

---

## 💻 Quick Start (Server & Web UI)

To manage containers and receive incoming migrations, start DockPorter in `server` mode.

```bash
./dockporter-linux-amd64 -mode server -port 8080
```

### 1. The Magic Link
When the server starts, it will print a **Magic Link** to the terminal.
```text
🚀 DockPorter by DevLouix is Online!
🌍 Local Control Panel: http://localhost:8080?token=a1b2c3d4e5f6...
```
Simply `Cmd+Click` or `Ctrl+Click` that link. The UI will open and **automatically authenticate** you.

### 2. Manual Login
If you are accessing the UI from a different machine, run the following to see your token:
```bash
./dockporter-linux-amd64 -show-key
```
Paste that key into the Auth Token field in the top navigation bar.

---

## 🚚 Headless Migration (CLI Mode)

If you don't want to use the Web UI, you can migrate containers purely from the terminal. 

**Scenario:** You want to move `my-database` and `my-web-app` from Server A to Server B.

1. Ensure DockPorter is running in `server` mode on **Server B** (Target).
2. Get the Auth Token from **Server B**.
3. On **Server A** (Source), run:

```bash
./dockporter-linux-amd64 -mode ship \
  -id my-database,my-web-app \
  -to 192.168.1.50:8080 \
  -token <SERVER_B_AUTH_TOKEN>
```

---

## 🛠️ Building from Source

**Prerequisites:** Go 1.21+ and Node.js 18+

1. **Build the React UI:**
   ```bash
   cd ui
   npm install
   npm run build
   cd ..
   ```
2. **Compile the Go Binary:**
   ```bash
   go build -o dockporter ./cmd/agent/main.go
   ```

To compile for all platforms at once, run the included script: `./build_all.sh`.

---

## 🔒 Security Architecture

- **CORS & Preflight Guards**: Allows secure cross-origin requests between different nodes in your network.
- **Smart Protocol Switching**: Automatically detects and upgrades to `HTTPS` and `WSS` when running behind secure proxies or cloud tunnels.
- **X-Auth-Token**: Every action requires a 256-bit hex token passed via headers or secure query parameters.
- **Tar-Slip Protection**: Prevents path traversal attacks during volume extraction.
- **Context-Aware Connections**: Instantly kills orphaned network streams if a migration is cancelled.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
**DockPorter** is built and maintained by **DevLouix**. 🧳⚡