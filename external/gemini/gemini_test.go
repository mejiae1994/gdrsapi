package gemini

import "testing"

func TestGemini(t *testing.T) {
	simpleCfg := map[string]interface{}{
		"temperature":        0.8,
		"response_mime_type": "application/json",
	}

	g := NewGeminiService(simpleCfg)
	prompt := "Write a short story about a cat named Fluffy"
	response, err := g.CallGeminiLLMApi(prompt)

	t.Log(string(response))
	t.Log(err)
}
