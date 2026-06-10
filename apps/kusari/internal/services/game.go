package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/kusari/internal/ent/game"
	"jurien.dev/yugen/kusari/internal/ent/history"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

var (
	ErrNoChannelIDConfigured = errors.New("no channel id configured")
	ErrAuthorIsBot           = errors.New("author is bot")

	firstLetterRegex = regexp.MustCompile("^[A-Za-z!]+$")
	lastLetterRegex  = regexp.MustCompile("^[A-Za-z]+$")
)

type GameService struct {
	client     *disgoplus.Bot
	cfg        *config.Config
	database   *ent.Client
	settings   *SettingsService
	saves      *SavesService
	points     *PointsService
	dictionary *DictionaryService
}

func CreateGameService(container *di.Container) *GameService {
	utils.Logger.Info("Creating Game Service")

	return &GameService{
		client:     container.Get(static.DiBot).(*disgoplus.Bot),
		cfg:        container.Get(static.DiConfig).(*config.Config),
		database:   container.Get(static.DiDatabase).(*ent.Client),
		settings:   container.Get(static.DiSettings).(*SettingsService),
		saves:      container.Get(localStatic.DiSaves).(*SavesService),
		points:     container.Get(localStatic.DiPoints).(*PointsService),
		dictionary: container.Get(localStatic.DiDictionary).(*DictionaryService),
	}
}

func (s *GameService) Start(
	ctx context.Context,
	guildID string,
	gameType game.Type,
	word string,
	recreate bool,
) (g *ent.Game, started bool, err error) {
	utils.Logger.Infof("Trying to start a game for %s", guildID)

	currentGame, exists, err := s.GetCurrentGame(ctx, guildID)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Errorw(
			"game: start: get current game failed",
			"error", err, "guildID", guildID,
		)

		return g, started, err
	}

	channelSnowflake, channel, err := s.resolveGameChannel(ctx, guildID)
	if err != nil {
		return g, started, err
	}

	if exists && !recreate {
		return g, started, nil
	}

	s.maybeEndCurrentGame(ctx, exists, recreate, currentGame, gameType, guildID)

	started = true

	g, word, err = s.createGameAndHistory(
		ctx, guildID, gameType, word, channelSnowflake,
	)
	if err != nil {
		return g, started, err
	}

	s.sendGameStartMessage(channel, channelSnowflake, word)

	return g, started, nil
}

func (s *GameService) resolveGameChannel(
	ctx context.Context,
	guildID string,
) (channelSnowflake snowflake.ID, channel discord.Channel, err error) {
	settings, sErr := s.settings.GetByGuildID(ctx, guildID)
	if sErr != nil {
		utils.Logger.Errorw(
			"game: start: get settings failed",
			"error", sErr, "guildID", guildID,
		)

		return channelSnowflake, channel,
			fmt.Errorf("game: start: get settings: %w", sErr)
	}

	channelID := settings.ChannelID
	if channelID == nil {
		utils.Logger.Errorw(
			"game: start: no channel id configured",
			"error", ErrNoChannelIDConfigured, "guildID", guildID,
		)

		return channelSnowflake, channel, ErrNoChannelIDConfigured
	}

	channelSnowflake, parseErr := snowflake.Parse(*channelID)
	if parseErr != nil {
		return channelSnowflake, channel, fmt.Errorf(
			"game: start: parse channel id: %w", parseErr,
		)
	}

	channel, err = s.client.Client().Rest.GetChannel(channelSnowflake)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: get channel failed",
			"error", err, "guildID", guildID, "channelID", *channelID,
		)

		return channelSnowflake, channel,
			fmt.Errorf("game: start: get channel: %w", err)
	}

	return channelSnowflake, channel, nil
}

