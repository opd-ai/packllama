# packllama Plan: OpenAI-Compatible LLM Service with Ebitengine UI

## Project Overview

**packllama** is a local LLM inference service built on the gollama.cpp library that provides an **OpenAI-compatible REST API**, a **rich Ebitengine-powered desktop interface**, and an **embeddable server package** for integration into other Go applications. It is optimized for coding assistance while supporting standard self-hosted LLM inference tasks.

This project leverages gollama.cpp's pure Go bindings (no CGO), automatic library management, and cross-platform GPU support to deliver a minimal, fast, and simple deployment experience.

### Key Objectives

- Provide OpenAI API compatibility with feature parity to Ollama/llama.cpp server
- Build an intuitive, responsive Ebitengine-based desktop UI
- Enable embedding the server into any Go application
- Optimize for coding assistant workflows
- Maintain simple, dependency-light architecture

---

## Phase 1: Foundation & API Server

### 1.1 Project Structure

- [x] Create new repository with clear module organization
  - `./cmd/packllama` - CLI application entry point
  - `./internal/api` - HTTP API handlers
  - `./internal/service` - Business logic layer
  - `./internal/ui` - Ebitengine UI components (package, not full service)
  - `./pkg/server` - Embeddable server library (public API)
  - `./examples` - Example applications
  - `./docs` - API documentation and guides
- [ ] Set up Go module with gollama.cpp dependency
- [x] Create Makefile with build targets (build, test, run, release)
- [x] Set up GitHub Actions CI/CD (test, cross-platform build)

### 1.2 HTTP Server Foundation

- [x] Create HTTP server wrapper around `net/http` with middleware stack
- [x] Implement request logging and error handling middleware
- [x] Add request ID tracking for tracing
- [x] Implement graceful shutdown with context cancellation
- [x] Set up structured logging (`slog`) with appropriate levels
- [x] Create health check endpoint (`/health`)
- [x] Implement CORS support (configurable)

### 1.3 gollama.cpp Integration Layer

- [ ] Create wrapper package around gollama.cpp library functions
- [ ] Implement model loading with error handling
- [ ] Create inference pipeline abstraction
- [ ] Implement token streaming callback mechanism
- [ ] Add model state management (loaded/unloaded)
- [ ] Create configuration for inference parameters
- [ ] Handle GPU acceleration detection and fallback

---

## Phase 2: OpenAI API Compatibility

### 2.1 Chat Completions Endpoint

- [x] Implement `/v1/chat/completions` POST endpoint
- [x] Support OpenAI request schema (messages array with roles)
- [x] Implement streaming responses with Server-Sent Events (SSE)
- [x] Support parameters: `model`, `temperature`, `top_p`, `max_tokens`, `stop`, `stream`
- [ ] Map OpenAI parameters to gollama.cpp inference settings
- [x] Implement response schema matching OpenAI format (including `finish_reason`)
- [x] Add request validation with helpful error messages
- [x] Write comprehensive endpoint tests

### 2.2 Text Completions Endpoint

- [x] Implement `/v1/completions` POST endpoint
- [x] Support legacy completion format (string prompt)
- [x] Implement streaming for completions
- [x] Support completion-specific parameters (`prompt`, `suffix`, `max_tokens`)
- [ ] Create compatibility tests with OpenAI SDK

### 2.3 Model Management Endpoints

- [x] Implement `/v1/models` GET (list available models)
- [x] Implement `/v1/models/{model_id}` GET (model info)
- [ ] Add model metadata (context length, parameters, quantization)
- [ ] Implement `POST /v1/models` to load new models
- [ ] Implement `DELETE /v1/models/{model_id}` to unload models
- [ ] Add model discovery from configurable directories

### 2.4 Embeddings Support

- [x] Implement `/v1/embeddings` POST endpoint
- [x] Support single and batch embedding requests
- [x] Return embeddings in OpenAI format
- [ ] Support dimension reduction options if applicable

### 2.5 API Testing

