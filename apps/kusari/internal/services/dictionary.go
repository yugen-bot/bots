package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/utils"
)

type DictionaryService struct {
	cfg    *config.Config
	client *http.Client
}

func CreateDictionaryService(cfg *config.Config) *DictionaryService {
	utils.Logger.Info("Creating Dictionary Service")
	return &DictionaryService{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (service *DictionaryService) Check(
	ctx context.Context,
	word string,
) (bool, error) {
	word = strings.ToLower(word)

	replacer := strings.NewReplacer(
		"’", "'",
		"‘", "'",
		"“", `"`,
		"”", `"`,
	)
	word = replacer.Replace(word)
	wiktionaryURL := fmt.Sprintf(
		"https://en.wiktionary.org/w/api.php?action=opensearch&format=json&formatversion=2&search=%s&namespace=0&limit=2",
		url.QueryEscape(word),
	)

	utils.Logger.Debug(wiktionaryURL)

	req, err := http.NewRequestWithContext(ctx, "GET", wiktionaryURL, nil)
	if err != nil {
		return false, fmt.Errorf("dictionary: new request: %w", err)
	}

	req.Header.Set(
		"User-Agent",
		"YugenKusari/1.0 (https://github.com/jurienhamaker/yugen;info@jurien.dev) Go-http-client/1.1",
	)
	req.SetBasicAuth(
		service.cfg.WiktionaryUsername,
		service.cfg.WiktionaryPassword,
	)

	resp, err := service.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("dictionary: do request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	utils.Logger.Debug(resp.Status)

	var respBody []any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return false, fmt.Errorf("dictionary: decode response: %w", err)
	}

	if len(respBody) == 0 {
		return false, nil
	}

	dataWords, err := json.Marshal(respBody[1])
	if err != nil {
		return false, fmt.Errorf("dictionary: marshal words: %w", err)
	}

	var words []string

	err = json.Unmarshal(dataWords, &words)
	if err != nil {
		return false, fmt.Errorf("dictionary: unmarshal words: %w", err)
	}

	found := slices.Contains(words, word)
	if !found {
		caser := cases.Title(language.English)
		found = slices.Contains(words, caser.String(word))
	}

	return found, nil
}
