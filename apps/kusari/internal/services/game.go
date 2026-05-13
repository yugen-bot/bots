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

func (service *GameService) Start(ctx context.Context, guildID string, gameType db.GameType, word string, recreate bool) (game *db.GameModel, started bool, err error) {
	utils.Logger.Infof("Trying to start a game for %s", guildID)

	currentGame, exists, err := service.GetCurrentGame(ctx, guildID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		utils.Logger.Error(err)
		return game, started, err
	}

	settings, err := service.settings.GetByGuildId(ctx, guildID)
	if err != nil {
		utils.Logger.Error(err)
		return game, started, fmt.Errorf("game: start: get settings: %w", err)
	}

	channelID, ok := settings.ChannelID()
	if !ok {
		err = ErrNoChannelIDConfigured
		utils.Logger.Error(err)
		return game, started, err
	}

	channel, err := service.bot.Channel(channelID)
	if err != nil {
		utils.Logger.Error(err)
		return game, started, fmt.Errorf("game: start: get channel: %w", err)
	}

	if exists && !recreate {
		started = false
		return game, started, err
	}

	if (exists && recreate) || (exists && currentGame.Type != gameType) {
		service.End(ctx, currentGame.ID, db.GameStatusFailed)
	}

	started = true

	game, err = service.database.Game.CreateOne(
		db.Game.Settings.Link(
			db.Settings.ID.Equals(settings.ID),
		),
		db.Game.Type.Set(gameType),
	).Exec(ctx)
	if err != nil {
		utils.Logger.Error(err)
		return game, started, fmt.Errorf("game: start: create game: %w", err)
	}

	self := service.bot.State.User

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
		utils.Logger.Error(err)
	}

	if channel.Type == discordgo.ChannelTypeGuildText || channel.Type == discordgo.ChannelTypeGuildPublicThread || channel.Type == discordgo.ChannelTypeGuildPrivateThread {
		service.bot.ChannelMessageSend(
			channelID,
			fmt.Sprintf(
				`**A new game has started!**
The first letter is: **%s**`,

				strings.ToUpper(string(word[len(word)-1])),
			),
		)
	}

	return game, started, err
}

func (service *GameService) End(ctx context.Context, gameID int, status db.GameStatus) (game *db.GameModel, err error) {
	game, err = service.database.Game.FindUnique(
		db.Game.ID.Equals(gameID),
	).Update(
		db.Game.Status.Set(status),
	).Exec(ctx)
	if err != nil {
		return game, fmt.Errorf("game: end: update game: %w", err)
	}

	return game, err
}