- [ ] Create test suite for all endpoints
- [ ] Test with OpenAI Python SDK (`openai` library)
- [ ] Test with other LLM frameworks (langchain, llama-index)
- [ ] Validate error handling and edge cases
- [ ] Create API usage examples and curl commands

---

## Phase 3: Configuration & Deployment

### 3.1 Configuration System

- [x] Create JSON configuration file support
- [x] Environment variable overrides
- [x] CLI flags for common settings
- [x] Configuration validation on startup
- [x] Default configuration with sane defaults
- [ ] Per-model parameter overrides

### 3.2 Model Discovery & Management

- [x] Support loading from configurable model directories
- [x] Implement model file detection (`*.gguf`)
- [x] Create model registry with metadata caching
- [x] Support model aliases (for example, `"default"` → actual model path)
- [ ] Auto-download models from Hugging Face (optional)
- [ ] Model preload on startup option

### 3.3 Logging & Observability

- [x] Structured logging with context
- [x] Debug logging mode flag
- [ ] Request/response logging options (can be expensive)
- [ ] Performance metrics logging (tokens/sec, latency)
- [ ] Error and warning aggregation
- [ ] Prometheus metrics export endpoint (optional Phase 3.5+)

### 3.4 Deployment Artifacts

- [ ] Cross-platform binary builds (Windows, macOS, Linux)
- [ ] Docker image with GPU support documentation
- [ ] Systemd service file for Linux deployments
- [ ] Example `docker-compose.yml`
- [ ] Installation documentation for each platform

---

## Phase 4: Ebitengine-Based Desktop Interface

### 4.1 UI Framework Setup

- [ ] Initialize Ebitengine game loop and window
- [ ] Create custom GUI widget library
  - Button with hover/click states
  - Text input field
  - Text display area with scrolling
  - Dropdown/select
  - Slider for numeric parameters
  - Checkbox for toggles
- [ ] Implement theme system (colors, fonts, spacing)
- [ ] Create layout manager for responsive positioning
- [ ] Implement keyboard navigation and focus system

### 4.2 Chat Interface

- [ ] Build message display area with scrollable history
- [ ] Implement syntax highlighting for code blocks (using existing library or custom)
- [ ] Create message input field with multi-line support
- [ ] Display streaming responses in real time
- [ ] Show token count and generation speed
- [ ] Implement message editing/deletion UI
- [ ] Add conversation management (new/load/save/delete)
- [ ] Create conversation browser/selector panel

### 4.3 Model & Parameter Controls

- [ ] Model selection dropdown with live model list
- [ ] Parameter adjustment UI (temperature, top_p, context, etc.)
- [ ] Preset buttons (Creative, Balanced, Precise)
- [ ] Advanced parameters toggle (collapsed by default)
- [ ] Real-time parameter validation
- [ ] Parameter saving/loading

### 4.4 System Information Display

- [ ] GPU/CPU usage visualization
- [ ] Memory utilization gauge
- [ ] Inference speed indicator (tokens/sec)
- [ ] Model load status
- [ ] Queue length display
- [ ] System health indicator

### 4.5 Code Assistant Features

- [ ] Enhanced code block display with copy button
- [ ] Language-aware syntax highlighting
- [ ] Code diff viewer for before/after
- [ ] Quick action buttons (Explain, Generate Tests, Refactor)
- [ ] File context browser
- [ ] Terminal output display

### 4.6 Settings Panel

- [ ] UI theme selector (dark/light)
- [ ] Font size adjustment
- [ ] Default model selection
- [ ] Auto-save preferences
- [ ] Clear conversation history option
- [ ] About/info panel

---

## Phase 5: Embeddable Server Library

### 5.1 Public `pkg/server` Package

- [ ] Design clean public API for embedding the packllama server
- [ ] Create `Server` struct with exported constructor `New(opts ...Option)`
- [ ] Implement functional options pattern for configuration
- [ ] Public methods: `Start()`, `Stop()`, `ListenAddr()`, `LoadModel()`, `UnloadModel()`
- [ ] Webhook/callback hooks for lifecycle events
- [ ] Thread-safe inference queue management

