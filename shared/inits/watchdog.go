package inits

import (
	"context"
	"errors"
	"time"

	"jurien.dev/yugen/shared/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

const (
	watchdogCheckInterval = 2 * time.Minute
	heartbeatACKTimeout   = 5 * time.Minute
)

// StartShardWatchdog monitors shard heartbeat ACKs and forces a reconnect if
// Discord drops the connection without triggering discordgo's built-in reconnect.
func StartShardWatchdog(bot *discordgoplus.Bot, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(watchdogCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				bot.Each(func(shard *discordgoplus.Bot) {
					shard.RLock()
					lastACK := shard.LastHeartbeatAck
					shard.RUnlock()

					if time.Since(lastACK) <= heartbeatACKTimeout {
						return
					}

					utils.Logger.Warnf(
						"shard %d hasn't received heartbeat ACK in %v, forcing reconnect",
						shard.ShardID,
						time.Since(lastACK).Round(time.Second),
					)

					if err := shard.Close(); err != nil {
						utils.Logger.Errorf("watchdog: shard %d close: %v", shard.ShardID, err)
					}
					if err := shard.Open(); err != nil && !errors.Is(err, discordgo.ErrWSAlreadyOpen) {
						utils.Logger.Errorf("watchdog: shard %d open: %v", shard.ShardID, err)
					}
				})
			}
		}
	}()
}
