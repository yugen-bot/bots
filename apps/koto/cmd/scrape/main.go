package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"jurien.dev/yugen/shared/utils"
)

const baseURL = "https://api.yourdictionary.com/wordfinder/v1/wordlist?order_by=alpha&dictionary=WL&word_length=6&suggest_links=true&group_by=word_length&has_definition=check&interlink_type=length&special=length"

type wordfinderResponse struct {
	Data struct {
		Meta struct {
			Total int `json:"total"`
		} `json:"_meta"`
		Groups []struct {
			Items []string `json:"_items"`
		} `json:"_groups"`
	} `json:"data"`
}

func fetchPage(client *http.Client, offset int) (*wordfinderResponse, error) {
	url := fmt.Sprintf("%s&offset=%d", baseURL, offset)

	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result wordfinderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}

func scrapeAllWords(client *http.Client) []string {
	var allWords []string

	offset := 0
	retries := 0

	for {
		if offset > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		utils.Logger.Infof("Scraping from offset %d...", offset)

		result, err := fetchPage(client, offset)
		if err != nil {
			retries++
			if retries >= 5 {
				utils.Logger.Fatalf("max retries exceeded: %v", err)
			}

			utils.Logger.Warnf("request failed (retry %d): %v", retries, err)

			continue
		}

		retries = 0

		if len(result.Data.Groups) == 0 {
			utils.Logger.Warn("no groups in response, stopping")

			break
		}

		words := result.Data.Groups[0].Items
		allWords = append(allWords, words...)
		offset += len(words)

		utils.Logger.Infof(
			"Got %d words (total so far: %d / %d)",
			len(words),
			offset,
			result.Data.Meta.Total,
		)

		if offset >= result.Data.Meta.Total {
			break
		}
	}

	return allWords
}

func writeWords(outputPath string, allWords []string) {
	data, err := json.MarshalIndent(allWords, "", "  ")
	if err != nil {
		utils.Logger.Fatalf("marshal words: %v", err)
	}

	if err := os.WriteFile(outputPath, data, 0o600); err != nil {
		utils.Logger.Fatalf("write output: %v", err)
	}

	utils.Logger.Infof("Words written to %s", outputPath)
}

func main() {
	outputPath := flag.String(
		"output",
		"internal/assets/words.json",
		"Output path for words.json",
	)

	flag.Parse()

	_ = godotenv.Load() // .env is optional in production environments

	utils.CreateLogger("koto-scrape")

	defer utils.Logger.Sync()

	utils.Logger.Info("Starting word scrape...")

	client := &http.Client{Timeout: 30 * time.Second}

	allWords := scrapeAllWords(client)

	utils.Logger.Infof("Scrape complete: %d words total", len(allWords))

	writeWords(*outputPath, allWords)
}
