package utils

import (
	"sync"

	"github.com/disgoorg/snowflake/v2"
)

// ActiveGames tracks which Discord channels currently have an active game.
// It is updated by each app's game service when games start and end.
var ActiveGames = newActiveGameRegistry()

type activeGameRegistry struct {
	mu        sync.RWMutex
	byGuild   map[snowflake.ID]snowflake.ID // guildID → channelID
	byChannel map[snowflake.ID]bool         // channelID → present
}

func newActiveGameRegistry() *activeGameRegistry {
	return &activeGameRegistry{
		byGuild:   make(map[snowflake.ID]snowflake.ID),
		byChannel: make(map[snowflake.ID]bool),
	}
}

// Register marks channelID as having an active game for guildID.
// If the guild previously had a game in a different channel, the old channel
// entry is removed first.
func (r *activeGameRegistry) Register(guildID, channelID snowflake.ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.byGuild[guildID]; ok && old != channelID {
		delete(r.byChannel, old)
	}
	r.byGuild[guildID] = channelID
	r.byChannel[channelID] = true
}

// Unregister removes the active game entry for guildID.
func (r *activeGameRegistry) Unregister(guildID snowflake.ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if channelID, ok := r.byGuild[guildID]; ok {
		delete(r.byGuild, guildID)
		delete(r.byChannel, channelID)
	}
}

// IsActiveChannel reports whether channelID currently has an active game.
func (r *activeGameRegistry) IsActiveChannel(channelID snowflake.ID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byChannel[channelID]
}
