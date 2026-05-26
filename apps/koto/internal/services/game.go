package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/koto/prisma/db"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type cooldownResult struct {
	Hit          bool
	RepeatHit    bool
	Result       time.Time
	RepeatResult time.Time
}

type GameService struct {
	database *db.PrismaClient
	settings *SettingsService
	words    *WordsService
	message  *MessageService
	points   *PointsService
	bot      *discordgoplus.Bot
}

func CreateGameService(container *di.Container) *GameService {
	utils.Logger.Info("Creating Game Service")

	return &GameService{
		database: container.Get(sharedStatic.DiDatabase).(*db.PrismaClient),
		settings: container.Get(sharedStatic.DiSettings).(*SettingsService),
		words:    container.Get(localStatic.DiWords).(*WordsService),
		message:  container.Get(localStatic.DiGameMessage).(*MessageService),
		points:   container.Get(localStatic.DiPoints).(*PointsService),
		bot:      container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

// Start creates a new game. schedule=true means cron-started. recreate=true ends current game first.
// word="" means pick randomly.
func (s *GameService) Start(
	ctx context.Context,
	guildID string,
	schedule bool,
	recreate bool,
	word string,
) (bool, error) {
	utils.Logger.Debugf("Trying to start a game for %s", guildID)

	b, err := s.bot.ShardByGuild(guildID)
	if err != nil {
		utils.Logger.Debugf(
			"Skipping game start, ShardByGuild failed for guild %s: %v",
			guildID,
			err,
		)

		return false, nil
	}

	if _, gErr := b.State.Guild(guildID); gErr != nil {
		if _, gErr2 := b.Guild(guildID); gErr2 != nil {
			utils.Logger.Debugf(
				"Skipping game start, bot not in guild %s",
				guildID,
			)

			return false, nil
		}
	}

	currentGame, err := s.GetCurrentGame(ctx, guildID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return false, err
	}

	if currentGame != nil && !recreate {
		utils.Logger.Debugf(
			"Already active game %d (no recreate) for %s",
			currentGame.ID,
			guildID,
		)

		return false, nil
	}

	if currentGame != nil && recreate {
		if endErr := s.EndGame(
			ctx,
			currentGame.ID,
			db.GameStatusFailed,
		); endErr != nil {
			utils.Logger.Warnw("game: start: end current failed",
				"error", endErr,
				"guildID", guildID,
				"gameID", currentGame.ID,
			)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Get past 50 games to avoid repeating words
	pastGames, err := s.database.Game.FindMany(
		db.Game.GuildID.Equals(guildID),
	).Select(
		db.Game.Word.Field(),
		db.Game.Number.Field(),
	).OrderBy(
		db.Game.CreatedAt.Order(db.SortOrderDesc),
	).Take(50).Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return false, fmt.Errorf("game: start: find past games: %w", err)
	}

	var lastNumber int
	if len(pastGames) > 0 {
		lastNumber = pastGames[0].Number
	}

	ignoredWords := make([]string, len(pastGames))
	for i, g := range pastGames {
		ignoredWords[i] = g.Word
	}

	if word == "" {
		word = s.words.GetRandom(ignoredWords, false)
	}

	if word == "" {
		return false, fmt.Errorf(
			"game: start: could not pick word for %s",
			guildID,
		)
	}

	settings, err := s.settings.GetByGuildID(ctx, guildID)
	if err != nil {
		return false, fmt.Errorf("game: start: get settings: %w", err)
	}

	var endingAt time.Time
	if settings.StartAfterFirstGuess {
		// Year-3000 sentinel: timer doesn't start until first guess
		endingAt = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		endingAt = roundToNearestMinute(
			time.Now().Add(time.Duration(settings.TimeLimit) * time.Minute),
		)
	}

	baseMeta := s.createBaseState(word)

	metaJSON, err := json.Marshal(baseMeta)
	if err != nil {
		return false, fmt.Errorf("game: start: marshal meta: %w", err)
	}

	game, err := s.database.Game.CreateOne(
		db.Game.Settings.Link(db.Settings.GuildID.Equals(guildID)),
		db.Game.Word.Set(word),
		db.Game.EndingAt.Set(endingAt),
		db.Game.ScheduleStarted.Set(schedule),
		db.Game.Meta.Set(metaJSON),
		db.Game.Number.Set(lastNumber+1),
	).Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("game: start: create game: %w", err)
	}

	if err := s.message.Create(ctx, game, []db.GuessModel{}, true); err != nil {
		utils.Logger.Warnw("game: start: create message failed",
			"error", err,
			"guildID", guildID,
			"gameID", game.ID,
		)

		// if we fail to send the message because the channel is not found or inaccessible, reset channel.
		if strings.Contains(err.Error(), "404 Not Found") ||
			strings.Contains(err.Error(), "403 Forbidden") {
			if _, err := s.settings.Set(
				context.Background(),
				guildID,
				db.Settings.ChannelID.SetOptional(nil),
			); err != nil {
				utils.Logger.Warnw("game: start: reset channel failed",
					"error", err)
			}

			if endErr := s.EndGame(
				ctx,
				game.ID,
				db.GameStatusFailed,
			); endErr != nil {
				utils.Logger.Warnw("game: start: end current failed",
					"error", endErr,
					"guildID", guildID,
					"gameID", currentGame.ID,
				)
			}

			reason := "Forbidden"
			if strings.Contains(err.Error(), "404 Not Found") {
				reason = "Not found"
			}

			return false, fmt.Errorf("game: start: %s: %w", reason, err)
		}
	}

	return true, nil
}

// Guess processes a player's word guess.
func (s *GameService) Guess(
	ctx context.Context,
	guildID string,
	word string,
	message *discordgo.Message,
	settings *db.SettingsModel,
) error {
	game, err := s.GetCurrentGame(ctx, guildID)
	if err != nil {
		return fmt.Errorf("game: guess: get current game: %w", err)
	}

	if game == nil {
		return nil
	}

	b, err := s.bot.ShardByGuild(guildID)
	if err != nil {
		utils.Logger.Warnw(
			"game: guess: ShardByGuild failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return nil
	}

	// Ensure player exists
	if _, pErr := s.points.GetPlayer(
		ctx,
		guildID,
		message.Author.ID,
	); pErr != nil {
		utils.Logger.Warnw("game: guess: get player failed",
			"error", pErr,
			"guildID", guildID,
			"gameID", game.ID,
			"userID", message.Author.ID,
		)
	}

	// Fetch guesses
	guesses, err := s.database.Guess.FindMany(
		db.Guess.GameID.Equals(game.ID),
	).OrderBy(
		db.Guess.CreatedAt.Order(db.SortOrderDesc),
	).Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return fmt.Errorf("game: guess: find guesses: %w", err)
	}

	// Dedupe check (production only)
	if os.Getenv("ENV") == "production" {
		for _, g := range guesses {
			if g.Word == word {
				b.MessageReactionAdd(
					message.ChannelID,
					message.ID,
					"❌",
				)

				return nil
			}
		}
	}

	// Cooldown check
	cooldown := s.checkCooldown(message.Author.ID, guesses, settings)
	if cooldown.Hit || cooldown.RepeatHit {
		b.MessageReactionAdd(
			message.ChannelID,
			message.ID,
			"🕒",
		)

		suffix := fmt.Sprintf(
			"you can guess again <t:%d:R>",
			cooldown.Result.Unix(),
		)
		if cooldown.Hit && cooldown.RepeatHit {
			suffix = fmt.Sprintf(
				"you can guess again <t:%d:R> on your own or <t:%d:R> after a guess from another player.",
				cooldown.RepeatResult.Unix(),
				cooldown.Result.Unix(),
			)
		} else if !cooldown.Hit && cooldown.RepeatHit {
			suffix = fmt.Sprintf(
				"you can guess again <t:%d:R> or immediately after a guess from another player.",
				cooldown.RepeatResult.Unix(),
			)
		}

		_, sendErr := b.ChannelMessageSendReply(
			message.ChannelID,
			fmt.Sprintf("You're on a cooldown, %s", suffix),
			message.Reference(),
		)
		utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)

		return nil
	}

	// Parse current game meta
	gameMeta, err := localUtils.ParseGameMeta(json.RawMessage(game.Meta))
	if err != nil {
		return fmt.Errorf("game: guess: parse meta: %w", err)
	}

	// Score the guess
	guessMeta, guessed, points, updatedGameMeta := s.checkWord(
		game.Word,
		word,
		gameMeta,
	)

	// Persist the guess
	guessMetaJSON, err := json.Marshal(guessMeta)
	if err != nil {
		return fmt.Errorf("game: guess: marshal guess meta: %w", err)
	}

	createdGuess, err := s.database.Guess.CreateOne(
		db.Guess.UserID.Set(message.Author.ID),
		db.Guess.Game.Link(db.Game.ID.Equals(game.ID)),
		db.Guess.Word.Set(word),
		db.Guess.Points.Set(points),
		db.Guess.Meta.Set(guessMetaJSON),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("game: guess: create guess: %w", err)
	}

	// Determine new game status
	newStatus := game.Status
	if guessed {
		newStatus = db.GameStatusCompleted
	} else if len(guesses)+1 >= localStatic.MaxGuesses {
		newStatus = db.GameStatusFailed
	}

	// Update game meta + status + possibly start timer
	updatedMetaJSON, err := json.Marshal(updatedGameMeta)
	if err != nil {
		return fmt.Errorf("game: guess: marshal game meta: %w", err)
	}

	updateParams := []db.GameSetParam{
		db.Game.Status.Set(newStatus),
		db.Game.Meta.Set(updatedMetaJSON),
	}
	// If startAfterFirstGuess and this is the first guess, set real endingAt
	if settings.StartAfterFirstGuess && len(guesses) == 0 {
		realEndingAt := roundToNearestMinute(
			time.Now().Add(time.Duration(settings.TimeLimit) * time.Minute),
		)
		updateParams = append(updateParams, db.Game.EndingAt.Set(realEndingAt))
	}

	updatedGame, err := s.database.Game.FindUnique(
		db.Game.ID.Equals(game.ID),
	).Update(updateParams...).Exec(ctx)
	if err != nil {
		return fmt.Errorf("game: guess: update game: %w", err)
	}

	// React to guess message
	if guessed {
		b.MessageReactionAdd(
			message.ChannelID,
			message.ID,
			"🎉",
		)
	} else {
		b.MessageReactionAdd(message.ChannelID, message.ID, "✅")
	}

	// Fetch updated guesses list (with new guess)
	allGuesses, _ := s.database.Guess.FindMany(
		db.Guess.GameID.Equals(game.ID),
	).OrderBy(
		db.Guess.CreatedAt.Order(db.SortOrderDesc),
	).Exec(ctx)

	// Apply points if won
	if guessed {
		go func() {
			if applyErr := s.points.ApplyPoints(
				context.Background(),
				updatedGame,
				allGuesses,
				message.Author.ID,
			); applyErr != nil {
				utils.Logger.Warnw("game: guess: apply points failed",
					"error", applyErr,
					"guildID", guildID,
					"gameID", game.ID,
				)
			}
		}()
	}

	// Inform cooldown
	if newStatus != db.GameStatusCompleted &&
		settings.InformCooldownAfterGuess {
		go func() {
			backToBackPart := ""
			if settings.EnableBackToBackCooldown {
				backToBackPart = fmt.Sprintf(
					"<t:%d:R> on your own or ",
					createdGuess.CreatedAt.Add(time.Duration(settings.BackToBackCooldown)*time.Second).
						Unix(),
				)
			}

			afterPart := ""
			if settings.EnableBackToBackCooldown {
				afterPart = " after a guess from another player"
			}

			msg := fmt.Sprintf(
				"You are now on a cooldown. You can guess again %s<t:%d:R>%s.",
				backToBackPart,
				createdGuess.CreatedAt.Add(time.Duration(settings.Cooldown)*time.Second).
					Unix(),
				afterPart,
			)
			_, sendErr := b.ChannelMessageSendReply(
				message.ChannelID,
				msg,
				message.Reference(),
			)
			utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
		}()
	}

	// Recreate embed message
	go func() {
		if msgErr := s.message.Create(
			context.Background(),
			updatedGame,
			allGuesses,
			false,
		); msgErr != nil {
			utils.Logger.Warnw("game: guess: recreate message failed",
				"error", msgErr,
				"guildID", guildID,
				"gameID", game.ID,
			)
		}
	}()

	// Handle terminal status
	if newStatus != db.GameStatusInProgress {
		// Delete guesses for privacy
		if _, delErr := s.database.Guess.FindMany(
			db.Guess.GameID.Equals(updatedGame.ID),
		).Delete().Exec(context.Background()); delErr != nil {
			utils.Logger.Warnw(
				"game: guess: delete guesses failed",
				"error", delErr,
				"guildID", guildID,
				"gameID", updatedGame.ID,
			)
		}

		// Auto-start if configured
		if settings.AutoStart {
			time.Sleep(500 * time.Millisecond)

			go func() {
				if _, startErr := s.Start(
					context.Background(),
					guildID,
					false,
					false,
					"",
				); startErr != nil {
					utils.Logger.Warnw(
						"game: guess: auto-start failed",
						"guildID", guildID,
						"error", startErr,
					)
				}
			}()
		}
	}

	return nil
}

// EndGame ends a game with the given status, recreates message, deletes guesses.
func (s *GameService) EndGame(
	ctx context.Context,
	gameID int,
	status db.GameStatus,
) error {
	game, err := s.database.Game.FindUnique(
		db.Game.ID.Equals(gameID),
	).Update(
		db.Game.Status.Set(status),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("game: end: update: %w", err)
	}

	b, err := s.bot.ShardByGuild(game.GuildID)
	if err != nil {
		utils.Logger.Debugf(
			"Skipping end message, ShardByGuild failed for guild %s: %v",
			game.GuildID,
			err,
		)

		return nil
	}

	if _, gErr := b.State.Guild(game.GuildID); gErr != nil {
		if _, gErr2 := b.Guild(game.GuildID); gErr2 != nil {
			utils.Logger.Debugf(
				"Skipping end message, bot not in guild %s",
				game.GuildID,
			)

			return nil
		}
	}

	guesses, _ := s.database.Guess.FindMany(
		db.Guess.GameID.Equals(game.ID),
	).Exec(ctx)

	if msgErr := s.message.Create(ctx, game, guesses, false); msgErr != nil {
		utils.Logger.Warnw("game: end: create message failed",
			"error", msgErr,
			"guildID", game.GuildID,
			"gameID", game.ID,
		)
	}

	// Delete guesses for privacy
	if _, delErr := s.database.Guess.FindMany(
		db.Guess.GameID.Equals(game.ID),
	).Delete().Exec(ctx); delErr != nil {
		utils.Logger.Warnw(
			"game: end: delete guesses failed",
			"error", delErr,
			"guildID", game.GuildID,
			"gameID", game.ID,
		)
	}

	return nil
}

// GetCurrentGame returns the active (IN_PROGRESS) game for a guild, or nil if none.
func (s *GameService) GetCurrentGame(
	ctx context.Context,
	guildID string,
) (*db.GameModel, error) {
	game, err := s.database.Game.FindFirst(
		db.Game.GuildID.Equals(guildID),
		db.Game.Status.Equals(db.GameStatusInProgress),
		db.Game.EndingAt.After(time.Now()),
	).OrderBy(
		db.Game.CreatedAt.Order(db.SortOrderDesc),
	).Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		return nil, nil
	}

	return game, err
}

