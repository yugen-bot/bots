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

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/kusari/prisma/db"
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
	bot        *discordgoplus.Bot
	cfg        *config.Config
	database   *db.PrismaClient
	settings   *SettingsService
	saves      *SavesService
	points     *PointsService
	dictionary *DictionaryService
}

func CreateGameService(container *di.Container) *GameService {
	utils.Logger.Info("Creating Game Service")

	return &GameService{
		bot:        container.Get(static.DiBot).(*discordgoplus.Bot),
		cfg:        container.Get(static.DiConfig).(*config.Config),
		database:   container.Get(static.DiDatabase).(*db.PrismaClient),
		settings:   container.Get(static.DiSettings).(*SettingsService),
		saves:      container.Get(localStatic.DiSaves).(*SavesService),
		points:     container.Get(localStatic.DiPoints).(*PointsService),
		dictionary: container.Get(localStatic.DiDictionary).(*DictionaryService),
	}
}

func (service *GameService) Start(
	ctx context.Context,
	guildID string,
	gameType db.GameType,
	word string,
	recreate bool,
) (game *db.GameModel, started bool, err error) {
	utils.Logger.Infof("Trying to start a game for %s", guildID)

	currentGame, exists, err := service.GetCurrentGame(ctx, guildID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		utils.Logger.Errorw(
			"game: start: get current game failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return game, started, err
	}

	settings, err := service.settings.GetByGuildId(ctx, guildID)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: get settings failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return game, started, fmt.Errorf("game: start: get settings: %w", err)
	}

	channelID, ok := settings.ChannelID()
	if !ok {
		err = ErrNoChannelIDConfigured
		utils.Logger.Errorw(
			"game: start: no channel id configured",
			"error",
			err,
			"guildID",
			guildID,
		)

		return game, started, err
	}

	channel, err := service.bot.Channel(channelID)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: get channel failed",
			"error",
			err,
			"guildID",
			guildID,
			"channelID",
			channelID,
		)

		return game, started, fmt.Errorf("game: start: get channel: %w", err)
	}

	if exists && !recreate {
		return game, started, err
	}

	if (exists && recreate) || (exists && currentGame.Type != gameType) {
		if _, endErr := service.End(
			ctx,
			currentGame.ID,
			db.GameStatusFailed,
		); endErr != nil {
			utils.Logger.Warnw(
				"game: start: end current game failed",
				"error", endErr,
				"guildID", guildID,
				"gameID", currentGame.ID,
			)
		}
	}

	started = true

	game, err = service.database.Game.CreateOne(
		db.Game.Settings.Link(
			db.Settings.ID.Equals(settings.ID),
		),
		db.Game.Type.Set(gameType),
	).Exec(ctx)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: create game failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return game, started, fmt.Errorf("game: start: create game: %w", err)
	}

	shard, err := service.bot.ShardByGuild(guildID)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: could not retrieve shard",
			"error",
			err,
			"guildID",
			guildID,
		)

		return game, started, fmt.Errorf(
			"game: start: could not retrieve shard: %w",
			err,
		)
	}

	self := shard.State.User

	if len(word) <= 0 {
		// get word
		word = service.getRandomLetter()
	}

	_, err = service.database.History.CreateOne(
		db.History.UserID.Set(self.ID),
		db.History.Game.Link(db.Game.ID.Equals(game.ID)),
		db.History.Word.Set(word),
	).Exec(ctx)
	if err != nil {
		return game, started, fmt.Errorf("game: start: create history: %w", err)
	}

	if channel.Type == discordgo.ChannelTypeGuildText ||
		channel.Type == discordgo.ChannelTypeGuildPublicThread ||
		channel.Type == discordgo.ChannelTypeGuildPrivateThread {
		go func() {
			_, sendErr := shard.ChannelMessageSend(
				channelID,
				fmt.Sprintf(
					`**A new game has started!**
The first letter is: **%s**`,

					strings.ToUpper(string(word[len(word)-1])),
				),
			)
			utils.LogIfErr(utils.Logger, "channel-message-send", sendErr)
		}()
	}

	return game, started, err
}