### 5.2 Embeddable Web UI

- [ ] Create minimal web UI (HTML/CSS/JS)
- [ ] Embed UI assets in binary using `//go:embed`
- [ ] Serve embedded UI from configurable path (`/ui`, etc.)
- [ ] WebSocket connection to server for real-time updates
- [ ] Option to disable UI for API-only deployments
- [ ] Mobile-responsive design

### 5.3 Configuration Options for Embedding

- [ ] `WithPort(port int)` - Server port
- [ ] `WithHost(host string)` - Bind address
- [ ] `WithModelsDir(path string)` - Models directory
- [ ] `WithDefaultModel(path string)` - Initial model to load
- [ ] `WithMaxContextLength(len int)` - Context window size
- [ ] `WithDisableUI(bool)` - Run API-only mode
- [ ] `WithCustomTheme(theme string)` - Theme configuration
- [ ] `WithCORSAllowed(origins []string)` - CORS origins

### 5.4 Example Applications

- [ ] Simple chat application using the embedded server
- [ ] Integration with existing Go web application
- [ ] Microservice example with server as library
- [ ] CLI tool using embedded inference
- [ ] Documentation for each example

### 5.5 Integration Testing

- [ ] Test embedding in example applications
- [ ] Verify API works from client library perspective
- [ ] Test lifecycle (start → load model → inference → stop)
- [ ] Memory leak detection

---

## Phase 6: Coding Assistant Optimization

### 6.1 Code-Specific UI Components

- [ ] Language selector for syntax highlighting (detect or manual)
- [ ] Multi-file context manager
- [ ] Code diff viewer component
- [ ] Terminal/output panel for execution results
- [ ] Variable/symbol lookup widget
- [ ] Template system for quick prompts

### 6.2 Specialized Prompts

- [ ] Code explanation template
- [ ] Unit test generation
- [ ] Documentation generation
- [ ] Refactoring suggestions
- [ ] Bug finding and fixes
- [ ] Performance optimization
- [ ] Custom prompt creation

### 6.3 Coding Workflow Features

- [ ] Paste code → get explanation workflow
- [ ] Generate tests for code snippet
- [ ] Side-by-side diff viewer
- [ ] Context file/folder selector
- [ ] Quick generate buttons in UI
- [ ] Export generated code with formatting

---

## Phase 7: Quality, Testing & Documentation

### 7.1 Testing

- [ ] Unit tests for service layer (>80% coverage)
- [ ] Integration tests for API endpoints
- [ ] End-to-end tests using the `openai` SDK
- [ ] Stress tests (many concurrent requests, long conversations)
- [ ] GPU memory tests (no leaks after many inferences)
- [ ] Cross-platform build verification
- [ ] UI rendering tests (headless mode)

### 7.2 Documentation

- [ ] OpenAPI/Swagger specification for REST API
- [ ] Quick start guide (5 minutes to first inference)
- [ ] Installation guide for each platform
- [ ] Configuration reference (all options)
- [ ] Model selection and optimization guide
- [ ] UI user guide with screenshots
- [ ] Code assistant feature guide
- [ ] Embedding integration guide with code examples
- [ ] API compatibility matrix (vs OpenAI spec)
- [ ] Troubleshooting and FAQ
- [ ] Architecture overview document

### 7.3 Performance & Monitoring

- [ ] Benchmark suite for inference speed
- [ ] Memory profiling and optimization
- [ ] GPU utilization monitoring
- [ ] Latency measurements
- [ ] Throughput testing
- [ ] Resource usage documentation

---

## Phase 8: Release & Distribution

### 8.1 Release Artifacts

- [ ] Binary releases for:
  - Windows (amd64, arm64)
  - macOS (amd64, arm64)
  - Linux (amd64, arm64)
  - Linux ARM (Raspberry Pi compatibility)
- [ ] Docker image with CUDA/ROCm variants
- [ ] Homebrew formula for macOS
- [ ] Portable ZIP archives
- [ ] Installation scripts

### 8.2 Installation Methods

