package cloudflare

import (
	"bytes"
	"fmt"
	"gdrsapi/pkg/config"
	"io"
	"log"
	"net/http"
	"time"
)

const baseUrl string = "https://api.cloudflare.com/client/v4/accounts/"

type CFService struct {
	httpClient *http.Client
	cfg        *config.Config
}

func NewCFService() *CFService {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	return &CFService{
		httpClient: client,
		cfg:        cfg,
	}
}

// This will return the response as bytes
func (cf *CFService) CallImgToTextApi(payload []byte) ([]byte, error) {
	var bearer string = "Bearer " + cf.cfg.CloudflareApiKey
	var url string = fmt.Sprintf("%s%s/ai/run/@cf/llava-hf/llava-1.5-7b-hf", baseUrl, cf.cfg.CloudflareAccountId)

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := cf.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to call img to text api: %w", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		log.Printf("Response Status: %d", resp.StatusCode)
		log.Printf("Response Body: %s", string(bodyBytes))
		return nil, fmt.Errorf("too many requests sent. Rate limited by Cloudflare")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}
