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

	b := container.Get(sharedStatic.DiBot).(*disgoplus.Bot)

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

	return s.createNewGame(ctx, guildID, schedule, word)
}

func (s *GameService) createNewGame(
	ctx context.Context,
	guildID string,
	schedule bool,
	word string,
) (bool, error) {
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
		endingAt = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		endingAt = roundToNearestMinute(
			time.Now().
				Add(time.Duration(guildSettings.TimeLimit) * time.Minute),
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

		if chanErr := s.handleChannelError(
			ctx, err, guildSettings, newGame,
		); chanErr != nil {
			return false, chanErr
		}
	}

	return true, nil
}

func (s *GameService) handleChannelError(
	ctx context.Context,
	msgErr error,
	guildSettings *ent.Settings,
	game *ent.Game,
) error {
	if !strings.Contains(msgErr.Error(), "404 Not Found") &&
		!strings.Contains(msgErr.Error(), "403 Forbidden") {
		return nil
	}

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
		game.ID,
		entgame.StatusFAILED,
	); endErr != nil {
		utils.Logger.Warnw("game: start: end current failed",
			"error", endErr,
			"guildID", game.GuildID,
			"gameID", game.ID,
		)
	}

	reason := "Forbidden"
	if strings.Contains(msgErr.Error(), "404 Not Found") {
		reason = "Not found"
	}

	return fmt.Errorf("game: start: %s: %w", reason, msgErr)
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

	if s.isDuplicate(guesses, word) {
		utils.LogIfErr(
			utils.Logger,
			"reaction-add",
			s.client.Rest.AddReaction(message.ChannelID, message.ID, "❌"),
		)

		return nil
	}

	cooldown := s.checkCooldown(
		message.Author.ID.String(),
		guesses,
		guildSettings,
	)
	if cooldown.Hit || cooldown.RepeatHit {
		return s.handleCooldown(message, cooldown)
	}

	return s.processGuess(
		ctx,
		guildID,
		word,
		message,
		currentGame,
		guesses,
		guildSettings,
	)
}

func (s *GameService) processGuess(
	ctx context.Context,
	guildID string,
	word string,
	message discord.Message,
	currentGame *ent.Game,
	guesses []*ent.Guess,
	guildSettings *ent.Settings,
) error {
	gameMeta, err := localUtils.ParseGameMeta(currentGame.Meta)
	if err != nil {
		return fmt.Errorf("game: guess: parse meta: %w", err)
	}

	guessMeta, guessed, points, updatedGameMeta := s.checkWord(
		currentGame.Word,
		word,
		gameMeta,
	)

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

	updatedGame, newStatus, err := s.updateGameAfterGuess(
		ctx, currentGame, guesses, guessed, updatedGameMeta, guildSettings,
	)
	if err != nil {
		return err
	}

	s.finalizeGuess(
		ctx, guildID, message, currentGame, updatedGame,
		createdGuess, newStatus, guessed, guildSettings,
	)

	return nil
}

func (s *GameService) updateGameAfterGuess(
	ctx context.Context,
	currentGame *ent.Game,
	guesses []*ent.Guess,
	guessed bool,
	updatedGameMeta *localUtils.GameMeta,
	guildSettings *ent.Settings,
) (*ent.Game, entgame.Status, error) {
	newStatus := currentGame.Status
	if guessed {
		newStatus = entgame.StatusCOMPLETED
	} else if len(guesses)+1 >= localStatic.MaxGuesses {
		newStatus = entgame.StatusFAILED
	}

	updatedGameMeta.CanHint = localUtils.ComputeCanHint(
		currentGame.Word,
		updatedGameMeta,
	)

	updatedMetaJSON, err := json.Marshal(updatedGameMeta)
	if err != nil {
		return nil, newStatus, fmt.Errorf(
			"game: guess: marshal game meta: %w",
			err,
		)
	}

	upd := s.database.Game.UpdateOneID(currentGame.ID).
		SetStatus(newStatus).
		SetMeta(updatedMetaJSON)

	if guildSettings.StartAfterFirstGuess && len(guesses) == 0 {
		realEndingAt := roundToNearestMinute(
			time.Now().
				Add(time.Duration(guildSettings.TimeLimit) * time.Minute),
		)
		upd = upd.SetEndingAt(realEndingAt)
	}

	updatedGame, err := upd.Save(ctx)
	if err != nil {
		return nil, newStatus, fmt.Errorf("game: guess: update game: %w", err)
	}

	return updatedGame, newStatus, nil
}

