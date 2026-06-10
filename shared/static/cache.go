package static

import "github.com/disgoorg/disgo/cache"

// DefaultCacheFlags is the baseline cache configuration shared by all bots.
// Bots that need message caching (kazu, kusari, hoshi) should OR in cache.FlagMessages.
const DefaultCacheFlags = cache.FlagGuilds | cache.FlagMembers | cache.FlagChannels | cache.FlagRoles
