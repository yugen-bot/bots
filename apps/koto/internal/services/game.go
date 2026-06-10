package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/ent"
	entgame "jurien.dev/yugen/koto/internal/ent/game"
	"jurien.dev/yugen/koto/internal/ent/guess"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
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
	database   *ent.Client
	settings   *SettingsService
	words      *WordsService
	message    *MessageService
	points     *PointsService
	hints      *HintsService
	bot        *disgoplus.Bot
	client     *bot.Client
	startLocks sync.Map // keyed by guildID → *sync.Mutex
}

var (
	ErrNoHints         = fmt.Errorf("game: hint: no hints available")
	ErrHintUnavailable = fmt.Errorf(
		"game: hint: hint not available for current state",
	)
)

func CreateGameService(container *di.Container) *GameService {
	utils.Logger.Info("Creating Game Service")

	b := container.Get(sharedStatic.DiClient).(*disgoplus.Bot)

	return &GameService{
		database: container.Get(sharedStatic.DiDatabase).(*ent.Client),
		settings: container.Get(sharedStatic.DiSettings).(*SettingsService),
		words:    container.Get(localStatic.DiWords).(*WordsService),
		message:  container.Get(localStatic.DiGameMessage).(*MessageService),
		points:   container.Get(localStatic.DiPoints).(*PointsService),
		hints:    container.Get(localStatic.DiHints).(*HintsService),
		bot:      b,
		client:   b.Client(),
	}
}