func (s *GameService) maybeEndCurrentGame(
	ctx context.Context,
	exists bool,
	recreate bool,
	currentGame *ent.Game,
	gameType game.Type,
	guildID string,
) {
	if !exists {
		return
	}

	if !recreate && currentGame.Type == gameType {
		return
	}

	_, endErr := s.End(ctx, currentGame.ID, game.StatusFAILED)
	if endErr != nil {
		utils.Logger.Warnw(
			"game: start: end current game failed",
			"error", endErr, "guildID", guildID, "gameID", currentGame.ID,
		)
	}
}

func (s *GameService) createGameAndHistory(
	ctx context.Context,
	guildID string,
	gameType game.Type,
	word string,
	channelSnowflake snowflake.ID,
) (g *ent.Game, resolvedWord string, err error) {
	g, err = s.database.Game.Create().
		SetGuildID(guildID).
		SetType(gameType).
		Save(ctx)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: create game failed",
			"error", err, "guildID", guildID,
		)

		return g, word, fmt.Errorf("game: start: create game: %w", err)
	}

	if guildSnowflake, parseErr := snowflake.Parse(guildID); parseErr == nil {
		utils.ActiveGames.Register(guildSnowflake, channelSnowflake)
		s.client.Client().Caches.RemoveMessagesByChannelID(channelSnowflake)
	}

	self, _ := s.client.Client().Caches.SelfUser()

	if len(word) == 0 {
		word = s.getRandomLetter()
	}

	_, err = s.database.History.Create().
		SetUserID(self.ID.String()).
		SetGameID(g.ID).
		SetWord(word).
		Save(ctx)
	if err != nil {
		return g, word,
			fmt.Errorf("game: start: create history: %w", err)
	}

	return g, word, nil
}

func (s *GameService) sendGameStartMessage(
	channel discord.Channel,
	channelSnowflake snowflake.ID,
	word string,
) {
	guildChannel, isGuildChannel := channel.(discord.GuildMessageChannel)
	if !isGuildChannel {
		return
	}

	chType := guildChannel.Type()
	if chType != discord.ChannelTypeGuildText &&
		chType != discord.ChannelTypeGuildPublicThread &&
		chType != discord.ChannelTypeGuildPrivateThread {
		return
	}

	go func() {
		_, createErr := s.client.Client().Rest.CreateMessage(
			channelSnowflake,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					"**A new game has started!**\n"+
						"The first letter is: **%s**",
					strings.ToUpper(string(word[len(word)-1])),
				),
			},
		)
		utils.LogIfErr(utils.Logger, "create-message", createErr)
	}()
}

func (s *GameService) End(
	ctx context.Context,
	gameID int,
	status game.Status,
) (g *ent.Game, err error) {
	g, err = s.database.Game.UpdateOneID(gameID).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return g, fmt.Errorf("game: end: update game: %w", err)
	}

	if guildSnowflake, parseErr := snowflake.Parse(g.GuildID); parseErr == nil {
		utils.ActiveGames.Unregister(guildSnowflake)
	}

	if _, delErr := s.database.History.Delete().
		Where(history.GameIDEQ(gameID)).
		Exec(ctx); delErr != nil {
		utils.Logger.Warnw(
			"game: end: delete history failed",
			"error", delErr,
			"gameID", gameID,
		)
	}

	return g, nil
}

func (s *GameService) ParseWord(
	message discord.Message,
) (word string, err error) {
	if message.Author.Bot {
		err = ErrAuthorIsBot
		return word, err
	}

	words := strings.Fields(message.Content)
	if len(words) == 0 || len(words) > 1 {
		return word, err
	}

	word = words[0]

	if !firstLetterRegex.MatchString(string(word[0])) {
		word = ""
		return word, err
	}

	if !lastLetterRegex.MatchString(string(word[len(word)-1])) {
		word = ""
		return word, err
	}

	if len(word) > 0 && string(word[0]) == "!" {
		word = word[1:]
	}

	word = strings.ToLower(word)

	return word, err
}

