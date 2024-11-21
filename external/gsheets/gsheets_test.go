package gsheets

import (
	"testing"
)

func TestInitSheetsApp(t *testing.T) {
	const credPath = "../config/gdreststop-cred.json"
	t.Log("Testing InitSheetsApp and inserting a row")
	svc := NewSheetsService()

	se := SheetsEntry{
		Title:      "test",
		AppId:      "test",
		Url:        "test",
		PromptType: "test",
		Score:      "test",
		Rating:     "test",
		Prompt:     "test",
	}

	err := svc.InsertSteamRatingEntry(se)
	if err != nil {
		t.Errorf("Error inserting entry: %v", err)
	}
}
