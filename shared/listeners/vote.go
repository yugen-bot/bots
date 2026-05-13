package listeners

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func doRequest(url string, token string, body []byte, source string) error {
	contentType := "application/json"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("vote: doRequest: new request: %w", err)
	}

	req.Header.Add("Content-Type", contentType)
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
		utils.Logger.With("url", url, "body", string(respBody)).Warnf("Something went wrong syncing %s", source)
	}

	utils.Logger.Infof("Synced %s", source)

	return nil
}

func postTopGG(token string, clientID string, servers int) error {
	url := fmt.Sprintf("https://top.gg/api/bots/%s/stats", clientID)
	body := []byte(fmt.Sprintf(`{"server_count": %d, "shard_count": 1}`, servers))

	return doRequest(url, token, body, "top-gg")
}

func postDiscordBotList(token string, clientID string, servers int) error {
	url := fmt.Sprintf("https://discordbotlist.com/api/v1/bots/%s/stats", clientID)
	body := []byte(fmt.Sprintf(`{"guilds": %d}`, servers))

	return doRequest(url, token, body, "discordbotlist")
}

func postStats(bot *discordgoplus.Bot) {
	clientID := os.Getenv(static.EnvDiscordAppID)

	servers := len(bot.State.Guilds)

	syncTopGG := os.Getenv(static.EnvTopGGSync) == "true"
	topGGToken := os.Getenv(static.EnvTopGGToken)

	if syncTopGG && len(topGGToken) > 0 {
		go func() {
			if err := postTopGG(topGGToken, clientID, servers); err != nil {
				utils.Logger.Errorw("vote: post top.gg failed", "error", err)
			}
		}()
	}

	syncDiscordBotList := os.Getenv(static.EnvDiscordBotListSync) == "true"
	discordBotListToken := os.Getenv(static.EnvDiscordBotListToken)

	if syncDiscordBotList && len(discordBotListToken) > 0 {
		go func() {
			if err := postDiscordBotList(discordBotListToken, clientID, servers); err != nil {
				utils.Logger.Errorw("vote: post discordbotlist failed", "error", err)
			}
		}()
	}
}

func AddVoteListeners(container *di.Container) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)
	cron := container.Get(static.DiCron).(*cron.Cron)

	cron.AddFunc("@every 3h", func() {
		go postStats(bot)
	})

	bot.AddHandler(func(session *discordgo.Session, event *discordgo.Ready) {
		go postStats(bot)
	})
}