func (service *GameService) End(
	ctx context.Context,
	gameID int,
	status db.GameStatus,
) (game *db.GameModel, err error) {
	game, err = service.database.Game.FindUnique(
		db.Game.ID.Equals(gameID),
	).Update(
		db.Game.Status.Set(status),
	).Exec(ctx)
	if err != nil {
		return game, fmt.Errorf("game: end: update game: %w", err)
	}

	if _, delErr := service.database.History.FindMany(
		db.History.GameID.Equals(gameID),
	).Delete().Exec(ctx); delErr != nil {
		utils.Logger.Warnw(
			"game: end: delete history failed",
			"error", delErr,
			"gameID", gameID,
		)
	}

	return game, err
}

func (service *GameService) ParseWord(
	message *discordgo.Message,
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

func (service *GameService) AddWord(
	ctx context.Context,
	guildID string,
	word string,
	message *discordgo.Message,
	settings *db.SettingsModel,
) {
	if len(word) == 0 {
		return
	}

	b, err := service.bot.ShardByGuild(guildID)
	if err != nil {
		utils.Logger.Warnw(
			"game: add word: ShardByGuild failed",
			"error",
			err,
			"guildID",
			guildID,
		)
		return
	}

	game, exists, err := service.GetCurrentGame(ctx, guildID)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: get current game failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return
	}

	if !exists {
		return
	}

	history, _, err := service.GetLastHistory(ctx, game)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: get last history failed",
			"error",
			err,
			"guildID",
			guildID,
			"gameID",
			game.ID,
		)

		return
	}

	isSameUser := false
	if history != nil && message != nil && message.Author != nil {
		isSameUser = service.cfg.Env != "development" &&
			message.Author.ID == history.UserID
	}

	if history == nil {
		utils.Logger.Debugw(
			"History is nil",
			"guildID",
			guildID,
			"gameID",
			game.ID,
		)
	}

	if message == nil {
		utils.Logger.Debugw(
			"Message is nil",
			"guildID",
			guildID,
			"gameID",
			game.ID,
		)
	}

	if message.Author == nil {
		utils.Logger.Debugw(
			"Author is nil",
			"guildID",
			guildID,
			"gameID",
			game.ID,
		)
	}

	if isSameUser {
		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			b.MessageReactionAdd(message.ChannelID, message.ID, "🕒"),
		)

		_, sendErr := b.ChannelMessageSendReply(
			message.ChannelID,
			"Sorry, but you can't add a word twice in a row! Please wait for another player to add a word.",
			message.Reference(),
		)
		utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)

		return
	}

	lastLetter := history.Word[len(history.Word)-1]
	isCorrectLetter := word[0] == lastLetter

	wordExists, err := service.dictionary.Check(ctx, word)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: dictionary check failed",
			"error",
			err,
			"guildID",
			guildID,
			"gameID",
			game.ID,
			"userID",
			message.Author.ID,
			"word",
			word,
		)

		return
	}

	if !isCorrectLetter || !wordExists {
		// Build failure reason
		failReason := fmt.Sprintf(
			`Sorry, I couldn't find "**%s**" in the [English dictionary](https://en.wiktionary.org/wiki/%s), try again!`,
			word,
			word,
		)
		if !isCorrectLetter {
			failReason = fmt.Sprintf(
				"The word %s does not start with the letter **%s**",
				word,
				string(lastLetter),
			)
		}

		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			b.MessageReactionAdd(message.ChannelID, message.ID, "❌"),
		)

		saves, err := service.saves.GetSaves(ctx, settings, message.Author.ID)
		if err != nil {
			utils.Logger.Errorw(
				"game: add word: get saves failed",
				"error",
				err,
				"guildID",
				guildID,
				"gameID",
				game.ID,
				"userID",
				message.Author.ID,
			)

			return
		}

		if saves.player >= 1 {
			leftoverSaves, maxSaves, err := service.saves.DeductSaveFromPlayer(
				ctx,
				message.Author.ID,
				1,
			)
			if err != nil {
				utils.Logger.Errorw(
					"game: add word: deduct player save failed",
					"error",
					err,
					"guildID",
					guildID,
					"gameID",
					game.ID,
					"userID",
					message.Author.ID,
				)

				return
			}

			go func() {
				_, sendErr := b.ChannelMessageSendReply(
					message.ChannelID,
					fmt.Sprintf(
						`%s
Used **1 of your own** saves, You have **%s/%s** saves left.`,
						failReason,
						strconv.FormatFloat(leftoverSaves, 'f', -1, 64),
						strconv.FormatFloat(maxSaves, 'f', -1, 64),
					),
					message.Reference(),
				)
				utils.LogIfErr(
					utils.Logger,
					"channel-message-send-reply",
					sendErr,
				)
			}()

			return
		}

		if saves.guild >= 1 {
			leftoverSaves, maxSaves, err := service.saves.DeductSaveFromGuild(
				ctx,
				message.GuildID,
				settings,
				1,
			)
			if err != nil {
				utils.Logger.Errorw(
					"game: deduct guild save failed",
					"error", err,
					"guildID", guildID,
				)

				return
			}

			go func() {
				_, sendErr := b.ChannelMessageSendReply(
					message.ChannelID,
					fmt.Sprintf(
						`%s
Used **1 server** save, There are **%s/%s** server saves left.`,
						failReason,
						strconv.FormatFloat(leftoverSaves, 'f', -1, 64),
						strconv.FormatFloat(maxSaves, 'f', -1, 64),
					),
					message.Reference(),
				)
				utils.LogIfErr(
					utils.Logger,
					"channel-message-send-reply",
					sendErr,
				)
			}()

			return
		}

		count, err := service.getCount(ctx, game.ID)
		if err != nil {
			utils.Logger.Errorw(
				"game: add word: get count failed",
				"error",
				err,
				"guildID",
				guildID,
				"gameID",
				game.ID,
			)

			return
		}

		isHighscore, _, err := service.checkStreak(ctx, settings, game, count)
		if err != nil {
			utils.Logger.Errorw(
				"game: add word: check streak failed",
				"error",
				err,
				"guildID",
				guildID,
				"gameID",
				game.ID,
			)

			return
		}

		highScoreText := ""
		if isHighscore {
			highScoreText = "\n**A new highscore has been set! 🎉**"
		}

		// Deduct points from the player who broke the chain
		pointsRemoved := int(count / 10)

		go func() {
			utils.LogIfErr(
				utils.Logger,
				"remove-game-points",
				service.points.RemoveGamePoints(
					ctx,
					guildID,
					message.Author.ID,
					pointsRemoved,
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
			pointsRemoved,
			pointText,
		)

		go func() {
			_, sendErr := b.ChannelMessageSendReply(
				message.ChannelID,
				fmt.Sprintf(
					`%s
**The game has ended on a streak of %d!**%s%s

**Want to save the game?** Make sure to **/vote** for Kusari and earn yourself saves to save the game!`,
					failReason,
					count,
					highScoreText,
					pointsRemovedText,
				),
				message.Reference(),
			)
			utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
		}()

		if _, _, startErr := service.Start(
			ctx,
			guildID,
			db.GameTypeNormal,
			"",
			true,
		); startErr != nil {
			utils.Logger.Warnw(
				"game: add word: restart failed",
				"error", startErr,
				"guildID", guildID,
				"gameID", game.ID,
			)
		}

		return
	}

	usedInPastHundred, err := service.checkUsedInPastHundred(ctx, game.ID, word)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		utils.Logger.Errorw(
			"game: add word: check used in past hundred failed",
			"error",
			err,
			"guildID",
			guildID,
			"gameID",
			game.ID,
			"userID",
			message.Author.ID,
		)

		return
	}

	if usedInPastHundred {
		go service.replyAndDelete(
			message,
			fmt.Sprintf(
				"The word %s has already been used in the past 100 words, try another word!",
				word,
			),
			true,
			"❌",
		)

		return
	}

	cooldown, err := service.checkCooldown(
		ctx,
		message.Author.ID,
		game.ID,
		settings.Cooldown,
	)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		utils.Logger.Errorw(
			"game: add word: check cooldown failed",
			"error",
			err,
			"guildID",
			guildID,
			"gameID",
			game.ID,
			"userID",
			message.Author.ID,
		)

		return
	}

	if cooldown.After(time.Now()) {
		go service.replyAndDelete(
			message,
			fmt.Sprintf(
				"You're on a cooldown, you can try again %s",
				hammertime.Format(cooldown, hammertime.Span),
			),
			true,
			"🕒",
		)

		return
	}

	// Record points and history
	go func() {
		utils.LogIfErr(
			utils.Logger,
			"add-game-points",
			service.points.AddGamePoints(ctx, guildID, message.Author.ID, 1),
		)
	}()

	_, err = service.database.History.CreateOne(
		db.History.UserID.Set(message.Author.ID),
		db.History.Game.Link(db.Game.ID.Equals(game.ID)),
		db.History.Word.Set(word),
		db.History.MessageID.Set(message.ID),
	).Exec(ctx)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: create history failed",
			"error",
			err,
			"guildID",
			guildID,
			"gameID",
			game.ID,
			"userID",
			message.Author.ID,
			"messageID",
			message.ID,
		)

		return
	}

	// Check streak and react
	count, err := service.getCount(ctx, game.ID)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: get count failed",
			"error",
			err,
			"guildID",
			guildID,
			"gameID",
			game.ID,
		)

		return
	}

	isHighscore, isGameHighscored, err := service.checkStreak(
		ctx,
		settings,
		game,
		count,
	)
	if err != nil {
		utils.Logger.Errorw(
			"game: add word: check streak failed",
			"error",
			err,
			"guildID",
			guildID,
			"gameID",
			game.ID,
		)
	}

	if isGameHighscored {
		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			b.MessageReactionAdd(message.ChannelID, message.ID, "🎉"),
		)
	}

	emoji := "✅"
	if isHighscore {
		emoji = "☑️"
	}

	utils.LogIfErr(
		utils.Logger,
		"message-reaction-add",
		b.MessageReactionAdd(message.ChannelID, message.ID, emoji),
	)
	service.checkSpecialReactions(message, word)
	service.setNumber(message, count)

	if utils.IsPalindrome(word) {
		go func() {
			utils.LogIfErr(
				utils.Logger,
				"message-reaction-add",
				b.MessageReactionAdd(
					message.ChannelID,
					message.ID,
					"🪞",
				),
			)
		}()
	}
}

