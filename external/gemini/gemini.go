package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gdrsapi/pkg/config"
	"io"
	"log"
	"net/http"
	"time"
)

type GeminiService struct {
	httpClient *http.Client
	cfg        *config.Config
	genConfig  map[string]interface{}
}

func NewGeminiService(genCfg map[string]interface{}) *GeminiService {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	return &GeminiService{
		httpClient: client,
		cfg:        cfg,
		genConfig:  genCfg,
	}
}

func (g *GeminiService) CallGeminiLLMApi(propmt string) ([]byte, error) {
	var url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + g.cfg.GeminiApiKey

	type TextPart struct {
		Text string `json:"text"`
	}

	type Content struct {
		Parts []TextPart `json:"parts"`
	}

	type Candidate struct {
		Content Content `json:"content"`
	}

	type GeminiResponse struct {
		Candidates []Candidate `json:"candidates"`
	}

	inputData := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": propmt,
					},
				},
			},
		},
		"generationConfig": g.genConfig,
	}

	jsonInput, err := json.Marshal(inputData)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonInput))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	log.Println("sent gemini request")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		log.Printf("Gemini LLM: Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	log.Println("received gemini response")
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		log.Printf("Response Status: %d", resp.StatusCode)
		log.Printf("Response Body: %s", string(bodyBytes))
		return nil, fmt.Errorf("too many requests sent. Rate limited by Cloudflare")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var response GeminiResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	responseBody := response.Candidates[0].Content.Parts[0].Text
	return []byte(responseBody), nil
}
