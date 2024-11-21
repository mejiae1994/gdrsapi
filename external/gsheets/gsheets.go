package gsheets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gdrsapi/pkg/config"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func NewSheetsService() *SheetsApp {
	ctx := context.Background()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Unable to get config: %v", err)
	}

	credBytes, err := base64.StdEncoding.DecodeString(cfg.GoogleSACred)
	if err != nil {
		log.Fatalf("Unable to decode base64 credentials: %v", err)
	}

	//validate json
	var js map[string]interface{}
	if err := json.Unmarshal(credBytes, &js); err != nil {
		log.Fatalf("Invalid JSON in decoded credentials: %v", err)
	}

	fmt.Println("loaded google service account credentials")

	sheetsService, err := sheets.NewService(ctx,
		option.WithCredentialsJSON(credBytes),
		option.WithScopes(sheets.SpreadsheetsScope),
	)
	if err != nil {
		log.Fatalf("Unable to create sheets service: %v", err)
	}

	fmt.Println("Sheets service created")
	return &SheetsApp{
		sheetSvc: sheetsService,
	}
}

func (sApp *SheetsApp) InsertSteamRatingEntry(se SheetsEntry) error {
	sheetsId := "1SHupRSsjmSuDFuAiYtrfHlpg0n0LgLKoXys1QVXGkdo"
	sheetsRange := "Sheet1!A:G"

	vr := &sheets.ValueRange{}
	objectList := []interface{}{se.Title, se.AppId, se.Url, se.PromptType, se.Score, se.Rating, se.Prompt}
	vr.Values = [][]interface{}{objectList}

	err := sApp.InsertSheetsRow(sheetsId, sheetsRange, vr)
	if err != nil {
		return err
	}

	return nil
}

func (sApp *SheetsApp) InsertSheetsRow(sheetId string, sheetsRange string, vr *sheets.ValueRange) error {
	_, err := sApp.sheetSvc.Spreadsheets.Values.Append(sheetId, sheetsRange, vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}
	return nil
}

func loadCredentials(path string) Credentials {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	var credentials Credentials
	if err := json.Unmarshal(fileBytes, &credentials); err != nil {
		log.Fatalf("Unable to parse credentials file: %v", err)
	}

	return credentials
}

type SheetsApp struct {
	sheetSvc *sheets.Service
}

type Credentials struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

type SheetsEntry struct {
	Title      string
	AppId      string
	Url        string
	PromptType string
	Score      string
	Rating     string
	Prompt     string
}