func (s *GameService) AddWord(
	ctx context.Context,
	guildID string,
	word string,
	message discord.Message,
	settings *ent.Settings,
) {
	if len(word) == 0 {
		return
	}

	g, exists, err := s.GetCurrentGame(ctx, guildID)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: get current game failed",
			"error", err, "guildID", guildID,
		)

		return
	}

	if !exists {
		return
	}

	h, _, err := s.GetLastHistory(ctx, g)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: get last history failed",
			"error", err, "guildID", guildID, "gameID", g.ID,
		)

		return
	}

	if h == nil {
		utils.Logger.Debugw(
			"History is nil", "guildID", guildID, "gameID", g.ID,
		)
	}

	if h != nil && s.cfg.Env != "development" &&
		message.Author.ID.String() == h.UserID {
		s.handleSameUser(message)
		return
	}

	lastLetter := h.Word[len(h.Word)-1]

	wordExists, err := s.dictionary.Check(ctx, word)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: dictionary check failed",
			"error", err, "guildID", guildID, "gameID", g.ID,
			"userID", message.Author.ID, "word", word,
		)

		return
	}

	if !wordExists || word[0] != lastLetter {
		s.handleInvalidWord(
			ctx, guildID, word, lastLetter, wordExists, message, settings, g,
		)

		return
	}

	s.handleValidWord(ctx, guildID, word, message, settings, g)
}

func (s *GameService) handleSameUser(message discord.Message) {
	client := s.client.Client()

	utils.LogIfErr(
		utils.Logger,
		"message-reaction-add",
		client.Rest.AddReaction(message.ChannelID, message.ID, "🕒"),
	)

	go func() {
		_, sendErr := client.Rest.CreateMessage(
			message.ChannelID,
			discord.MessageCreate{
				Content: "Sorry, but you can't add a word twice in a row!" +
					" Please wait for another player to add a word.",
				MessageReference: &discord.MessageReference{
					MessageID: &message.ID,
					ChannelID: &message.ChannelID,
					GuildID:   message.GuildID,
				},
			},
		)
		utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
	}()
}

func (s *GameService) handleInvalidWord(
	ctx context.Context,
	guildID string,
	word string,
	lastLetter byte,
	_ bool,
	message discord.Message,
	settings *ent.Settings,
	g *ent.Game,
) {
	client := s.client.Client()

	failReason := fmt.Sprintf(
		`Sorry, I couldn't find "**%s**" in the `+
			`[English dictionary](https://en.wiktionary.org/wiki/%s), try again!`,
		word, word,
	)
	if word[0] != lastLetter {
		failReason = fmt.Sprintf(
			"The word %s does not start with the letter **%s**",
			word, string(lastLetter),
		)
	}

	utils.LogIfErr(
		utils.Logger, "message-reaction-add",
		client.Rest.AddReaction(message.ChannelID, message.ID, "❌"),
	)

	saves, savesErr := s.saves.GetSaves(
		ctx, settings, message.Author.ID.String(),
	)
	if savesErr != nil {
		utils.Logger.Errorw(
			"game: add word: get saves failed",
			"error", savesErr, "guildID", guildID, "gameID", g.ID,
			"userID", message.Author.ID,
		)

		return
	}

	if saves.player >= 1 {
		s.handlePlayerSave(ctx, guildID, failReason, message, g)
		return
	}

	if saves.guild >= 1 {
		s.handleGuildSave(ctx, guildID, failReason, message, settings, g)
		return
	}

	s.handleGameEnd(ctx, guildID, failReason, message, settings, g)
}

