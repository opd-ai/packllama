# Installation Guide

## Prerequisites

- Go 1.24 or later
- `.gguf` model files (downloadable from HuggingFace)

---

## Linux

### From binary release

```bash
# Download the latest release for your architecture
curl -LO https://github.com/opd-ai/packllama/releases/latest/download/packllama-linux-amd64

chmod +x packllama-linux-amd64
sudo mv packllama-linux-amd64 /usr/local/bin/packllama

# Verify
packllama --help
```

### As a systemd service

```bash
# Create a dedicated user
sudo useradd -r -s /bin/false packllama
sudo mkdir -p /var/lib/packllama/models
sudo chown packllama:packllama /var/lib/packllama

# Copy the binary
sudo cp packllama-linux-amd64 /usr/local/bin/packllama

# Install the service unit
sudo cp packllama.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now packllama

# Check status
sudo systemctl status packllama
sudo journalctl -u packllama -f
```

---

## macOS

### From binary release

```bash
# Apple Silicon
curl -LO https://github.com/opd-ai/packllama/releases/latest/download/packllama-darwin-arm64
chmod +x packllama-darwin-arm64
mv packllama-darwin-arm64 /usr/local/bin/packllama
```

### Build from source

```bash
git clone https://github.com/opd-ai/packllama.git
cd packllama
make build
./bin/packllama --models-dir ~/models
```

---

## Windows

```powershell
# Download the binary
Invoke-WebRequest -Uri `
  https://github.com/opd-ai/packllama/releases/latest/download/packllama-windows-amd64.exe `
  -OutFile packllama.exe

# Run
.\packllama.exe --host 127.0.0.1 --port 8080 --models-dir C:\models
```

---

## Docker

```bash
# Pull and run
docker run -d \
  -p 8080:8080 \
  -v /path/to/your/models:/models \
  -e PACKLLAMA_MODELS_DIR=/models \
  ghcr.io/opd-ai/packllama:latest

# Or use docker-compose
docker compose up -d
```

---

## Go library

```bash
go get github.com/opd-ai/packllama
```

---

## Quick start

After starting the server, test it:

```bash
curl http://127.0.0.1:8080/health
# → {"status":"ok"}

curl http://127.0.0.1:8080/v1/models
# → list of discovered .gguf models
```

See [api-examples.md](api-examples.md) for complete API usage.

---

## Configuration

All configuration can be supplied via:
1. JSON file (`--config path/to/config.json`)
2. Environment variables (`PACKLLAMA_*`)
3. CLI flags (highest priority)

Key options:

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--host` | `PACKLLAMA_HOST` | `127.0.0.1` | Bind address |
| `--port` | `PACKLLAMA_PORT` | `8080` | Listen port |
| `--models-dir` | `PACKLLAMA_MODELS_DIR` | _(empty)_ | Directory of `.gguf` files |
| `--default-model` | `PACKLLAMA_DEFAULT_MODEL` | _(empty)_ | Model loaded on startup |
| `--download-models` | `PACKLLAMA_MODEL_DOWNLOADS` | _(empty)_ | Comma-separated Hugging Face refs (`owner/repo/path.gguf`) auto-downloaded into `models_dir` |
| `--log-level` | `PACKLLAMA_LOG_LEVEL` | `info` | `debug/info/warn/error` |
| `--log-format` | `PACKLLAMA_LOG_FORMAT` | `text` | `text` or `json` |
| `--no-ui` | `PACKLLAMA_DISABLE_UI` | `false` | API-only mode |
| _(file only)_ | `PACKLLAMA_ENABLE_METRICS` | `false` | Prometheus `/metrics` |
