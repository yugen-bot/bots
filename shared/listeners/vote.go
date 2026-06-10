package listeners

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func doRequest(url string, token string, body []byte, source string) error {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("vote: doRequest: new request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("vote: doRequest: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("vote: doRequest: read body: %w", err)
	}

	if resp.StatusCode != 200 {
		utils.Logger.With("url", url, "body", string(respBody)).
			Warnf("Something went wrong syncing %s", source)
	}

	utils.Logger.Infof("Synced %s", source)

	return nil
}

func postTopGG(token string, clientID string, servers int, shards int) error {
	url := fmt.Sprintf("https://top.gg/api/bots/%s/stats", clientID)
	body := []byte(
		fmt.Sprintf(`{"server_count": %d, "shard_count": %d}`, servers, shards),
	)

	return doRequest(url, token, body, "top-gg")
}

func postDiscordBotList(token string, clientID string, servers int) error {
	url := fmt.Sprintf(
		"https://discordbotlist.com/api/v1/bots/%s/stats",
		clientID,
	)
	body := []byte(fmt.Sprintf(`{"guilds": %d}`, servers))

	return doRequest(url, token, body, "discordbotlist")
}

func postStats(client *bot.Client, cfg *config.Config) {
	clientID := cfg.DiscordAppID
	servers := client.Caches.GuildsLen()

	if cfg.TopGGSync && len(cfg.TopGGToken) > 0 {
		go func() {
			if err := postTopGG(
				cfg.TopGGToken,
				clientID,
				servers,
				1,
			); err != nil {
				utils.Logger.Errorw("vote: post top.gg failed", "error", err)
			}
		}()
	}

	if cfg.DiscordBotListSync && len(cfg.DiscordBotListToken) > 0 {
		go func() {
			if err := postDiscordBotList(
				cfg.DiscordBotListToken,
				clientID,
				servers,
			); err != nil {
				utils.Logger.Errorw(
					"vote: post discordbotlist failed",
					"error",
					err,
				)
			}
		}()
	}
}

func AddVoteListeners(container *di.Container) {
	disgoBot := container.Get(static.DiBot).(*disgoplus.Bot)
	client := disgoBot.Client()
	cron := container.Get(static.DiCron).(*cron.Cron)
	cfg := container.Get(static.DiConfig).(*config.Config)

	if _, err := cron.AddFunc("@every 3h", func() {
		go postStats(client, cfg)
	}); err != nil {
		panic(err)
	}

	client.EventManager.AddEventListeners(
		bot.NewListenerFunc(func(e *events.GuildsReady) {
			if e.ShardID() != 0 {
				return
			}

			go postStats(client, cfg)
		}),
	)
}