func (s *GameService) handlePlayerSave(
	ctx context.Context,
	guildID string,
	failReason string,
	message discord.Message,
	g *ent.Game,
) {
	client := s.client.Client()

	leftoverSaves, maxSaves, playerSaveErr := s.saves.DeductSaveFromPlayer(
		ctx, message.Author.ID.String(), 1,
	)
	if playerSaveErr != nil {
		utils.Logger.Errorw(
			"game: add word: deduct player save failed",
			"error", playerSaveErr, "guildID", guildID, "gameID", g.ID,
			"userID", message.Author.ID,
		)

		return
	}

	go func() {
		_, sendErr := client.Rest.CreateMessage(
			message.ChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					"%s\nUsed **1 of your own** saves,"+
						" You have **%s/%s** saves left.",
					failReason,
					strconv.FormatFloat(leftoverSaves, 'f', -1, 64),
					strconv.FormatFloat(maxSaves, 'f', -1, 64),
				),
				MessageReference: &discord.MessageReference{
					MessageID: &message.ID,
					ChannelID: &message.ChannelID,
					GuildID:   message.GuildID,
				},
			},
		)
		utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
	}()
}

func (s *GameService) handleGuildSave(
	ctx context.Context,
	guildID string,
	failReason string,
	message discord.Message,
	settings *ent.Settings,
	_ *ent.Game,
) {
	client := s.client.Client()

	leftoverSaves, maxSaves, guildSaveErr := s.saves.DeductSaveFromGuild(
		ctx, message.GuildID.String(), settings, 1,
	)
	if guildSaveErr != nil {
		utils.Logger.Errorw(
			"game: deduct guild save failed",
			"error", guildSaveErr, "guildID", guildID,
		)

		return
	}

	go func() {
		_, sendErr := client.Rest.CreateMessage(
			message.ChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					"%s\nUsed **1 server** save,"+
						" There are **%s/%s** server saves left.",
					failReason,
					strconv.FormatFloat(leftoverSaves, 'f', -1, 64),
					strconv.FormatFloat(maxSaves, 'f', -1, 64),
				),
				MessageReference: &discord.MessageReference{
					MessageID: &message.ID,
					ChannelID: &message.ChannelID,
					GuildID:   message.GuildID,
				},
			},
		)
		utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
	}()
}

func (s *GameService) handleGameEnd(
	ctx context.Context,
	guildID string,
	failReason string,
	message discord.Message,
	settings *ent.Settings,
	g *ent.Game,
) {
	client := s.client.Client()

	failCount, countErr := s.getCount(ctx, g.ID)
	if countErr != nil {
		utils.Logger.Errorw(
			"game: add word: get count failed",
			"error", countErr, "guildID", guildID, "gameID", g.ID,
		)

		return
	}

	isHighscore, _ := s.checkStreak(ctx, settings, g, failCount)

	highScoreText := ""
	if isHighscore {
		highScoreText = "\n**A new highscore has been set! 🎉**"
	}

	pointsRemoved := failCount / 10

	go func() {
		utils.LogIfErr(
			utils.Logger, "remove-game-points",
			s.points.RemoveGamePoints(
				ctx, guildID, message.Author.ID.String(), pointsRemoved,
			),
		)
	}()

	if pointsRemoved == 0 {
		pointsRemoved = 1
	}

	pointText := "Points have"
	if pointsRemoved == 1 {
		pointText = "Point has"
	}

	pointsRemovedText := fmt.Sprintf(
		"\n\n**%d %s been removed from your account.**",
		pointsRemoved, pointText,
	)

	go func() {
		_, sendErr := client.Rest.CreateMessage(
			message.ChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					"%s\n**The game has ended on a streak of %d!**%s%s\n\n"+
						"**Want to save the game?** Make sure to **/vote**"+
						" for Kusari and earn yourself saves to save the game!",
					failReason, failCount, highScoreText, pointsRemovedText,
				),
				MessageReference: &discord.MessageReference{
					MessageID: &message.ID,
					ChannelID: &message.ChannelID,
					GuildID:   message.GuildID,
				},
			},
		)
		utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
	}()

	if _, _, startErr := s.Start(
		ctx, guildID, game.TypeNORMAL, "", true,
	); startErr != nil {
		utils.Logger.Warnw(
			"game: add word: restart failed",
			"error", startErr, "guildID", guildID, "gameID", g.ID,
		)
	}
}

