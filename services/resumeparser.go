package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"recruitment-system/config"
)

const resumeParserURL = "https://api.apilayer.com/resume_parser/upload"

// APILayerResponse matches the JSON structure from the third-party API
type APILayerResponse struct {
	Education  []map[string]interface{} `json:"education"`
	Email      string                   `json:"email"`
	Experience []map[string]interface{} `json:"experience"`
	Name       string                   `json:"name"`
	Phone      string                   `json:"phone"`
	Skills     []string                 `json:"skills"`
}

func ParseResume(filePath string) (*APILayerResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequest("POST", resumeParserURL, file)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("apikey", config.AppConfig.ResumeParserApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to resume parser: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Resume Parser API Error: %s", string(body))
		return nil, fmt.Errorf("resume parser API returned status %d", resp.StatusCode)
	}

	var apiResponse APILayerResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	return &apiResponse, nil
}

// Helper function to convert complex objects to string for DB storage
func ConvertToJSONString(data interface{}) string {
	if data == nil {
		return ""
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data to string: %v", err)
		return ""
	}
	// Use RawMessage to avoid double escaping
	return string(bytes.Replace(b, []byte(`"`), []byte(`'`), -1))
}
