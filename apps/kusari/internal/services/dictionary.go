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

	"github.com/valkey-io/valkey-go"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	localMetrics "jurien.dev/yugen/kusari/internal/metrics"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/utils"
)

const (
	dictionaryCachePrefix = "kusari:dictionary:"
	dictionaryCacheTTL    = 12 * time.Hour
)

var smartQuoteReplacer = strings.NewReplacer(
	"’", "'",
	"‘", "'",
	"“", `"`,
	"”", `"`,
)

type DictionaryService struct {
	cfg    *config.Config
	client *http.Client
	valkey valkey.Client
}

func CreateDictionaryService(
	cfg *config.Config,
	vk valkey.Client,
) *DictionaryService {
	utils.Logger.Info("Creating Dictionary Service")

	return &DictionaryService{
		cfg:    cfg,
		valkey: vk,
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
	word = smartQuoteReplacer.Replace(word)

	if service.valkey != nil {
		cacheKey := dictionaryCachePrefix + word
		cmd := service.valkey.B().Get().Key(cacheKey).Build()
		val, err := service.valkey.Do(ctx, cmd).ToString()

		if err == nil {
			localMetrics.DictionaryCacheHits.Inc()
			return val == "1", nil
		}

		if !valkey.IsValkeyNil(err) {
			utils.Logger.Warnw("valkey GET failed", "error", err)
		}
	}

	localMetrics.DictionaryCacheMisses.Inc()

	found, err := service.checkWiktionary(ctx, word)
	if err != nil {
		return false, err
	}

	if service.valkey != nil {
		cacheKey := dictionaryCachePrefix + word
		cacheVal := "0"

		if found {
			cacheVal = "1"
		}

		cmd := service.valkey.B().
			Set().
			Key(cacheKey).
			Value(cacheVal).
			Ex(dictionaryCacheTTL).
			Build()
		if setErr := service.valkey.Do(ctx, cmd).Error(); setErr != nil {
			utils.Logger.Warnw("valkey SET failed", "error", setErr)
		}
	}

	return found, nil
}

func (service *DictionaryService) Clear() int {
	if service.valkey == nil {
		return 0
	}

	ctx := context.Background()
	var cursor uint64
	var keys []string

	for {
		cmd := service.valkey.B().
			Scan().
			Cursor(cursor).
			Match(dictionaryCachePrefix + "*").
			Count(100).
			Build()
		entry, err := service.valkey.Do(ctx, cmd).AsScanEntry()
		if err != nil {
			utils.Logger.Warnw("valkey SCAN failed during Clear", "error", err)
			break
		}
		keys = append(keys, entry.Elements...)
		cursor = entry.Cursor
		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return 0
	}

	cmd := service.valkey.B().Del().Key(keys...).Build()
	if err := service.valkey.Do(ctx, cmd).Error(); err != nil {
		utils.Logger.Warnw("valkey DEL failed during Clear", "error", err)
		return 0
	}

	return len(keys)
}

func (service *DictionaryService) checkWiktionary(
	ctx context.Context,
	word string,
) (bool, error) {
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