func (service *GameService) IsEqualToLast(
	ctx context.Context,
	message *discordgo.Message,
	settings *db.SettingsModel,
	isDelete bool,
) (ok bool, word string) {
	ok = true

	if message == nil {
		ok = false
		return ok, word
	}

	game, exists, err := service.GetCurrentGame(ctx, message.GuildID)
	if err != nil || !exists {
		utils.Logger.Info("Couldnt find game", err)
		return ok, word
	}

	history, _, err := service.GetLastHistory(ctx, game)
	if err != nil || history == nil {
		utils.Logger.Info("Couldnt find last history", err)
		return ok, word
	}

	messageID, messageIDOk := history.MessageID()
	if !messageIDOk {
		return ok, word
	}

	if messageID != message.ID {
		return ok, word
	}

	word = history.Word

	if isDelete {
		ok = false
		return ok, word
	}

	parsedWord, err := service.ParseWord(message)
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

func (service *GameService) GetCurrentGame(
	ctx context.Context,
	guildID string,
) (game *db.GameModel, exists bool, err error) {
	exists = true
	game, err = service.database.Game.FindFirst(
		db.Game.GuildID.Equals(guildID),
		db.Game.Status.Equals(db.GameStatusInProgress),
	).OrderBy(
		db.Game.CreatedAt.Order(db.SortOrderDesc),
	).Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		err = nil
		exists = false

		return game, exists, err
	}

	if err != nil {
		return game, exists, fmt.Errorf("game: get current: %w", err)
	}

	return game, exists, err
}