func (s *GameService) finalizeGuess(
	ctx context.Context,
	guildID string,
	message discord.Message,
	currentGame *ent.Game,
	updatedGame *ent.Game,
	createdGuess *ent.Guess,
	newStatus entgame.Status,
	guessed bool,
	guildSettings *ent.Settings,
) {
	reactionEmoji := "✅"
	if guessed {
		reactionEmoji = "🎉"
	}

	utils.LogIfErr(utils.Logger, "reaction-add",
		s.client.Rest.AddReaction(message.ChannelID, message.ID, reactionEmoji),
	)

	allGuesses, _ := s.database.Guess.Query().
		Where(guess.GameIDEQ(currentGame.ID)).
		Order(ent.Desc(guess.FieldCreatedAt)).
		All(ctx)

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

	if newStatus != entgame.StatusCOMPLETED &&
		guildSettings.InformCooldownAfterGuess {
		go s.sendCooldownInfo(message, createdGuess, guildSettings)
	}

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

	if newStatus != entgame.StatusIN_PROGRESS {
		s.handleTerminalStatus(guildID, updatedGame, newStatus, guildSettings)
	}
}

func (s *GameService) sendCooldownInfo(
	message discord.Message,
	createdGuess *ent.Guess,
	guildSettings *ent.Settings,
) {
	backToBackPart := ""
	if guildSettings.EnableBackToBackCooldown {
		backToBackPart = fmt.Sprintf(
			"<t:%d:R> on your own or ",
			createdGuess.CreatedAt.Add(
				time.Duration(guildSettings.BackToBackCooldown)*time.Second,
			).Unix(),
		)
	}

	afterPart := ""
	if guildSettings.EnableBackToBackCooldown {
		afterPart = " after a guess from another player"
	}

	msg := fmt.Sprintf(
		"You are now on a cooldown. You can guess again %s<t:%d:R>%s.",
		backToBackPart,
		createdGuess.CreatedAt.Add(
			time.Duration(guildSettings.Cooldown)*time.Second,
		).Unix(),
		afterPart,
	)
	msgID := message.ID
	_, sendErr := s.client.Rest.CreateMessage(
		message.ChannelID,
		discord.MessageCreate{
			Content: msg,
			MessageReference: &discord.MessageReference{
				MessageID: &msgID,
			},
		},
	)
	utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
}

