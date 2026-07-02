package inference

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeBackend struct {
	gpu       bool
	model     any
	loadErr   error
	generate  func(context.Context, any, string, Parameters, TokenCallback) error
	freedWith []any
}

func (f *fakeBackend) Init() error { return nil }
func (f *fakeBackend) Close()      {}
func (f *fakeBackend) SupportsGPUOffload() bool {
	return f.gpu
}
func (f *fakeBackend) LoadModel(_ string) (any, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	if f.model == nil {
		f.model = "model"
	}
	return f.model, nil
}
func (f *fakeBackend) FreeModel(model any) {
	f.freedWith = append(f.freedWith, model)
}
func (f *fakeBackend) Generate(ctx context.Context, model any, prompt string, p Parameters, cb TokenCallback) error {
	if f.generate != nil {
		return f.generate(ctx, model, prompt, p, cb)
	}
	return nil
}

func TestEngineLoadAndUnloadModelState(t *testing.T) {
	backend := &fakeBackend{gpu: true, model: "m1"}
	engine := NewEngine(backend)

	if err := engine.LoadModel("/models/m1.gguf"); err != nil {
		t.Fatalf("LoadModel() error = %v", err)
	}
	state := engine.State()
	if !state.Loaded || state.Path != "/models/m1.gguf" || !state.GPUAcceleration {
		t.Fatalf("unexpected state after load: %+v", state)
	}

	engine.UnloadModel()
	if got := engine.State(); got.Loaded || got.Path != "" {
		t.Fatalf("unexpected state after unload: %+v", got)
	}
	if len(backend.freedWith) != 1 || backend.freedWith[0] != "m1" {
		t.Fatalf("expected model to be freed once, got %#v", backend.freedWith)
	}
}

func TestLoadModelWrapsErrors(t *testing.T) {
	backend := &fakeBackend{loadErr: errors.New("boom")}
	engine := NewEngine(backend)

	err := engine.LoadModel("/models/m1.gguf")
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "load model \"/models/m1.gguf\": boom" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestPipelineRunUsesCallbackStreaming(t *testing.T) {
	backend := &fakeBackend{
		model: "m1",
		generate: func(_ context.Context, _ any, _ string, _ Parameters, cb TokenCallback) error {
			if err := cb(TokenEvent{Token: "hello"}); err != nil {
				return err
			}
			return cb(TokenEvent{Token: " world", FinishReason: "stop"})
		},
	}
	engine := NewEngine(backend)
	if err := engine.LoadModel("/models/m1.gguf"); err != nil {
		t.Fatalf("LoadModel() error = %v", err)
	}

	var got []TokenEvent
	pipeline := engine.NewPipeline(Parameters{MaxTokens: 8}, func(ev TokenEvent) error {
		got = append(got, ev)
		return nil
	})

	if err := pipeline.Run(context.Background(), "hi"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := []TokenEvent{{Token: "hello"}, {Token: " world", FinishReason: "stop"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("callback events mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestPipelineRunRequiresLoadedModel(t *testing.T) {
	engine := NewEngine(&fakeBackend{})
	pipeline := engine.NewPipeline(Parameters{}, nil)

	err := pipeline.Run(context.Background(), "prompt")
	if !errors.Is(err, ErrModelNotLoaded) {
		t.Fatalf("expected ErrModelNotLoaded, got %v", err)
	}
}
