package utils

// TODO: test requires mock
// ShowLeaderboard, LeaderboardCommandHandler, LeaderboardMessageComponentHandler,
// doLeaderboardResponse, and doError all require a live *disgoplus.Ctx and
// *di.Container (which wraps a Discord client and embed colour).
// They cannot be exercised here without a mock framework; integration-level
// tests would need a fake container satisfying the DI interface.
//
// The leaderboard source-type constants and the helper type aliases are purely
// declarative; there is no branch or computation to cover with a unit test.
// Coverage for this file should come from integration / end-to-end tests.
