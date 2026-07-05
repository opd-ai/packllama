package api

import (
	"context"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/packllama/internal/service"
)

func TestPythonSDK_OpenAIChatCompatibility(t *testing.T) {
	python := requirePythonAndModules(t, "openai")

	svc := &stubInference{
		chatChunks: []service.ChatChunk{
			{Content: "Hello"},
			{Content: " world", FinishReason: "stop"},
		},
	}
	server := httptest.NewServer(newTestHandler(svc))
	t.Cleanup(server.Close)

	code := `
import os
from openai import OpenAI

base_url = os.environ["PACKLLAMA_BASE_URL"] + "/v1"
client = OpenAI(api_key="test-key", base_url=base_url)

resp = client.chat.completions.create(
    model="test-model",
    messages=[{"role": "user", "content": "Say hello"}],
)
assert resp.choices[0].message.content == "Hello world"
assert resp.choices[0].finish_reason == "stop"

chunks = list(client.chat.completions.create(
    model="test-model",
    messages=[{"role": "user", "content": "Stream please"}],
    stream=True,
))
assert len(chunks) > 0
`
	runPythonCode(t, python, code, map[string]string{
		"PACKLLAMA_BASE_URL": server.URL,
	})
}

func TestPythonFrameworks_LangChainAndLlamaIndex(t *testing.T) {
	python := requirePythonAndModules(t, "openai", "langchain_openai", "langchain_core", "llama_index.llms.openai")

	svc := &stubInference{
		chatChunks: []service.ChatChunk{
			{Content: "packllama"},
			{Content: " works", FinishReason: "stop"},
		},
	}
	server := httptest.NewServer(newTestHandler(svc))
	t.Cleanup(server.Close)

	code := `
import os
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
from llama_index.llms.openai import OpenAI as LlamaOpenAI

base_url = os.environ["PACKLLAMA_BASE_URL"] + "/v1"
api_key = "test-key"

lc = ChatOpenAI(
    model="test-model",
    api_key=api_key,
    base_url=base_url,
    temperature=0,
)
lc_resp = lc.invoke([HumanMessage(content="Say hello")])
assert "packllama works" in lc_resp.content

llm = LlamaOpenAI(
    model="test-model",
    api_key=api_key,
    api_base=base_url,
    max_tokens=16,
    temperature=0,
)
li_resp = llm.complete("Say hello")
assert "packllama works" in str(li_resp)
`
	runPythonCode(t, python, code, map[string]string{
		"PACKLLAMA_BASE_URL": server.URL,
	})
}

func requirePythonAndModules(t *testing.T, modules ...string) string {
	t.Helper()
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available")
	}
	if os.Getenv("PACKLLAMA_PYTHON_INTEGRATION") != "1" {
		t.Skip("set PACKLLAMA_PYTHON_INTEGRATION=1 to run python integration tests")
	}
	args := append([]string{"-c", "import " + strings.Join(modules, ", ")}, []string{}...)
	cmd := exec.Command(python, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("python modules unavailable: %s", strings.TrimSpace(string(output)))
	}
	return python
}

func runPythonCode(t *testing.T, python string, code string, extraEnv map[string]string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, python, "-c", code)
	cmd.Env = os.Environ()
	for k, v := range extraEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("python command timed out: %s", strings.TrimSpace(string(output)))
	}
	if err != nil {
		t.Fatalf("python command failed: %v\n%s", err, strings.TrimSpace(string(output)))
	}
}
