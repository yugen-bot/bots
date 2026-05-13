package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type DictionaryService struct{}

func CreateDictionaryService() *DictionaryService {
	utils.Logger.Info("Creating Dictionary Service")
	return &DictionaryService{}
}

func (service *DictionaryService) Check(word string) (bool, error) {
	word = strings.ToLower(word)

	replacer := strings.NewReplacer(
		"‘", "'",
		"’", "'",
		"“", `"`,
		"”", `"`,
	)
	word = replacer.Replace(word)
	wiktionaryURL := fmt.Sprintf(
		"https://en.wiktionary.org/w/api.php?action=opensearch&format=json&formatversion=2&search=%s&namespace=0&limit=2",
		url.QueryEscape(word),
	)

	client := &http.Client{
		Transport: &http.Transport{},
	}

	utils.Logger.Debug(wiktionaryURL)
	req, err := http.NewRequest("GET", wiktionaryURL, nil)
	if err != nil {
		return false, fmt.Errorf("dictionary: new request: %w", err)
	}

	req.Header.Set("User-Agent", "YugenKusari/1.0 (https://github.com/jurienhamaker/yugen;info@jurien.dev) Go-http-client/1.1")
	req.SetBasicAuth(os.Getenv(static.EnvWiktionaryUsername), os.Getenv(static.EnvWiktionaryPassword))

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("dictionary: do request: %w", err)
	}
	defer resp.Body.Close()

	utils.Logger.Debug(resp.Status)
	utils.Logger.Debug(resp.Body)

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