func (s *GameService) handleValidWord(
	ctx context.Context,
	guildID string,
	word string,
	message discord.Message,
	settings *ent.Settings,
	g *ent.Game,
) {
	if !s.passesWordGuards(ctx, guildID, word, message, settings, g) {
		return
	}

	s.recordWordAndReact(ctx, guildID, word, message, settings, g)
}

func (s *GameService) passesWordGuards(
	ctx context.Context,
	guildID string,
	word string,
	message discord.Message,
	settings *ent.Settings,
	g *ent.Game,
) bool {
	usedInPastHundred, err := s.checkUsedInPastHundred(ctx, g.ID, word)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Errorw(
			"game: add word: check used in past hundred failed",
			"error", err, "guildID", guildID, "gameID", g.ID,
			"userID", message.Author.ID,
		)

		return false
	}

	if usedInPastHundred {
		go s.replyAndDelete(
			message,
			fmt.Sprintf(
				"The word %s has already been used in the past 100 words,"+
					" try another word!",
				word,
			),
			true, "❌",
		)

		return false
	}

	cooldown, err := s.checkCooldown(
		ctx, message.Author.ID.String(), g.ID, settings.Cooldown,
	)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Errorw(
			"game: add word: check cooldown failed",
			"error", err, "guildID", guildID, "gameID", g.ID,
			"userID", message.Author.ID,
		)

		return false
	}

	if cooldown.After(time.Now()) {
		go s.replyAndDelete(
			message,
			fmt.Sprintf(
				"You're on a cooldown, you can try again %s",
				hammertime.Format(cooldown, hammertime.Span),
			),
			true, "🕒",
		)

		return false
	}

	return true
}

func (s *GameService) recordWordAndReact(
	ctx context.Context,
	guildID string,
	word string,
	message discord.Message,
	settings *ent.Settings,
	g *ent.Game,
) {
	client := s.client.Client()

	go func() {
		utils.LogIfErr(
			utils.Logger, "add-game-points",
			s.points.AddGamePoints(ctx, guildID, message.Author.ID.String(), 1),
		)
	}()

	msgID := message.ID.String()

	_, err := s.database.History.Create().
		SetUserID(message.Author.ID.String()).
		SetGameID(g.ID).
		SetWord(word).
		SetNillableMessageID(&msgID).
		Save(ctx)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: create history failed",
			"error", err, "guildID", guildID, "gameID", g.ID,
			"userID", message.Author.ID, "messageID", message.ID,
		)

		return
	}

	count, err := s.getCount(ctx, g.ID)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: get count failed",
			"error", err, "guildID", guildID, "gameID", g.ID,
		)

		return
	}

	isHighscore, isGameHighscored := s.checkStreak(ctx, settings, g, count)

	if isGameHighscored {
		utils.LogIfErrNoRateLimit(
			utils.Logger, "message-reaction-add",
			client.Rest.AddReaction(message.ChannelID, message.ID, "🎉"),
		)
	}

	emoji := "✅"
	if isHighscore {
		emoji = "☑️"
	}

	utils.LogIfErrNoRateLimit(
		utils.Logger, "message-reaction-add",
		client.Rest.AddReaction(message.ChannelID, message.ID, emoji),
	)
	s.checkSpecialReactions(message, word)
	s.setNumber(message, count)

	if utils.IsPalindrome(word) {
		go func() {
			utils.LogIfErrNoRateLimit(
				utils.Logger, "message-reaction-add",
				client.Rest.AddReaction(
					message.ChannelID, message.ID, "🪞",
				),
			)
		}()
	}
}