// checkWord scores a guess against the target word.
// Returns: per-letter meta, guessed(bool), total points, updated game meta.
func (s *GameService) checkWord(
	word string,
	guess string,
	state *localUtils.GameMeta,
) (localUtils.GuessMetaSlice, bool, int, *localUtils.GameMeta) {
	meta := make(localUtils.GuessMetaSlice, len(guess))
	unmatched := map[rune]int{}   // unmatched word letters
	letterCount := map[rune]int{} // matched count per letter

	wordRunes := []rune(word)
	guessRunes := []rune(guess)

	// Pass 1: find exact matches (CORRECT)
	for i, letter := range wordRunes {
		if i >= len(guessRunes) {
			break
		}

		if letter == guessRunes[i] {
			letterCount[letter]++

			var pts int
			if prev, ok := state.Discovery.Correct[string(letter)]; ok &&
				prev < letterCount[letter] {
				pts = 2

				state.Discovery.Correct[string(letter)] = letterCount[letter]
				if prev2, ok2 := state.Discovery.Almost[string(letter)]; ok2 &&
					prev2 >= letterCount[letter] {
					pts = 1
				}
			}

			if prev, ok := state.Discovery.Almost[string(letter)]; !ok ||
				prev < letterCount[letter] {
				state.Discovery.Almost[string(letter)] = letterCount[letter]
			}

			meta[i] = localUtils.GuessMeta{
				Type:   localUtils.GameTypeCorrect,
				Points: pts,
				Letter: string(letter),
			}

			if existing, ok := state.Keyboard[string(letter)]; !ok ||
				existing != localUtils.GameTypeCorrect {
				state.Keyboard[string(letter)] = localUtils.GameTypeCorrect
			}

			if i < len(state.Word) {
				state.Word[i].Type = localUtils.GameTypeCorrect
			}

			continue
		}

		unmatched[letter]++
	}

	// Pass 2: handle non-exact-match positions (ALMOST or WRONG)
	for i, element := range wordRunes {
		if i >= len(guessRunes) {
			break
		}

		letter := guessRunes[i]
		if letter != element {
			if unmatched[letter] > 0 {
				letterCount[letter]++

				var pts int
				if prev, ok := state.Discovery.Almost[string(letter)]; !ok ||
					prev < letterCount[letter] {
					pts = 1
					state.Discovery.Almost[string(letter)] = letterCount[letter]
				}

				meta[i] = localUtils.GuessMeta{
					Type:   localUtils.GameTypeAlmost,
					Points: pts,
					Letter: string(letter),
				}
				unmatched[letter]--

				if existing, ok := state.Keyboard[string(letter)]; !ok ||
					existing != localUtils.GameTypeCorrect {
					state.Keyboard[string(letter)] = localUtils.GameTypeAlmost
				}

				continue
			}

			meta[i] = localUtils.GuessMeta{
				Type:   localUtils.GameTypeWrong,
				Points: 0,
				Letter: string(letter),
			}

			if _, ok := state.Keyboard[string(letter)]; !ok {
				state.Keyboard[string(letter)] = localUtils.GameTypeWrong
			}
		}
	}

	// Calculate total points
	totalPoints := 0
	for _, m := range meta {
		totalPoints += m.Points
	}

	return meta, word == guess, totalPoints, state
}