func (s *GameService) lockFor(guildID string) *sync.Mutex {
	v, _ := s.startLocks.LoadOrStore(guildID, &sync.Mutex{})
	return v.(*sync.Mutex)
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

	mu := s.lockFor(guildID)
	mu.Lock()
	defer mu.Unlock()

	if !utils.IsBotInGuildClient(s.client, guildID) {
		utils.Logger.Debugf(
			"Skipping game start, bot not in guild %s",
			guildID,
		)

		return false, nil
	}

	currentGame, err := s.GetCurrentGame(ctx, guildID)
	if err != nil {
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
			entgame.StatusFAILED,
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
	pastGames, err := s.database.Game.Query().
		Where(entgame.GuildIDEQ(guildID)).
		Order(ent.Desc(entgame.FieldCreatedAt)).
		Limit(50).
		All(ctx)
	if err != nil && !ent.IsNotFound(err) {
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

	guildSettings, err := s.settings.GetByGuildID(ctx, guildID)
	if err != nil {
		return false, fmt.Errorf("game: start: get settings: %w", err)
	}

	var endingAt time.Time
	if guildSettings.StartAfterFirstGuess {
		// Year-3000 sentinel: timer doesn't start until first guess
		endingAt = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		endingAt = roundToNearestMinute(
			time.Now().Add(time.Duration(guildSettings.TimeLimit) * time.Minute),
		)
	}

	baseMeta := s.createBaseState(word)

	metaJSON, err := json.Marshal(baseMeta)
	if err != nil {
		return false, fmt.Errorf("game: start: marshal meta: %w", err)
	}

	newGame, err := s.database.Game.Create().
		SetGuildID(guildID).
		SetWord(word).
		SetEndingAt(endingAt).
		SetScheduleStarted(schedule).
		SetMeta(metaJSON).
		SetNumber(lastNumber + 1).
		Save(ctx)
	if err != nil {
		return false, fmt.Errorf("game: start: create game: %w", err)
	}

	if err := s.message.Create(ctx, newGame, []*ent.Guess{}, true); err != nil {
		utils.Logger.Warnw("game: start: create message failed",
			"error", err,
			"guildID", guildID,
			"gameID", newGame.ID,
		)

		// if we fail to send the message because the channel is not found or inaccessible, reset channel.
		if strings.Contains(err.Error(), "404 Not Found") ||
			strings.Contains(err.Error(), "403 Forbidden") {
			if _, updateErr := s.settings.Update(
				context.Background(),
				guildSettings.ID,
				func(u *ent.SettingsUpdateOne) { u.ClearChannelID() },
			); updateErr != nil {
				utils.Logger.Warnw("game: start: reset channel failed",
					"error", updateErr)
			}

			if endErr := s.EndGame(
				ctx,
				newGame.ID,
				entgame.StatusFAILED,
			); endErr != nil {
				utils.Logger.Warnw("game: start: end current failed",
					"error", endErr,
					"guildID", guildID,
					"gameID", newGame.ID,
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
	message discord.Message,
	guildSettings *ent.Settings,
) error {
	currentGame, err := s.GetCurrentGame(ctx, guildID)
	if err != nil {
		return fmt.Errorf("game: guess: get current game: %w", err)
	}

	if currentGame == nil {
		return nil
	}

	// Ensure player exists
	if _, pErr := s.points.GetPlayer(
		ctx,
		guildID,
		message.Author.ID.String(),
	); pErr != nil {
		utils.Logger.Warnw("game: guess: get player failed",
			"error", pErr,
			"guildID", guildID,
			"gameID", currentGame.ID,
			"userID", message.Author.ID.String(),
		)
	}

	// Fetch guesses
	guesses, err := s.database.Guess.Query().
		Where(guess.GameIDEQ(currentGame.ID)).
		Order(ent.Desc(guess.FieldCreatedAt)).
		All(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("game: guess: find guesses: %w", err)
	}

	// Dedupe check (production only)
	if os.Getenv("ENV") == "production" {
		for _, g := range guesses {
			if g.Word == word {
				utils.LogIfErr(utils.Logger, "reaction-add",
					s.client.Rest.AddReaction(message.ChannelID, message.ID, "❌"),
				)

				return nil
			}
		}
	}

	// Cooldown check
	cooldown := s.checkCooldown(message.Author.ID.String(), guesses, guildSettings)
	if cooldown.Hit || cooldown.RepeatHit {
		utils.LogIfErr(utils.Logger, "reaction-add",
			s.client.Rest.AddReaction(message.ChannelID, message.ID, "🕒"),
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

		msgID := message.ID
		_, sendErr := s.client.Rest.CreateMessage(message.ChannelID, discord.MessageCreate{
			Content: fmt.Sprintf("You're on a cooldown, %s", suffix),
			MessageReference: &discord.MessageReference{
				MessageID: &msgID,
			},
		})
		utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)

		return nil
	}

	// Parse current game meta
	gameMeta, err := localUtils.ParseGameMeta(json.RawMessage(currentGame.Meta))
	if err != nil {
		return fmt.Errorf("game: guess: parse meta: %w", err)
	}

	// Score the guess
	guessMeta, guessed, points, updatedGameMeta := s.checkWord(
		currentGame.Word,
		word,
		gameMeta,
	)

	// Persist the guess
	guessMetaJSON, err := json.Marshal(guessMeta)
	if err != nil {
		return fmt.Errorf("game: guess: marshal guess meta: %w", err)
	}

	createdGuess, err := s.database.Guess.Create().
		SetUserID(message.Author.ID.String()).
		SetGameID(currentGame.ID).
		SetWord(word).
		SetPoints(points).
		SetMeta(guessMetaJSON).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("game: guess: create guess: %w", err)
	}

	// Determine new game status
	newStatus := currentGame.Status
	if guessed {
		newStatus = entgame.StatusCOMPLETED
	} else if len(guesses)+1 >= localStatic.MaxGuesses {
		newStatus = entgame.StatusFAILED
	}

	// Recompute hint availability after scoring
	updatedGameMeta.CanHint = localUtils.ComputeCanHint(
		currentGame.Word,
		updatedGameMeta,
	)

	// Update game meta + status + possibly start timer
	updatedMetaJSON, err := json.Marshal(updatedGameMeta)
	if err != nil {
		return fmt.Errorf("game: guess: marshal game meta: %w", err)
	}

	upd := s.database.Game.UpdateOneID(currentGame.ID).
		SetStatus(newStatus).
		SetMeta(updatedMetaJSON)

	// If startAfterFirstGuess and this is the first guess, set real endingAt
	if guildSettings.StartAfterFirstGuess && len(guesses) == 0 {
		realEndingAt := roundToNearestMinute(
			time.Now().Add(time.Duration(guildSettings.TimeLimit) * time.Minute),
		)
		upd = upd.SetEndingAt(realEndingAt)
	}

	updatedGame, err := upd.Save(ctx)
	if err != nil {
		return fmt.Errorf("game: guess: update game: %w", err)
	}

	// React to guess message
	if guessed {
		utils.LogIfErr(utils.Logger, "reaction-add",
			s.client.Rest.AddReaction(message.ChannelID, message.ID, "🎉"),
		)
	} else {
		utils.LogIfErr(utils.Logger, "reaction-add",
			s.client.Rest.AddReaction(message.ChannelID, message.ID, "✅"),
		)
	}

	// Fetch updated guesses list (with new guess)
	allGuesses, _ := s.database.Guess.Query().
		Where(guess.GameIDEQ(currentGame.ID)).
		Order(ent.Desc(guess.FieldCreatedAt)).
		All(ctx)

	// Apply points if won
	if guessed {
		go func() {
			if applyErr := s.points.ApplyPoints(
				context.Background(),
				updatedGame,
				allGuesses,
				message.Author.ID.String(),
			); applyErr != nil {
				utils.Logger.Warnw("game: guess: apply points failed",
					"error", applyErr,
					"guildID", guildID,
					"gameID", currentGame.ID,
				)
			}
		}()
	}

	// Inform cooldown
	if newStatus != entgame.StatusCOMPLETED &&
		guildSettings.InformCooldownAfterGuess {
		go func() {
			backToBackPart := ""
			if guildSettings.EnableBackToBackCooldown {
				backToBackPart = fmt.Sprintf(
					"<t:%d:R> on your own or ",
					createdGuess.CreatedAt.Add(time.Duration(guildSettings.BackToBackCooldown)*time.Second).
						Unix(),
				)
			}

			afterPart := ""
			if guildSettings.EnableBackToBackCooldown {
				afterPart = " after a guess from another player"
			}

			msg := fmt.Sprintf(
				"You are now on a cooldown. You can guess again %s<t:%d:R>%s.",
				backToBackPart,
				createdGuess.CreatedAt.Add(time.Duration(guildSettings.Cooldown)*time.Second).
					Unix(),
				afterPart,
			)
			msgID := message.ID
			_, sendErr := s.client.Rest.CreateMessage(message.ChannelID, discord.MessageCreate{
				Content: msg,
				MessageReference: &discord.MessageReference{
					MessageID: &msgID,
				},
			})
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
				"gameID", currentGame.ID,
			)
		}
	}()

	// Handle terminal status
	if newStatus != entgame.StatusIN_PROGRESS {
		// Delete guesses for privacy; keep the solving guess for completed games.
		delQuery := s.database.Guess.Delete().
			Where(guess.GameIDEQ(updatedGame.ID))
		if newStatus == entgame.StatusCOMPLETED {
			delQuery = s.database.Guess.Delete().
				Where(
					guess.GameIDEQ(updatedGame.ID),
					guess.WordNotIn(updatedGame.Word),
				)
		}
		if _, delErr := delQuery.Exec(context.Background()); delErr != nil {
			utils.Logger.Warnw(
				"game: guess: delete guesses failed",
				"error", delErr,
				"guildID", guildID,
				"gameID", updatedGame.ID,
			)
		}

		// Auto-start if configured
		if guildSettings.AutoStart {
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
	status entgame.Status,
) error {
	endedGame, err := s.database.Game.UpdateOneID(gameID).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("game: end: update: %w", err)
	}

	if !utils.IsBotInGuildClient(s.client, endedGame.GuildID) {
		utils.Logger.Debugf(
			"Skipping end message, bot not in guild %s",
			endedGame.GuildID,
		)

		return nil
	}

	guesses, _ := s.database.Guess.Query().
		Where(guess.GameIDEQ(endedGame.ID)).
		All(ctx)

	if msgErr := s.message.Create(ctx, endedGame, guesses, false); msgErr != nil {
		utils.Logger.Warnw("game: end: create message failed",
			"error", msgErr,
			"guildID", endedGame.GuildID,
			"gameID", endedGame.ID,
		)
	}

	// Delete guesses for privacy; keep the solving guess for completed games.
	delQuery := s.database.Guess.Delete().
		Where(guess.GameIDEQ(endedGame.ID))
	if status == entgame.StatusCOMPLETED {
		delQuery = s.database.Guess.Delete().
			Where(
				guess.GameIDEQ(endedGame.ID),
				guess.WordNotIn(endedGame.Word),
			)
	}
	if _, delErr := delQuery.Exec(ctx); delErr != nil {
		utils.Logger.Warnw(
			"game: end: delete guesses failed",
			"error", delErr,
			"guildID", endedGame.GuildID,
			"gameID", endedGame.ID,
		)
	}

	return nil
}

// LastSolvedGame holds the last completed game and the userID of whoever solved it.
type LastSolvedGame struct {
	Game   *ent.Game
	Solver string // userID; empty if unknown
}

// GetLastSolvedGame returns the most recent COMPLETED game for the guild and the
// userID of the player whose guess matched the word. Both fields may be nil/empty
// when no completed game exists yet.
func (s *GameService) GetLastSolvedGame(
	ctx context.Context,
	guildID string,
) (*LastSolvedGame, error) {
	lastGame, err := s.database.Game.Query().
		Where(
			entgame.GuildIDEQ(guildID),
			entgame.StatusEQ(entgame.StatusCOMPLETED),
		).
		Order(ent.Desc(entgame.FieldCreatedAt)).
		First(ctx)

	if ent.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("game: last solved: %w", err)
	}

	result := &LastSolvedGame{Game: lastGame}

	solverGuess, gErr := s.database.Guess.Query().
		Where(
			guess.GameIDEQ(lastGame.ID),
			guess.WordEQ(lastGame.Word),
		).
		First(ctx)
	if gErr == nil {
		result.Solver = solverGuess.UserID
	}

	return result, nil
}

// GetCurrentGame returns the active (IN_PROGRESS) game for a guild, or nil if none.
func (s *GameService) GetCurrentGame(
	ctx context.Context,
	guildID string,
) (*ent.Game, error) {
	currentGame, err := s.database.Game.Query().
		Where(
			entgame.GuildIDEQ(guildID),
			entgame.StatusEQ(entgame.StatusIN_PROGRESS),
			entgame.EndingAtGT(time.Now()),
		).
		Order(ent.Desc(entgame.FieldCreatedAt)).
		First(ctx)

	if ent.IsNotFound(err) {
		return nil, nil
	}

	return currentGame, err
}

// GetNextGameStart computes when the next scheduled game will start for the guild.
// Returns nil when no previous game exists (next game starts immediately).
func (s *GameService) GetNextGameStart(
	ctx context.Context,
	guildID string,
	guildSettings *ent.Settings,
) (*time.Time, error) {
	lastGames, err := s.database.Game.Query().
		Where(
			entgame.GuildIDEQ(guildID),
			entgame.StatusNEQ(entgame.StatusIN_PROGRESS),
		).
		Order(ent.Desc(entgame.FieldCreatedAt)).
		Limit(1).
		All(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("game: next start: %w", err)
	}

	if len(lastGames) == 0 {
		return nil, nil
	}

	lastGame := lastGames[0]
	baseTime := lastGame.CreatedAt

	if guildSettings.StartAfterFirstGuess {
		firstGuesses, gErr := s.database.Guess.Query().
			Where(guess.GameIDEQ(lastGame.ID)).
			Order(ent.Asc(guess.FieldCreatedAt)).
			Limit(1).
			All(ctx)
		if gErr == nil && len(firstGuesses) > 0 {
			baseTime = firstGuesses[0].CreatedAt
		}
	}

	nextStart := baseTime.Add(time.Duration(guildSettings.Frequency) * time.Minute)
	return &nextStart, nil
}

// checkWord scores a guess against the target word.
// Returns: per-letter meta, guessed(bool), total points, updated game meta.
func (s *GameService) checkWord(
	word string,
	guessWord string,
	state *localUtils.GameMeta,
) (localUtils.GuessMetaSlice, bool, int, *localUtils.GameMeta) {
	meta := make(localUtils.GuessMetaSlice, len(guessWord))
	unmatched := map[rune]int{}   // unmatched word letters
	letterCount := map[rune]int{} // matched count per letter

	wordRunes := []rune(word)
	guessRunes := []rune(guessWord)

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

	return meta, word == guessWord, totalPoints, state
}

// checkCooldown returns cooldown status for a user.
func (s *GameService) checkCooldown(
	userID string,
	guesses []*ent.Guess,
	guildSettings *ent.Settings,
) cooldownResult {
	if len(guesses) == 0 {
		return cooldownResult{}
	}

	// guesses are ordered desc by createdAt from the query
	lastGuess := guesses[0]

	var lastGuessByUser *ent.Guess

	for i := range guesses {
		if guesses[i].UserID == userID {
			lastGuessByUser = guesses[i]
			break
		}
	}

	if lastGuess == nil || lastGuessByUser == nil {
		return cooldownResult{}
	}

	now := time.Now()
	backToBackHit := guildSettings.EnableBackToBackCooldown &&
		lastGuessByUser.CreatedAt.After(
			now.Add(-time.Duration(guildSettings.BackToBackCooldown)*time.Second),
		) &&
		userID == lastGuess.UserID
	cooldownHit := lastGuessByUser.CreatedAt.After(
		now.Add(-time.Duration(guildSettings.Cooldown) * time.Second),
	)

	if backToBackHit || cooldownHit {
		return cooldownResult{
			Hit:       cooldownHit,
			RepeatHit: backToBackHit,
			Result: lastGuessByUser.CreatedAt.Add(
				time.Duration(guildSettings.Cooldown) * time.Second,
			),
			RepeatResult: lastGuessByUser.CreatedAt.Add(
				time.Duration(guildSettings.BackToBackCooldown) * time.Second,
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

	state := &localUtils.GameMeta{
		Keyboard:  map[string]localUtils.GameType{},
		Word:      wordStates,
		Discovery: discovery,
	}
	state.CanHint = localUtils.ComputeCanHint(word, state)

	return state
}

// roundToNearestMinute rounds t to nearest minute.
func roundToNearestMinute(t time.Time) time.Time {
	return t.Round(time.Minute)
}

// UseHint consumes a hint from the pressing user's pool (player first, guild
// fallback) and applies the next tier of hint reveal to the game.
// Returns a human-readable description of what was revealed.
func (s *GameService) UseHint(
	ctx context.Context,
	gameID int,
	userID string,
	guildSettings *ent.Settings,
) (string, error) {
	currentGame, err := s.database.Game.Get(ctx, gameID)
	if err != nil {
		return "", fmt.Errorf("game: hint: find game: %w", err)
	}

	if currentGame.Status != entgame.StatusIN_PROGRESS {
		return "", ErrHintUnavailable
	}

	gameMeta, err := localUtils.ParseGameMeta(json.RawMessage(currentGame.Meta))
	if err != nil {
		return "", fmt.Errorf("game: hint: parse meta: %w", err)
	}

	if !gameMeta.CanHint {
		return "", ErrHintUnavailable
	}

	guesses, err := s.database.Guess.Query().
		Where(guess.GameIDEQ(currentGame.ID)).
		Order(ent.Desc(guess.FieldCreatedAt)).
		All(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return "", fmt.Errorf("game: hint: find guesses: %w", err)
	}

	if len(guesses) >= localStatic.MaxGuesses-1 {
		return "", ErrHintUnavailable
	}

	hintsResult, err := s.hints.GetHints(ctx, guildSettings, userID)
	if err != nil {
		return "", fmt.Errorf("game: hint: get hints: %w", err)
	}

	var (
		leftover     float64
		maxHints     float64
		usedPersonal bool
	)

	if hintsResult.player >= 1 {
		leftover, maxHints, err = s.hints.DeductHintFromPlayer(ctx, userID, 1)
		if err != nil {
			return "", fmt.Errorf("game: hint: deduct player: %w", err)
		}
		usedPersonal = true
	} else if hintsResult.guild >= 1 {
		leftover, maxHints, err = s.hints.DeductHintFromGuild(
			ctx,
			currentGame.GuildID,
			guildSettings,
			1,
		)
		if err != nil {
			return "", fmt.Errorf("game: hint: deduct guild: %w", err)
		}
	} else {
		return "", ErrNoHints
	}

	hintMeta, updatedGameMeta, description, computeErr := s.computeHint(
		currentGame.Word,
		gameMeta,
	)
	if computeErr != nil {
		return "", computeErr
	}

	hintMetaJSON, err := json.Marshal(hintMeta)
	if err != nil {
		return "", fmt.Errorf("game: hint: marshal hint meta: %w", err)
	}

	_, err = s.database.Guess.Create().
		SetUserID(userID).
		SetGameID(currentGame.ID).
		SetWord("").
		SetPoints(0).
		SetMeta(hintMetaJSON).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("game: hint: create guess: %w", err)
	}

	updatedGameMeta.CanHint = localUtils.ComputeCanHint(
		currentGame.Word,
		updatedGameMeta,
	)

	updatedMetaJSON, err := json.Marshal(updatedGameMeta)
	if err != nil {
		return "", fmt.Errorf("game: hint: marshal game meta: %w", err)
	}

	updatedGame, err := s.database.Game.UpdateOneID(currentGame.ID).
		SetMeta(updatedMetaJSON).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("game: hint: update game: %w", err)
	}

	allGuesses, _ := s.database.Guess.Query().
		Where(guess.GameIDEQ(currentGame.ID)).
		Order(ent.Desc(guess.FieldCreatedAt)).
		All(ctx)

	go func() {
		if msgErr := s.message.Create(
			context.Background(),
			updatedGame,
			allGuesses,
			false,
		); msgErr != nil {
			utils.Logger.Warnw("game: hint: recreate message failed",
				"error", msgErr,
				"guildID", currentGame.GuildID,
				"gameID", currentGame.ID,
			)
		}
	}()

	poolKind := "personal"
	if !usedPersonal {
		poolKind = "server"
	}

	return fmt.Sprintf(
		"%s\nUsed **1** %s hint. You have **%s/%s** %s hints remaining.",
		description,
		poolKind,
		strconv.FormatFloat(leftover, 'f', -1, 64),
		strconv.FormatFloat(maxHints, 'f', -1, 64),
		poolKind,
	), nil
}

// computeHint applies hints in priority order and returns the row meta, updated
// state, and a human-readable description. Caller must verify state.CanHint.
func (s *GameService) computeHint(
	word string,
	state *localUtils.GameMeta,
) (localUtils.GuessMetaSlice, *localUtils.GameMeta, string, error) {
	meta := make(localUtils.GuessMetaSlice, localStatic.WordLength)
	description := ""

	wordRunes := []rune(word)
	wordCount := localUtils.WordLetterCount(word)

	nonCorrect := 0
	for _, ws := range state.Word {
		if ws.Type != localUtils.GameTypeCorrect {
			nonCorrect++
		}
	}

	hintDone := false

	// Tier 1: upgrade the first ALMOST letter to a correct position.
	if nonCorrect >= 2 {
		for i, ws := range state.Word {
			if ws.Type == localUtils.GameTypeCorrect {
				continue
			}
			// Eligible when keyboard is ALMOST, or CORRECT with unplaced occurrences.
			kb := state.Keyboard[ws.Letter]
			hasUnplaced := kb == localUtils.GameTypeCorrect &&
				state.Discovery.Almost[ws.Letter] > state.Discovery.Correct[ws.Letter]
			if kb != localUtils.GameTypeAlmost && !hasUnplaced {
				continue
			}

			letter := ws.Letter
			state.Word[i].Type = localUtils.GameTypeCorrect
			state.Keyboard[letter] = localUtils.GameTypeCorrect
			state.Discovery.Correct[letter]++
			description = fmt.Sprintf(
				"Solved position **%d** with **%s**",
				i+1, strings.ToUpper(letter),
			)

			hintDone = true
			break
		}
	}

	// Tier 2: reveal a letter occurrence not yet found (supports duplicate letters).
	if !hintDone {
		seen := map[rune]bool{}
		for _, r := range wordRunes {
			if seen[r] {
				continue
			}
			seen[r] = true

			letter := string(r)
			discovered := state.Discovery.Almost[letter]
			if wordCount[r] <= discovered {
				continue
			}

			newAlmost := discovered + 1
			state.Discovery.Almost[letter] = newAlmost
			if state.Keyboard[letter] != localUtils.GameTypeCorrect {
				state.Keyboard[letter] = localUtils.GameTypeAlmost
			}

			description = fmt.Sprintf(
				"Revealed letter **%s**",
				strings.ToUpper(letter),
			)
			hintDone = true
			break
		}
	}

	// Tier 3: all letters discovered — solve the first unsolved position.
	if !hintDone {
		if nonCorrect < 2 {
			return nil, nil, "", ErrHintUnavailable
		}
		for i, ws := range state.Word {
			if ws.Type != localUtils.GameTypeCorrect {
				letter := ws.Letter
				state.Word[i].Type = localUtils.GameTypeCorrect
				state.Keyboard[letter] = localUtils.GameTypeCorrect
				state.Discovery.Correct[letter]++

				description = fmt.Sprintf(
					"Solved position **%d** with **%s**",
					i+1, strings.ToUpper(letter),
				)

				break
			}
		}
	}

	// Build synthetic guess row: CORRECT positions first
	for i, ws := range state.Word {
		if ws.Type == localUtils.GameTypeCorrect {
			meta[i] = localUtils.GuessMeta{
				Type:   localUtils.GameTypeCorrect,
				Letter: ws.Letter,
			}
		}
	}

	// Collect ALMOST letters with multiplicity, deduped by unique letter.
	var almostLetters []string
	seenLetters := map[string]bool{}
	for _, r := range wordRunes {
		letter := string(r)
		if seenLetters[letter] {
			continue
		}
		seenLetters[letter] = true

		var visible int
		switch state.Keyboard[letter] {
		case localUtils.GameTypeAlmost:
			visible = state.Discovery.Almost[letter]
		case localUtils.GameTypeCorrect:
			visible = state.Discovery.Almost[letter] - state.Discovery.Correct[letter]
		}
		for k := 0; k < visible; k++ {
			almostLetters = append(almostLetters, letter)
		}
	}

	// Fill remaining positions with ALMOST letters (must be wrong-spot) or blank
	for i := range meta {
		if meta[i].Type == localUtils.GameTypeCorrect {
			continue
		}
		placed := false
		for j, letter := range almostLetters {
			if string(wordRunes[i]) != letter {
				meta[i] = localUtils.GuessMeta{
					Type:   localUtils.GameTypeAlmost,
					Letter: letter,
				}
				almostLetters = append(
					almostLetters[:j],
					almostLetters[j+1:]...)
				placed = true
				break
			}
		}
		if !placed {
			meta[i] = localUtils.GuessMeta{
				Type:   localUtils.GameTypeDefault,
				Letter: localStatic.EmojiLetterBlank,
			}
		}
	}

	return meta, state, description, nil
}

func (s *GameService) CountByGuildIDs(
	ctx context.Context,
	guildIDs []string,
) (int, int, error) {
	gameCount, err := s.database.Game.Query().
		Where(entgame.GuildIDIn(guildIDs...)).
		Count(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, 0, fmt.Errorf("game: count by guild ids: games: %w", err)
	}

	// Get game IDs for the guilds
	gameIDs, err := s.database.Game.Query().
		Where(entgame.GuildIDIn(guildIDs...)).
		IDs(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, 0, fmt.Errorf("game: count by guild ids: game ids: %w", err)
	}

	guessCount := 0
	if len(gameIDs) > 0 {
		guessCount, err = s.database.Guess.Query().
			Where(guess.GameIDIn(gameIDs...)).
			Count(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return 0, 0, fmt.Errorf("game: count by guild ids: guesses: %w", err)
		}
	}

	return gameCount, guessCount, nil
}

func (s *GameService) DeleteByGuildIDs(
	ctx context.Context,
	guildIDs []string,
) (int, int, error) {
	// Get game IDs first so we can delete guesses
	gameIDs, err := s.database.Game.Query().
		Where(entgame.GuildIDIn(guildIDs...)).
		IDs(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, 0, fmt.Errorf("game: delete by guild ids: game ids: %w", err)
	}

	guessCount := 0
	if len(gameIDs) > 0 {
		guessCount, err = s.database.Guess.Delete().
			Where(guess.GameIDIn(gameIDs...)).
			Exec(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return 0, 0, fmt.Errorf("game: delete by guild ids: guesses: %w", err)
		}
	}

	gameCount, err := s.database.Game.Delete().
		Where(entgame.GuildIDIn(guildIDs...)).
		Exec(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, 0, fmt.Errorf("game: delete by guild ids: games: %w", err)
	}

	return gameCount, guessCount, nil
}

type GuildIDRow struct {
	GuildID string `json:"guildId"`
}

func (s *GameService) FindAllGuildIDs(
	ctx context.Context,
) ([]GuildIDRow, error) {
	guildIDs, err := s.database.Game.Query().
		Where(entgame.StatusEQ(entgame.StatusIN_PROGRESS)).
		GroupBy(entgame.FieldGuildID).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("game: find distinct guild ids: %w", err)
	}

	rows := make([]GuildIDRow, len(guildIDs))
	for i, id := range guildIDs {
		rows[i] = GuildIDRow{GuildID: id}
	}

	return rows, nil
}

// suppress unused import warning - errors package used in original but not needed now
var _ = errors.New