func (s *GameService) IsEqualToLast(
	ctx context.Context,
	message discord.Message,
	settings *ent.Settings,
	isDelete bool,
) (ok bool, word string) {
	ok = true

	if message.ID == 0 {
		ok = false
		return ok, word
	}

	g, exists, err := s.GetCurrentGame(ctx, message.GuildID.String())
	if err != nil || !exists {
		utils.Logger.Info("Couldnt find game", err)
		return ok, word
	}

	h, _, err := s.GetLastHistory(ctx, g)
	if err != nil || h == nil {
		utils.Logger.Info("Couldnt find last history", err)
		return ok, word
	}

	if h.MessageID == nil {
		return ok, word
	}

	if *h.MessageID != message.ID.String() {
		return ok, word
	}

	word = h.Word

	if isDelete {
		ok = false
		return ok, word
	}

	parsedWord, err := s.ParseWord(message)
	if err != nil {
		ok = false
		return ok, word
	}

	utils.Logger.Info("Checking is equal", message.Content)

	if parsedWord != word {
		ok = false
	}

	return ok, word
}

func (s *GameService) GetCurrentGame(
	ctx context.Context,
	guildID string,
) (g *ent.Game, exists bool, err error) {
	exists = true
	g, err = s.database.Game.Query().
		Where(
			game.GuildIDEQ(guildID),
			game.StatusEQ(game.StatusIN_PROGRESS),
		).
		Order(ent.Desc(game.FieldCreatedAt)).
		First(ctx)

	if ent.IsNotFound(err) {
		err = nil
		exists = false

		return g, exists, err
	}

	if err != nil {
		return g, exists, fmt.Errorf("game: get current: %w", err)
	}

	return g, exists, err
}

func (s *GameService) GetLastHistory(
	ctx context.Context,
	g *ent.Game,
) (h *ent.History, exists bool, err error) {
	if g == nil || g.Status != game.StatusIN_PROGRESS {
		exists = false
		return h, exists, err
	}

	exists = true
	h, err = s.database.History.Query().
		Where(history.GameIDEQ(g.ID)).
		Order(ent.Desc(history.FieldCreatedAt)).
		First(ctx)

	if ent.IsNotFound(err) {
		err = nil
		exists = false

		return h, exists, err
	}

	if err != nil {
		return h, exists, fmt.Errorf("game: get last history: %w", err)
	}

	return h, exists, err
}