func (service *GameService) GetLastHistory(
	ctx context.Context,
	game *db.GameModel,
) (history *db.HistoryModel, exists bool, err error) {
	if game == nil || game.Status != db.GameStatusInProgress {
		exists = false
		return history, exists, err
	}

	exists = true
	history, err = service.database.History.FindFirst(
		db.History.Game.Where(db.Game.ID.Equals(game.ID)),
	).OrderBy(
		db.History.CreatedAt.Order(db.SortOrderDesc),
	).Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		err = nil
		exists = false

		return history, exists, err
	}

	if err != nil {
		return history, exists, fmt.Errorf("game: get last history: %w", err)
	}

	return history, exists, err
}

func (service *GameService) getCount(
	ctx context.Context,
	gameID int,
) (count int, err error) {
	var res []struct {
		Count string `json:"count"`
	}

	err = service.database.Prisma.QueryRaw(
		`SELECT count(*) as count FROM "History" WHERE "gameId" = $1`,
		gameID,
	).Exec(ctx, &res)
	if err != nil {
		return count, err
	}

	if len(res) > 0 {
		count, err = strconv.Atoi(res[0].Count)
		count = count - 1
	}

	return count, err
}

func (service *GameService) checkStreak(
	ctx context.Context,
	settings *db.SettingsModel,
	game *db.GameModel,
	count int,
) (isHighscore bool, isGameHighscored bool, err error) {
	if count <= settings.Highscore {
		return false, false, nil
	}

	isHighscore = true

	go service.settings.SetHighscoreByGuildID(ctx, settings.GuildID, count)

	if game.IsHighscored {
		return isHighscore, false, nil
	}

	isGameHighscored = true

	go service.database.Game.FindUnique(
		db.Game.ID.Equals(game.ID),
	).Update(
		db.Game.IsHighscored.Set(true),
	).Exec(ctx) //nolint:errcheck

	return isHighscore, isGameHighscored, nil
}