// checkCooldown returns cooldown status for a user.
func (s *GameService) checkCooldown(
	userID string,
	guesses []db.GuessModel,
	settings *db.SettingsModel,
) cooldownResult {
	if len(guesses) == 0 {
		return cooldownResult{}
	}

	// guesses are ordered desc by createdAt from the query
	lastGuess := &guesses[0]

	var lastGuessByUser *db.GuessModel

	for i := range guesses {
		if guesses[i].UserID == userID {
			lastGuessByUser = &guesses[i]
			break
		}
	}

	if lastGuess == nil || lastGuessByUser == nil {
		return cooldownResult{}
	}

	now := time.Now()
	backToBackHit := settings.EnableBackToBackCooldown &&
		lastGuessByUser.CreatedAt.After(
			now.Add(-time.Duration(settings.BackToBackCooldown)*time.Second),
		) &&
		userID == lastGuess.UserID
	cooldownHit := lastGuessByUser.CreatedAt.After(
		now.Add(-time.Duration(settings.Cooldown) * time.Second),
	)

	if backToBackHit || cooldownHit {
		return cooldownResult{
			Hit:       cooldownHit,
			RepeatHit: backToBackHit,
			Result: lastGuessByUser.CreatedAt.Add(
				time.Duration(settings.Cooldown) * time.Second,
			),
			RepeatResult: lastGuessByUser.CreatedAt.Add(
				time.Duration(settings.BackToBackCooldown) * time.Second,
			),
		}
	}

	return cooldownResult{}
}