func (s *GameService) getCount(
	ctx context.Context,
	gameID int,
) (count int, err error) {
	total, err := s.database.History.Query().
		Where(history.GameIDEQ(gameID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("game: get count: %w", err)
	}

	// Subtract 1 because the bot's initial word is counted
	count = total - 1
	if count < 0 {
		count = 0
	}

	return count, nil
}

func (s *GameService) checkStreak(
	ctx context.Context,
	settings *ent.Settings,
	g *ent.Game,
	count int,
) (isHighscore bool, isGameHighscored bool) {
	if count <= settings.Highscore {
		return false, false
	}

	isHighscore = true

	go func() {
		if _, setErr := s.settings.SetHighscoreByGuildID(
			ctx,
			settings.GuildID,
			count,
		); setErr != nil {
			utils.Logger.Errorw(
				"game: check streak: set highscore failed",
				"error", setErr,
				"guildID", settings.GuildID,
			)
		}
	}()

	if g.IsHighscored {
		return isHighscore, false
	}

	isGameHighscored = true

	go func() {
		if _, saveErr := s.database.Game.UpdateOneID(g.ID).
			SetIsHighscored(true).
			Save(ctx); saveErr != nil {
			utils.Logger.Errorw(
				"game: check streak: update game failed",
				"error", saveErr,
				"gameID", g.ID,
			)
		}
	}()

	return isHighscore, isGameHighscored
}

func (s *GameService) checkUsedInPastHundred(
	ctx context.Context,
	gameID int,
	word string,
) (used bool, err error) {
	histories, err := s.database.History.Query().
		Where(history.GameIDEQ(gameID)).
		Order(ent.Desc(history.FieldCreatedAt)).
		Limit(100).
		All(ctx)
	if err != nil {
		return used, fmt.Errorf("game: check used in past hundred: %w", err)
	}

	used = slices.ContainsFunc(histories, func(h *ent.History) bool {
		return h.Word == word
	})

	return used, nil
}

func (s *GameService) checkCooldown(
	ctx context.Context,
	userID string,
	gameID int,
	settingsCooldown int,
) (cooldown time.Time, err error) {
	if settingsCooldown == 0 {
		cooldown = time.Now().Add(-time.Second * 600)
		return cooldown, err
	}

	seconds := -time.Second * time.Duration(settingsCooldown)
	lastHistory, err := s.database.History.Query().
		Where(
			history.UserIDEQ(userID),
			history.GameIDEQ(gameID),
			history.CreatedAtGT(time.Now().Add(seconds)),
		).
		Order(ent.Desc(history.FieldCreatedAt)).
		First(ctx)

	if ent.IsNotFound(err) {
		cooldown = time.Now().Add(-time.Second * 10)
		err = nil

		return cooldown, err
	}

	if err != nil {
		return cooldown, err
	}

	cooldown = lastHistory.CreatedAt.Add(
		time.Second * time.Duration(settingsCooldown),
	)

	return cooldown, err
}

func (s *GameService) replyAndDelete(
	message discord.Message,
	messageToSend string,
	deleteAfter bool,
	emoji string,
) {
	client := s.client.Client()

	if len(emoji) > 0 {
		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			client.Rest.AddReaction(
				message.ChannelID,
				message.ID,
				emoji,
			),
		)
	}

	sentMessage, err := client.Rest.CreateMessage(
		message.ChannelID,
		discord.MessageCreate{
			Content: messageToSend,
			MessageReference: &discord.MessageReference{
				MessageID: &message.ID,
				ChannelID: &message.ChannelID,
				GuildID:   message.GuildID,
			},
		},
	)
	if err != nil {
		utils.Logger.Errorw(
			"game: reply and delete: send reply failed",
			"error",
			err,
			"channelID",
			message.ChannelID,
			"messageID",
			message.ID,
		)

		return
	}

	if deleteAfter {
		time.AfterFunc(time.Second*5, func() {
			utils.LogIfErr(
				utils.Logger,
				"channel-message-delete",
				client.Rest.DeleteMessage(
					sentMessage.ChannelID,
					sentMessage.ID,
					rest.WithReason("auto-delete after reply"),
				),
			)
		})
	}
}

func (s *GameService) checkSpecialReactions(
	message discord.Message,
	word string,
) {
}

func (s *GameService) getRandomLetter() string {
	letters := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
		"g",
		"h",
		"i",
		"j",
		"k",
		"l",
		"m",
		"n",
		"o",
		"p",
		"q",
		"r",
		"s",
		"t",
		"u",
		"v",
		"w",
		"y",
		"z",
	}
	weights := []int{
		382, 963, 1276, 1351, 1411, 1493, 1544, 1603, 1637, 1647, 1657, 1730,
		1801, 1828, 1858, 1970, 1975, 2077, 2286, 2387, 2408, 2443, 2493, 2503,
		2513,
	}

	maxCumulativeWeight := weights[len(weights)-1]

	randomNumber := rand.IntN(maxCumulativeWeight-1) + 1
	index := slices.IndexFunc(weights, func(v int) bool {
		return v >= randomNumber
	})

	return letters[index]
}

func (s *GameService) setNumber(message discord.Message, count int) {
	client := s.client.Client()

	stringCount := strconv.Itoa(count)
	usedEmojis := []string{}

	for _, number := range stringCount {
		i, err := strconv.Atoi(string(number))
		if err != nil {
			continue
		}

		availableEmojis := localStatic.NumberEmojis[i]
		for _, emoji := range availableEmojis {
			if slices.Contains(usedEmojis, emoji) {
				continue
			}

			usedEmojis = append(usedEmojis, emoji)
			utils.LogIfErrNoRateLimit(
				utils.Logger,
				"message-reaction-add",
				client.Rest.AddReaction(
					message.ChannelID,
					message.ID,
					emoji,
				),
			)

			break
		}
	}
}