func (service *GameService) checkUsedInPastHundred(
	ctx context.Context,
	gameID int,
	word string,
) (used bool, err error) {
	histories, err := service.database.History.FindMany(
		db.History.Game.Where(db.Game.ID.Equals(gameID)),
	).Take(100).OrderBy(
		db.History.CreatedAt.Order(db.SortOrderDesc),
	).Exec(ctx)
	if err != nil {
		return used, err
	}

	used = slices.ContainsFunc(histories, func(history db.HistoryModel) bool {
		return history.Word == word
	})

	return used, err
}

func (service *GameService) checkCooldown(
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
	lastHistory, err := service.database.History.FindFirst(
		db.History.UserID.Equals(userID),
		db.History.GameID.Equals(gameID),
		db.History.CreatedAt.After(time.Now().Add(seconds)),
	).Select(
		db.History.CreatedAt.Field(),
	).Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		cooldown = time.Now().Add(-time.Second * 10)
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

func (service *GameService) replyAndDelete(
	message *discordgo.Message,
	messageToSend string,
	deleteAfter bool,
	emoji string,
) {
	b, err := service.bot.ShardByChannel(message.ChannelID)
	if err != nil {
		utils.Logger.Warnw(
			"game: reply and delete: ShardByChannel failed",
			"error",
			err,
			"channelID",
			message.ChannelID,
		)
		return
	}

	if len(emoji) > 0 {
		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			b.MessageReactionAdd(
				message.ChannelID,
				message.ID,
				emoji,
			),
		)
	}

	sentMessage, err := b.ChannelMessageSendReply(
		message.ChannelID,
		messageToSend,
		message.Reference(),
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
				b.ChannelMessageDelete(
					sentMessage.ChannelID,
					sentMessage.ID,
				),
			)
		})
	}
}

func (service *GameService) checkSpecialReactions(
	message *discordgo.Message,
	word string,
) {
}