// createBaseState initializes the game meta for a new game.
func (s *GameService) createBaseState(word string) *localUtils.GameMeta {
	discovery := localUtils.DiscoveryState{
		Almost:  map[string]int{},
		Correct: map[string]int{},
	}

	wordStates := make([]localUtils.WordState, len([]rune(word)))
	for i, letter := range []rune(word) {
		discovery.Almost[string(letter)] = 0
		discovery.Correct[string(letter)] = 0
		wordStates[i] = localUtils.WordState{
			Index:  i,
			Letter: string(letter),
			Type:   localUtils.GameTypeWrong,
		}
	}

	return &localUtils.GameMeta{
		Keyboard:  map[string]localUtils.GameType{},
		Word:      wordStates,
		Discovery: discovery,
	}
}

// roundToNearestMinute rounds t to nearest minute.
func roundToNearestMinute(t time.Time) time.Time {
	return t.Round(time.Minute)
}

func (s *GameService) CountByGuildIDs(
	ctx context.Context,
	guildIDs []string,
) (games int, guesses int, err error) {
	gameResult, err := s.database.Game.FindMany(
		db.Game.GuildID.In(guildIDs),
	).Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return 0, 0, fmt.Errorf("game: count by guild ids: games: %w", err)
	}

	guessResult, err := s.database.Guess.FindMany(
		db.Guess.Game.Where(db.Game.GuildID.In(guildIDs)),
	).Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return 0, 0, fmt.Errorf("game: count by guild ids: guesses: %w", err)
	}

	return len(gameResult), len(guessResult), nil
}

func (s *GameService) DeleteByGuildIDs(
	ctx context.Context,
	guildIDs []string,
) (games int, guesses int, err error) {
	guessResult, gErr := s.database.Guess.FindMany(
		db.Guess.Game.Where(db.Game.GuildID.In(guildIDs)),
	).Delete().Exec(ctx)
	if gErr != nil && !errors.Is(gErr, db.ErrNotFound) {
		return 0, 0, fmt.Errorf("game: delete by guild ids: guesses: %w", gErr)
	}

	gameResult, err := s.database.Game.FindMany(
		db.Game.GuildID.In(guildIDs),
	).Delete().Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return 0, 0, fmt.Errorf("game: delete by guild ids: games: %w", err)
	}

	return gameResult.Count, guessResult.Count, nil
}

type GuildIDRow struct {
	GuildID string `json:"guildId"`
}

func (s *GameService) FindAllGuildIDs(
	ctx context.Context,
) ([]GuildIDRow, error) {
	var rows []GuildIDRow
	if err := s.database.Prisma.QueryRaw(
		`SELECT DISTINCT "guildId" FROM "Game"`,
	).Exec(ctx, &rows); err != nil {
		return nil, fmt.Errorf("game: find distinct guild ids: %w", err)
	}

	return rows, nil
}