func (s *GameService) CountByGuildIDs(
	ctx context.Context,
	guildIDs []string,
) (games int, h int, err error) {
	gameResult, err := s.database.Game.Query().
		Where(game.GuildIDIn(guildIDs...)).
		Count(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, 0, fmt.Errorf("game: count by guild ids: games: %w", err)
	}

	historyResult, err := s.database.History.Query().
		Where(history.HasGameWith(game.GuildIDIn(guildIDs...))).
		Count(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, 0, fmt.Errorf("game: count by guild ids: history: %w", err)
	}

	return gameResult, historyResult, nil
}

func (s *GameService) DeleteByGuildIDs(
	ctx context.Context,
	guildIDs []string,
) (games int, h int, err error) {
	historyResult, hErr := s.database.History.Delete().
		Where(history.HasGameWith(game.GuildIDIn(guildIDs...))).
		Exec(ctx)
	if hErr != nil && !ent.IsNotFound(hErr) {
		return 0, 0, fmt.Errorf("game: delete by guild ids: history: %w", hErr)
	}

	gameResult, err := s.database.Game.Delete().
		Where(game.GuildIDIn(guildIDs...)).
		Exec(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, 0, fmt.Errorf("game: delete by guild ids: games: %w", err)
	}

	return gameResult, historyResult, nil
}

func (s *GameService) ResetEmptyGames(ctx context.Context) (int, error) {
	games, err := s.database.Game.Query().
		Where(
			game.StatusEQ(game.StatusIN_PROGRESS),
			game.Not(game.HasHistory()),
		).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("game: reset empty: query: %w", err)
	}

	count := 0

	for _, g := range games {
		_, started, startErr := s.Start(
			ctx,
			g.GuildID,
			g.Type,
			"",
			true,
		)
		if startErr != nil {
			utils.Logger.Warnw(
				"game: reset empty: restart failed",
				"error",
				startErr,
				"gameID",
				g.ID,
			)

			continue
		}

		if started {
			count++
		}
	}

	return count, nil
}

func (s *GameService) CountEmptyGames(ctx context.Context) (int, error) {
	count, err := s.database.Game.Query().
		Where(
			game.StatusEQ(game.StatusIN_PROGRESS),
			game.Not(game.HasHistory()),
		).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("game: count empty: %w", err)
	}

	return count, nil
}

type GuildIDRow struct {
	GuildID string `json:"guildId"`
}

func (s *GameService) FindAllGuildIDs(
	ctx context.Context,
) ([]GuildIDRow, error) {
	games, err := s.database.Game.Query().
		Select(game.FieldGuildID).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("game: find distinct guild ids: %w", err)
	}

	seen := make(map[string]struct{})
	rows := make([]GuildIDRow, 0)

	for _, g := range games {
		if _, ok := seen[g.GuildID]; !ok {
			seen[g.GuildID] = struct{}{}
			rows = append(rows, GuildIDRow{GuildID: g.GuildID})
		}
	}

	return rows, nil
}

// LoadActiveGameChannels pre-populates the shared active-game registry from
// the database. Call this once on bot startup so the message cache policy
// works correctly for games that were running before the bot restarted.
func (s *GameService) LoadActiveGameChannels(ctx context.Context) error {
	games, err := s.database.Game.Query().
		Where(game.StatusEQ(game.StatusIN_PROGRESS)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("load active game channels: query: %w", err)
	}

	for _, g := range games {
		settings, settErr := s.settings.GetByGuildID(ctx, g.GuildID)
		if settErr != nil || settings.ChannelID == nil ||
			*settings.ChannelID == "" {
			continue
		}

		guildSnowflake, guildErr := snowflake.Parse(g.GuildID)

		channelSnowflake, chanErr := snowflake.Parse(*settings.ChannelID)
		if guildErr != nil || chanErr != nil {
			continue
		}

		utils.ActiveGames.Register(guildSnowflake, channelSnowflake)
	}

	utils.Logger.Infof(
		"pre-populated active games cache with %d games",
		len(games),
	)

	return nil
}