func (service *GameService) getRandomLetter() string {
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

func (service *GameService) setNumber(message *discordgo.Message, count int) {
	b, err := service.bot.ShardByChannel(message.ChannelID)
	if err != nil {
		utils.Logger.Warnw(
			"game: set number: ShardByChannel failed",
			"error",
			err,
			"channelID",
			message.ChannelID,
		)
		return
	}

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
			utils.LogIfErr(
				utils.Logger,
				"message-reaction-add",
				b.MessageReactionAdd(
					message.ChannelID,
					message.ID,
					emoji,
				),
			)

			break
		}
	}
}

func (service *GameService) CountByGuildIDs(
	ctx context.Context,
	guildIDs []string,
) (games int, history int, err error) {
	gameResult, err := service.database.Game.FindMany(
		db.Game.GuildID.In(guildIDs),
	).Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return 0, 0, fmt.Errorf("game: count by guild ids: games: %w", err)
	}

	historyResult, err := service.database.History.FindMany(
		db.History.Game.Where(db.Game.GuildID.In(guildIDs)),
	).Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return 0, 0, fmt.Errorf("game: count by guild ids: history: %w", err)
	}

	return len(gameResult), len(historyResult), nil
}

func (service *GameService) DeleteByGuildIDs(
	ctx context.Context,
	guildIDs []string,
) (games int, history int, err error) {
	historyResult, hErr := service.database.History.FindMany(
		db.History.Game.Where(db.Game.GuildID.In(guildIDs)),
	).Delete().Exec(ctx)
	if hErr != nil && !errors.Is(hErr, db.ErrNotFound) {
		return 0, 0, fmt.Errorf("game: delete by guild ids: history: %w", hErr)
	}

	gameResult, err := service.database.Game.FindMany(
		db.Game.GuildID.In(guildIDs),
	).Delete().Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return 0, 0, fmt.Errorf("game: delete by guild ids: games: %w", err)
	}

	return gameResult.Count, historyResult.Count, nil
}

type emptyGameRow struct {
	ID      int    `json:"id"`
	GuildID string `json:"guildId"`
	Type    string `json:"type"`
}

func (service *GameService) ResetEmptyGames(ctx context.Context) (int, error) {
	var rows []emptyGameRow

	err := service.database.Prisma.QueryRaw(
		`SELECT id, "guildId", type FROM "Game" WHERE status = 'IN_PROGRESS' AND id NOT IN (SELECT DISTINCT "gameId" FROM "History")`,
	).Exec(ctx, &rows)
	if err != nil {
		return 0, fmt.Errorf("game: reset empty: query: %w", err)
	}

	count := 0

	for _, row := range rows {
		_, started, startErr := service.Start(
			ctx,
			row.GuildID,
			db.GameType(row.Type),
			"",
			true,
		)
		if startErr != nil {
			utils.Logger.Warnw(
				"game: reset empty: restart failed",
				"error",
				startErr,
				"gameID",
				row.ID,
			)

			continue
		}

		if started {
			count++
		}
	}

	return count, nil
}

func (service *GameService) CountEmptyGames(ctx context.Context) (int, error) {
	var res []struct {
		Count string `json:"count"`
	}

	err := service.database.Prisma.QueryRaw(
		`SELECT count(*) as count FROM "Game" WHERE status = 'IN_PROGRESS' AND id NOT IN (SELECT DISTINCT "gameId" FROM "History")`,
	).Exec(ctx, &res)
	if err != nil {
		return 0, fmt.Errorf("game: count empty: %w", err)
	}

	if len(res) == 0 {
		return 0, nil
	}

	count, err := strconv.Atoi(res[0].Count)
	if err != nil {
		return 0, fmt.Errorf("game: count empty: parse count: %w", err)
	}

	return count, nil
}

type GuildIDRow struct {
	GuildID string `json:"guildId"`
}

func (service *GameService) FindAllGuildIDs(
	ctx context.Context,
) ([]GuildIDRow, error) {
	var rows []GuildIDRow
	if err := service.database.Prisma.QueryRaw(
		`SELECT DISTINCT "guildId" FROM "Game"`,
	).Exec(ctx, &rows); err != nil {
		return nil, fmt.Errorf("game: find distinct guild ids: %w", err)
	}

	return rows, nil
}