func (s *GameService) handleTerminalStatus(
	guildID string,
	updatedGame *ent.Game,
	newStatus entgame.Status,
	guildSettings *ent.Settings,
) {
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

	if !guildSettings.AutoStart {
		return
	}

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

	if msgErr := s.message.Create(
		ctx,
		endedGame,
		guesses,
		false,
	); msgErr != nil {
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

	if err != nil {
		return nil, fmt.Errorf("game: get current game: %w", err)
	}

	return currentGame, nil
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

	nextStart := baseTime.Add(
		time.Duration(guildSettings.Frequency) * time.Minute,
	)

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
	unmatched := map[rune]int{}
	letterCount := map[rune]int{}

	wordRunes := []rune(word)
	guessRunes := []rune(guessWord)

	checkWordPass1(wordRunes, guessRunes, state, meta, unmatched, letterCount)
	checkWordPass2(wordRunes, guessRunes, state, meta, unmatched, letterCount)

	totalPoints := 0
	for _, m := range meta {
		totalPoints += m.Points
	}

	return meta, word == guessWord, totalPoints, state
}

func checkWordPass1(
	wordRunes, guessRunes []rune,
	state *localUtils.GameMeta,
	meta localUtils.GuessMetaSlice,
	unmatched, letterCount map[rune]int,
) {
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
}

func checkWordPass2(
	wordRunes, guessRunes []rune,
	state *localUtils.GameMeta,
	meta localUtils.GuessMetaSlice,
	unmatched, letterCount map[rune]int,
) {
	for i, element := range wordRunes {
		if i >= len(guessRunes) {
			break
		}

		letter := guessRunes[i]
		if letter == element {
			continue
		}

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
			now.Add(
				-time.Duration(guildSettings.BackToBackCooldown)*time.Second,
			),
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

func (s *GameService) isDuplicate(guesses []*ent.Guess, word string) bool {
	if os.Getenv("ENV") != "production" {
		return false
	}

	for _, g := range guesses {
		if g.Word == word {
			return true
		}
	}

	return false
}

func (s *GameService) handleCooldown(
	message discord.Message,
	cooldown cooldownResult,
) error {
	utils.LogIfErr(utils.Logger, "reaction-add",
		s.client.Rest.AddReaction(message.ChannelID, message.ID, "🕒"),
	)

	suffix := fmt.Sprintf(
		"you can guess again <t:%d:R>",
		cooldown.Result.Unix(),
	)

	switch {
	case cooldown.Hit && cooldown.RepeatHit:
		suffix = fmt.Sprintf(
			"you can guess again <t:%d:R> on your own or <t:%d:R> after a guess from another player.",
			cooldown.RepeatResult.Unix(),
			cooldown.Result.Unix(),
		)
	case !cooldown.Hit && cooldown.RepeatHit:
		suffix = fmt.Sprintf(
			"you can guess again <t:%d:R> or immediately after a guess from another player.",
			cooldown.RepeatResult.Unix(),
		)
	}

	msgID := message.ID
	_, sendErr := s.client.Rest.CreateMessage(
		message.ChannelID,
		discord.MessageCreate{
			Content: fmt.Sprintf("You're on a cooldown, %s", suffix),
			MessageReference: &discord.MessageReference{
				MessageID: &msgID,
			},
		},
	)
	utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)

	return nil
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

	gameMeta, err := localUtils.ParseGameMeta(currentGame.Meta)
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

	switch {
	case hintsResult.player >= 1:
		leftover, maxHints, err = s.hints.DeductHintFromPlayer(ctx, userID, 1)
		if err != nil {
			return "", fmt.Errorf("game: hint: deduct player: %w", err)
		}

		usedPersonal = true
	case hintsResult.guild >= 1:
		leftover, maxHints, err = s.hints.DeductHintFromGuild(
			ctx,
			currentGame.GuildID,
			guildSettings,
			1,
		)
		if err != nil {
			return "", fmt.Errorf("game: hint: deduct guild: %w", err)
		}
	default:
		return "", ErrNoHints
	}

	hintMeta, updatedGameMeta, description, computeErr := s.computeHint(
		currentGame.Word,
		gameMeta,
	)
	if computeErr != nil {
		return "", computeErr
	}

	if err = s.persistHintGuess(
		ctx, currentGame, userID, hintMeta, updatedGameMeta,
	); err != nil {
		return "", err
	}

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

func (s *GameService) persistHintGuess(
	ctx context.Context,
	currentGame *ent.Game,
	userID string,
	hintMeta localUtils.GuessMetaSlice,
	updatedGameMeta *localUtils.GameMeta,
) error {
	hintMetaJSON, err := json.Marshal(hintMeta)
	if err != nil {
		return fmt.Errorf("game: hint: marshal hint meta: %w", err)
	}

	_, err = s.database.Guess.Create().
		SetUserID(userID).
		SetGameID(currentGame.ID).
		SetWord("").
		SetPoints(0).
		SetMeta(hintMetaJSON).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("game: hint: create guess: %w", err)
	}

	updatedGameMeta.CanHint = localUtils.ComputeCanHint(
		currentGame.Word,
		updatedGameMeta,
	)

	updatedMetaJSON, err := json.Marshal(updatedGameMeta)
	if err != nil {
		return fmt.Errorf("game: hint: marshal game meta: %w", err)
	}

	updatedGame, err := s.database.Game.UpdateOneID(currentGame.ID).
		SetMeta(updatedMetaJSON).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("game: hint: update game: %w", err)
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

	return nil
}

// computeHint applies hints in priority order and returns the row meta, updated
// state, and a human-readable description. Caller must verify state.CanHint.
func (s *GameService) computeHint(
	word string,
	state *localUtils.GameMeta,
) (localUtils.GuessMetaSlice, *localUtils.GameMeta, string, error) {
	meta := make(localUtils.GuessMetaSlice, localStatic.WordLength)

	wordRunes := []rune(word)
	wordCount := localUtils.WordLetterCount(word)

	nonCorrect := countNonCorrect(state)

	description, hintDone := applyHintTier1(state, nonCorrect)

	if !hintDone {
		description, hintDone = applyHintTier2(state, wordRunes, wordCount)
	}

	if !hintDone {
		var err error
		if description, err = applyHintTier3(state, nonCorrect); err != nil {
			return nil, nil, "", err
		}
	}

	buildHintRow(meta, state, wordRunes)

	return meta, state, description, nil
}

func countNonCorrect(state *localUtils.GameMeta) int {
	nonCorrect := 0

	for _, ws := range state.Word {
		if ws.Type != localUtils.GameTypeCorrect {
			nonCorrect++
		}
	}

	return nonCorrect
}

func applyHintTier1(
	state *localUtils.GameMeta,
	nonCorrect int,
) (string, bool) {
	if nonCorrect < 2 {
		return "", false
	}

	for i, ws := range state.Word {
		if ws.Type == localUtils.GameTypeCorrect {
			continue
		}

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

		return fmt.Sprintf(
			"Solved position **%d** with **%s**",
			i+1, strings.ToUpper(letter),
		), true
	}

	return "", false
}

func applyHintTier2(
	state *localUtils.GameMeta,
	wordRunes []rune,
	wordCount map[rune]int,
) (string, bool) {
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

		state.Discovery.Almost[letter] = discovered + 1
		if state.Keyboard[letter] != localUtils.GameTypeCorrect {
			state.Keyboard[letter] = localUtils.GameTypeAlmost
		}

		return fmt.Sprintf(
			"Revealed letter **%s**",
			strings.ToUpper(letter),
		), true
	}

	return "", false
}

func applyHintTier3(
	state *localUtils.GameMeta,
	nonCorrect int,
) (string, error) {
	if nonCorrect < 2 {
		return "", ErrHintUnavailable
	}

	for i, ws := range state.Word {
		if ws.Type != localUtils.GameTypeCorrect {
			letter := ws.Letter
			state.Word[i].Type = localUtils.GameTypeCorrect
			state.Keyboard[letter] = localUtils.GameTypeCorrect
			state.Discovery.Correct[letter]++

			return fmt.Sprintf(
				"Solved position **%d** with **%s**",
				i+1, strings.ToUpper(letter),
			), nil
		}
	}

	return "", nil
}

func buildHintRow(
	meta localUtils.GuessMetaSlice,
	state *localUtils.GameMeta,
	wordRunes []rune,
) {
	for i, ws := range state.Word {
		if ws.Type == localUtils.GameTypeCorrect {
			meta[i] = localUtils.GuessMeta{
				Type:   localUtils.GameTypeCorrect,
				Letter: ws.Letter,
			}
		}
	}

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
			return 0, 0, fmt.Errorf(
				"game: count by guild ids: guesses: %w",
				err,
			)
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
			return 0, 0, fmt.Errorf(
				"game: delete by guild ids: guesses: %w",
				err,
			)
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
