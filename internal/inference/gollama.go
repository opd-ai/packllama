package inference

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"sync"

	"github.com/dianlight/gollama.cpp"
	"github.com/opd-ai/packllama/internal/service"
)

var ErrModelNotLoaded = errors.New("model not loaded")

const (
	defaultTemperature = 0.8
	defaultTopP        = 0.95
	defaultMaxTokens   = 256
)

// Parameters configures inference sampling behavior.
type Parameters struct {
	Temperature float64
	TopP        float64
	MaxTokens   int
	Stop        []string
	GPULayers   int
}

// MapOpenAIRequest maps OpenAI-style request fields into backend parameters.
func MapOpenAIRequest(req service.InferenceRequest) Parameters {
	params := Parameters{
		Temperature: defaultTemperature,
		TopP:        defaultTopP,
		MaxTokens:   defaultMaxTokens,
		Stop:        slices.Clone(req.Stop),
	}
	if req.Temperature != nil && *req.Temperature >= 0 && *req.Temperature <= 2 {
		params.Temperature = *req.Temperature
	}
	if req.TopP != nil && *req.TopP >= 0 && *req.TopP <= 1 {
		params.TopP = *req.TopP
	}
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		params.MaxTokens = *req.MaxTokens
	}
	return params
}

// TokenEvent is emitted for each streamed token.
type TokenEvent struct {
	Token        string
	FinishReason string
}

// TokenCallback receives streamed token events.
type TokenCallback func(TokenEvent) error

// ModelState tracks currently loaded model information.
type ModelState struct {
	Path            string
	Loaded          bool
	GPUAcceleration bool
}

// Engine wraps gollama.cpp operations behind a testable abstraction.
type Engine struct {
	backend Backend

	mu    sync.RWMutex
	model any
	state ModelState
}

// Pipeline represents one inference operation.
type Pipeline struct {
	engine   *Engine
	params   Parameters
	callback TokenCallback
}

// Backend defines the operations needed from an inference backend.
type Backend interface {
	Init() error
	Close()
	SupportsGPUOffload() bool
	LoadModel(path string) (any, error)
	FreeModel(model any)
	Generate(ctx context.Context, model any, prompt string, params Parameters, callback TokenCallback) error
}

// NewEngine creates an Engine with the given backend.
func NewEngine(backend Backend) *Engine {
	return &Engine{backend: backend}
}

// NewGollamaBackend creates a production backend implementation.
func NewGollamaBackend() Backend {
	return gollamaBackend{}
}

// Init prepares the backend runtime.
func (e *Engine) Init() error {
	if e.backend == nil {
		return errors.New("backend is nil")
	}
	return e.backend.Init()
}

// Close frees backend runtime resources.
func (e *Engine) Close() {
	if e == nil || e.backend == nil {
		return
	}
	e.UnloadModel()
	e.backend.Close()
}

// DetectGPUAcceleration reports whether GPU acceleration is available.
func (e *Engine) DetectGPUAcceleration() bool {
	if e == nil || e.backend == nil {
		return false
	}
	return e.backend.SupportsGPUOffload()
}

// LoadModel loads a model file and updates state.
// The entire load-and-swap sequence is serialized under the write mutex so that
// two concurrent callers cannot free each other's models.
func (e *Engine) LoadModel(path string) error {
	if path == "" {
		return errors.New("model path is required")
	}
	if e.backend == nil {
		return errors.New("backend is nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	model, err := e.backend.LoadModel(path)
	if err != nil {
		return fmt.Errorf("load model %q: %w", path, err)
	}

	if e.model != nil {
		e.backend.FreeModel(e.model)
	}
	e.model = model
	e.state = ModelState{Path: path, Loaded: true, GPUAcceleration: e.backend.SupportsGPUOffload()}
	return nil
}

// UnloadModel frees the current model and clears state.
func (e *Engine) UnloadModel() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.backend != nil && e.model != nil {
		e.backend.FreeModel(e.model)
	}
	e.model = nil
	e.state = ModelState{}
}

// State returns a snapshot of current model state.
func (e *Engine) State() ModelState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// NewPipeline creates an inference pipeline with parameters and token callback.
func (e *Engine) NewPipeline(params Parameters, callback TokenCallback) *Pipeline {
	return &Pipeline{engine: e, params: params, callback: callback}
}

// Run executes one inference request and streams token callbacks.
// The read lock is held for the duration of Generate to prevent UnloadModel
// from freeing the model while generation is in-flight.  A nil callback is
// normalized to a no-op to avoid panics in the backend.
func (p *Pipeline) Run(ctx context.Context, prompt string) error {
	if p == nil || p.engine == nil {
		return errors.New("pipeline is nil")
	}
	if p.engine.backend == nil {
		return errors.New("backend is nil")
	}

	cb := p.callback
	if cb == nil {
		cb = func(TokenEvent) error { return nil }
	}

	p.engine.mu.RLock()
	defer p.engine.mu.RUnlock()
	if p.engine.model == nil {
		return ErrModelNotLoaded
	}

	return p.engine.backend.Generate(ctx, p.engine.model, prompt, p.params, cb)
}

type gollamaBackend struct{}

func (g gollamaBackend) Init() error {
	return gollama.Backend_init()
}

func (g gollamaBackend) Close() {
	gollama.Backend_free()
}

func (g gollamaBackend) SupportsGPUOffload() bool {
	return gollama.Supports_gpu_offload()
}

func (g gollamaBackend) LoadModel(path string) (any, error) {
	params := gollama.Model_default_params()
	model, err := gollama.Model_load_from_file(path, params)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (g gollamaBackend) FreeModel(model any) {
	m, ok := model.(gollama.LlamaModel)
	if !ok {
		return
	}
	gollama.Model_free(m)
}

func (g gollamaBackend) Generate(_ context.Context, _ any, _ string, _ Parameters, _ TokenCallback) error {
	if runtime.GOOS != "darwin" {
		return errors.New("token generation is currently supported only on macOS in gollama.cpp")
	}
	return errors.New("token generation is not yet wired in this backend")
}
