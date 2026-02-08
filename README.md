# PingAI

AI API Availability Tester - a desktop tool to check AI API connectivity, chat, streaming, and model list support.

## Features

- Multi-protocol support: OpenAI / Anthropic / Gemini
- 16 built-in providers: OpenAI, Anthropic, Gemini, DeepSeek, Qwen, Doubao, Zhipu, Moonshot, Baichuan, SiliconFlow, 01.AI, Groq, Mistral, OpenRouter, Antigravity Tools, Ollama
- 5 check items: Connectivity, Chat, Streaming, Model List, Multi-turn
- Batch key checking
- Provider management with custom providers
- History records with SQLite storage
- i18n support (English / Chinese)
- Cross-platform: macOS / Windows / Linux

## Tech Stack

- **Backend**: Go + Wails v2
- **Frontend**: Vue 3 + TypeScript + Vite
- **Database**: SQLite (pure Go, modernc.org/sqlite)

## Development

```bash
# Install dependencies
cd frontend && npm install && cd ..

# Run in dev mode
wails dev

# Build for current platform
wails build -clean

# Build for all platforms
make build-all
```

## Build

```bash
# macOS (Universal)
make build-mac

# Windows (AMD64)
make build-windows

# Linux (AMD64)
make build-linux
```

## License

MIT