- [ ] Direct binary downloads
- [ ] Docker: `docker run ghcr.io/opd-ai/packllama`
- [ ] Systemd service for Linux
- [ ] Windows installer/portable
- [ ] macOS DMG or Homebrew
- [ ] Go library: `go get github.com/opd-ai/packllama`

### 8.3 Release Management

- [ ] Semantic versioning (MAJOR.MINOR.PATCH)
- [ ] `CHANGELOG.md` maintenance
- [ ] Release notes with highlighted features
- [ ] Breaking change policy
- [ ] Deprecation timeline
- [ ] Security advisory process

### 8.4 Community & Support

- [ ] GitHub issue templates (bug report, feature request)
- [ ] Discussion forum or Discord channel
- [ ] Contributing guidelines
- [ ] Code of conduct
- [ ] Community model collection/sharing
- [ ] Showcase of applications using packllama

---

## Phase 9: Advanced Features (Post-MVP)

### 9.1 Extended Model Support

- [ ] Model hot-loading (switch without restart)
- [ ] Multiple simultaneous models (if memory permits)
- [ ] Model scheduling/queuing for resource limits
- [ ] Model warming/preload
- [ ] Vision models (image input)
- [ ] Audio models (speech input/output)

### 9.2 Advanced Inference

- [ ] Structured output generation (JSON schema guidance)
- [ ] Function calling / tool use
- [ ] Guided generation
- [ ] Batch inference API
- [ ] Retrieval-augmented generation (RAG) helpers
- [ ] Few-shot example management

### 9.3 Monitoring & Analytics

- [ ] Prometheus metrics export
- [ ] Usage statistics dashboard
- [ ] Performance benchmarking
- [ ] Model performance analytics
- [ ] User activity logging (optional/privacy-respecting)
- [ ] Cost tracking for resource optimization

### 9.4 Security Features

- [ ] API key authentication
- [ ] Per-key rate limiting
- [ ] Request signing/verification
- [ ] Audit logging
- [ ] TLS/HTTPS support
- [ ] Input sanitization
- [ ] Output filtering/moderation options

### 9.5 IDE Integration

- [ ] VS Code extension for local packllama server
- [ ] Extension marketplace listing
- [ ] Language-specific IDE plugin templates
- [ ] Editor context passing protocol
- [ ] Keyboard shortcut configuration

---

## Technical Architecture

### Module Organization

```text
.
├── cmd/packllama
├── docs
├── examples
├── internal
│   ├── api
│   ├── service
│   └── ui
└── pkg
    └── server
```

### Technology Stack

| Component | Technology | Rationale |
| --- | --- | --- |
| Core Library | gollama.cpp (pure Go) | No CGO, automatic library management, cross-platform |
| HTTP Server | `net/http` | Standard library, lightweight |
| Desktop UI | Ebitengine | Native performance, rich graphics, cross-platform |
| Config | Go `flag` + YAML/TOML | Simple, standard approach |
| Logging | `slog` | Go 1.21+ standard, structured logging |
| Testing | Go testing + testify | Standard library + assertion helpers |
| Build | Makefile | Portable, simple |

### Key Design Principles

- **Simple**: No unnecessary dependencies beyond gollama.cpp
- **Pure Go**: No CGO, no compilation required for users
- **OpenAI Compatible**: Maximize ecosystem integration
- **Streaming**: Server-Sent Events for real-time responses
- **Embeddable**: `pkg/server` is production-ready for integration
- **UI is Optional**: Can run as API-only or with the Ebitengine UI
- **Cross-Platform**: Single binary works everywhere

---

## Success Criteria

- [ ] `/v1/chat/completions` endpoint fully compatible with the OpenAI SDK
- [ ] `/v1/models` endpoint returns accurate model metadata
- [ ] Streaming responses work with Server-Sent Events
- [ ] Ebitengine UI renders smoothly at 60 FPS
- [ ] Chat interface handles 100+ messages in a single conversation without degraded usability
- [ ] Embedded server API is stable enough for integration into external Go applications