func (service *GameService) ParseWord(message *discordgo.Message) (word string, err error) {
	if message.Author.Bot {
		err = ErrAuthorIsBot
		return word, err
	}

	words := strings.Fields(message.Content)
	if len(words) == 0 || len(words) > 1 {
		return word, err
	}

	word = words[0]

	firstLetterRegex, _ := regexp.Compile("^[A-Za-z!]+$")
	lastLetterRegex, _ := regexp.Compile("^[A-Za-z]+$")

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

func (service *GameService) AddWord(ctx context.Context, guildID string, word string, message *discordgo.Message, settings *db.SettingsModel) {
	if len(word) == 0 {
		return
	}

	game, exists, err := service.GetCurrentGame(ctx, guildID)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	if !exists {
		return
	}

	history, _, err := service.GetLastHistory(ctx, game)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	isSameUser := service.cfg.Env != "development" && message.Author.ID == history.UserID
	if isSameUser {
		service.bot.MessageReactionAdd(message.ChannelID, message.ID, "🕒")
		go service.bot.ChannelMessageSendReply(
			message.ChannelID,
			"Sorry, but you can't add a word twice in a row! Please wait for another player to add a word.",
			message.Reference(),
		)
		return
	}

	lastLetter := history.Word[len(history.Word)-1]
	isCorrectLetter := word[0] == lastLetter
	wordExists, err := service.dictionary.Check(ctx, word)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	if !isCorrectLetter || !wordExists {
		failReason := fmt.Sprintf(`Sorry, I couldn't find "**%s**" in the [English dictionary](https://en.wiktionary.org/wiki/%s), try again!`, word, word)

		if !isCorrectLetter {
			failReason = fmt.Sprintf("The word %s does not start with the letter **%s**", word, string(lastLetter))
		}

		if err != nil {
			utils.Logger.Error(err)
			return
		}

		saves, err := service.saves.GetSaves(ctx, settings, message.Author.ID)
		if err != nil {
			utils.Logger.Error(err)
			return
		}

		service.bot.MessageReactionAdd(message.ChannelID, message.ID, "❌")

		if saves.player >= 1 {
			leftoverSaves, maxSaves, err := service.saves.DeductSaveFromPlayer(ctx, message.Author.ID, 1)
			if err != nil {
				utils.Logger.Error(err)
				return
			}

			go service.bot.ChannelMessageSendReply(
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
			return
		}

		if saves.guild >= 1 {
			leftoverSaves, maxSaves, err := service.saves.DeductSaveFromGuild(ctx, message.GuildID, settings, 1)
			if err != nil {
				return
			}

			go service.bot.ChannelMessageSendReply(
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
			return
		}

		count, err := service.getCount(ctx, game.ID)
		if err != nil {
			utils.Logger.Error(err)
			return
		}

		isHighscore, _, err := service.checkStreak(ctx, settings, game, count)
		if err != nil {
			utils.Logger.Error(err)
			return
		}

		highScoreText := ""
		if isHighscore {
			highScoreText = "\n**A new highscore has been set! 🎉**"
		}

		pointsRemoved := int(count / 10)
		go service.points.RemoveGamePoints(ctx, guildID, message.Author.ID, pointsRemoved)

		if pointsRemoved == 0 {
			pointsRemoved = 1
		}

		pointsRemovedText := ""
		pointText := "Points have"
		if pointsRemoved == 1 {
			pointText = "Point has"
		}

		pointsRemovedText = fmt.Sprintf("\n\n**%d %s been removed from your account.**", pointsRemoved, pointText)

		go service.bot.ChannelMessageSendReply(
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

		service.Start(ctx, guildID, db.GameTypeNormal, "", true)
		return
	}

	usedInPastHundred, err := service.checkUsedInPastHundred(ctx, game.ID, word)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		utils.Logger.Error(err)
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
		utils.Logger.Error(err)
		return
	}

	if cooldown.After(time.Now()) {
		go service.replyAndDelete(
			message,
			fmt.Sprintf("You're on a cooldown, you can try again %s", hammertime.Format(cooldown, hammertime.Span)),
			true,
			"🕒",
		)
		return
	}

	go service.points.AddGamePoints(ctx, guildID, message.Author.ID, 1)
	_, err = service.database.History.CreateOne(
		db.History.UserID.Set(message.Author.ID),
		db.History.Game.Link(db.Game.ID.Equals(game.ID)),
		db.History.Word.Set(word),
		db.History.MessageID.Set(message.ID),
	).Exec(ctx)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	count, err := service.getCount(ctx, game.ID)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	isHighscore, isGameHighscored, err := service.checkStreak(ctx, settings, game, count)

	if isGameHighscored {
		service.bot.MessageReactionAdd(message.ChannelID, message.ID, "🎉")
	}

	emoji := "✅"
	if isHighscore {
		emoji = "☑️"
	}
	service.bot.MessageReactionAdd(message.ChannelID, message.ID, emoji)
	service.checkSpecialReactions(message, word)

	service.setNumber(message, count)

	if utils.IsPalindrome(word) {
		go service.bot.MessageReactionAdd(message.ChannelID, message.ID, "🪞")
	}
}

func (service *GameService) IsEqualToLast(ctx context.Context, message *discordgo.Message, settings *db.SettingsModel, isDelete bool) (ok bool, word string) {
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
	if err != nil {
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

func (service *GameService) GetCurrentGame(ctx context.Context, guildID string) (game *db.GameModel, exists bool, err error) {
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

func (service *GameService) GetLastHistory(ctx context.Context, game *db.GameModel) (history *db.HistoryModel, exists bool, err error) {
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

func (service *GameService) getCount(ctx context.Context, gameID int) (count int, err error) {
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

func (service *GameService) checkStreak(ctx context.Context, settings *db.SettingsModel, game *db.GameModel, count int) (isHighscore bool, isGameHighscored bool, err error) {
	isHighscore = false
	isGameHighscored = false

	if count > settings.Highscore {
		isHighscore = true
		go service.settings.SetHighscoreByGuildID(ctx, settings.GuildID, count)

		if !game.IsHighscored {
			isGameHighscored = true

			go service.database.Game.FindUnique(
				db.Game.ID.Equals(game.ID),
			).Update(
				db.Game.IsHighscored.Set(true),
			).Exec(ctx)
		}
	}

	return isHighscore, isGameHighscored, err
}

func (service *GameService) checkUsedInPastHundred(ctx context.Context, gameID int, word string) (used bool, err error) {
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

func (service *GameService) checkCooldown(ctx context.Context, userID string, gameID int, settingsCooldown int) (cooldown time.Time, err error) {
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

	cooldown = lastHistory.CreatedAt.Add(time.Second * time.Duration(settingsCooldown))
	return cooldown, err
}

func (service *GameService) replyAndDelete(message *discordgo.Message, messageToSend string, deleteAfter bool, emoji string) {
	if len(emoji) > 0 {
		service.bot.MessageReactionAdd(message.ChannelID, message.ID, emoji)
	}

	sentMessage, err := service.bot.ChannelMessageSendReply(
		message.ChannelID,
		messageToSend,
		message.Reference(),
	)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	if deleteAfter {
		time.AfterFunc(time.Second*5, func() {
			service.bot.ChannelMessageDelete(sentMessage.ChannelID, sentMessage.ID)
		})
	}
}

func (service *GameService) checkSpecialReactions(message *discordgo.Message, word string) {
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
			service.bot.MessageReactionAdd(message.ChannelID, message.ID, emoji)
			break
		}
	}
}
