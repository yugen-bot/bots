package static

import "github.com/disgoorg/disgo/cache"

// DefaultCacheFlags is the baseline cache configuration shared by all bots.
const DefaultCacheFlags = cache.FlagGuilds | cache.FlagMembers | cache.FlagChannels | cache.FlagRoles
