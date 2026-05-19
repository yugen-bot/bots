package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/expr-lang/expr"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"
	localStatic "jurien.dev/yugen/kazu/internal/static"
	"jurien.dev/yugen/kazu/prisma/db"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

var (
	ErrNoChannelIDConfigured = errors.New("no channel id configured")
	ErrAuthorIsBot           = errors.New("author is bot")
	ErrNumberCannotBeZero    = errors.New("number cannot be zero")
	ErrCouldNotParseNumber   = errors.New("could not parse to a valid number")
	ErrExprTooLong           = errors.New("expression too long")
	ErrExprNotNumber         = errors.New(
		"expression did not evaluate to a number",
	)
)

type GameService struct {
	bot      *discordgoplus.Bot
	cfg      *config.Config
	database *db.PrismaClient
	settings *SettingsService
	saves    *SavesService
	points   *PointsService
}

func CreateGameService(container *di.Container) *GameService {
	utils.Logger.Info("Creating Game Service")

	return &GameService{
		bot:      container.Get(static.DiBot).(*discordgoplus.Bot),
		cfg:      container.Get(static.DiConfig).(*config.Config),
		database: container.Get(static.DiDatabase).(*db.PrismaClient),
		settings: container.Get(static.DiSettings).(*SettingsService),
		saves:    container.Get(localStatic.DiSaves).(*SavesService),
		points:   container.Get(localStatic.DiPoints).(*PointsService),
	}
}

type ShameOptions struct {
	message  *discordgo.Message
	settings *db.SettingsModel
}

func (service *GameService) Start(
	ctx context.Context,
	guildID string,
	gameType db.GameType,
	startingNumber int,
	recreate bool,
	shame ...*ShameOptions,
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
			shame...); endErr != nil {
			utils.Logger.Warnw(
				"game: start: end current game failed",
				"error", endErr,
				"guildID", guildID,
				"gameID", currentGame.ID,
			)
		}
	}

	started = true
	number := startingNumber - 1

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

	self := service.bot.State.User

	if number < 0 {
		number = 0
	}

	_, err = service.database.History.CreateOne(
		db.History.UserID.Set(self.ID),
		db.History.Game.Link(db.Game.ID.Equals(game.ID)),
		db.History.Number.Set(number),
	).Exec(ctx)
	if err != nil {
		return game, started, fmt.Errorf("game: start: create history: %w", err)
	}

	if channel.Type == discordgo.ChannelTypeGuildText ||
		channel.Type == discordgo.ChannelTypeGuildPublicThread ||
		channel.Type == discordgo.ChannelTypeGuildPrivateThread {
		go func() {
			_, sendErr := service.bot.ChannelMessageSend(
				channelID,
				fmt.Sprintf(`**A new game has started!**
Start the count from **%d**`, number+1),
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
	shame ...*ShameOptions,
) (game *db.GameModel, err error) {
	hasShame := len(shame) > 0

	game, err = service.database.Game.FindUnique(
		db.Game.ID.Equals(gameID),
	).Update(
		db.Game.Status.Set(status),
	).Exec(ctx)
	if err != nil {
		return game, fmt.Errorf("game: end: update game: %w", err)
	}

	if hasShame {
		shame := shame[0]
		roleID, okRoleID := shame.settings.ShameRoleID()

		lastShameUserID, okLastShameUserID := shame.settings.LastShameUserID()
		if okLastShameUserID && okRoleID {
			go func() {
				utils.LogIfErr(
					utils.Logger,
					"guild-member-role-remove",
					service.bot.GuildMemberRoleRemove(
						shame.settings.GuildID,
						lastShameUserID,
						roleID,
					),
				)
			}()
		}

		if okRoleID {
			go func() {
				utils.LogIfErr(
					utils.Logger,
					"guild-member-role-add",
					service.bot.GuildMemberRoleAdd(
						shame.settings.GuildID,
						shame.message.Author.ID,
						roleID,
					),
				)
			}()
		}

		_, err = service.settings.Update(
			ctx,
			shame.settings.ID,
			db.Settings.LastShameUserID.Set(shame.message.Author.ID),
		)
		if err != nil {
			utils.Logger.Errorw(
				"game: end: update shame settings failed",
				"error",
				err,
				"guildID",
				shame.settings.GuildID,
				"gameID",
				gameID,
				"userID",
				shame.message.Author.ID,
			)
			return game, fmt.Errorf("game: end: update shame settings: %w", err)
		}
	}

	return game, err
}

func (service *GameService) ParseNumber(
	ctx context.Context,
	message *discordgo.Message,
	math bool,
) (i int, err error) {
	if message.Author.Bot {
		i = -1
		err = ErrAuthorIsBot

		return i, err
	}

	if !math {
		i, err = strconv.Atoi(message.Content)

		if i == 0 {
			i = -1
			err = ErrNumberCannotBeZero
		}

		return i, err
	}

	const maxExprLen = 256
	if len(message.Content) > maxExprLen {
		return 0, ErrExprTooLong
	}

	utils.Logger.With("Message", message.Content).Debug("Compiling expression")

	program, compileErr := expr.Compile(message.Content, expr.AsFloat64())
	if compileErr != nil {
		return 0, fmt.Errorf("services/game: compile expr: %w", compileErr)
	}

	utils.Logger.With("Message", message.Content).Debug("Evaluating expression")

	result, evalErr := expr.Run(program, nil)
	if evalErr != nil {
		return 0, fmt.Errorf("services/game: eval expr: %w", evalErr)
	}

	utils.Logger.With("Message", message.Content, "result", result).
		Debug("Evaluation result")

	parsedAsFloat, ok := result.(float64)
	if !ok {
		return 0, ErrExprNotNumber
	}

	i = int(parsedAsFloat)

	if i == 0 {
		i = -1
		err = ErrNumberCannotBeZero
	}

	return i, err
}

func (service *GameService) AddNumber(
	ctx context.Context,
	guildID string,
	number int,
	message *discordgo.Message,
	settings *db.SettingsModel,
) {
	game, exists, err := service.GetCurrentGame(ctx, guildID)
	if err != nil {
		utils.Logger.Errorw(
			"game: add number: get current game failed",
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
			"game: add number: get last history failed",
			"error",
			err,
			"guildID",
			guildID,
			"gameID",
			game.ID,
		)
		return
	}

	isNextNumber := number == history.Number+1
	isSameUser := message.Author.ID == history.UserID &&
		service.cfg.Env != "development"

	if !isNextNumber || isSameUser {
		// Build failure reason
		failReason := fmt.Sprintf(
			"<@%s> counted twice in a row!",
			message.Author.ID,
		)
		if !isNextNumber {
			failReason = fmt.Sprintf("%d is not the next number!", number)
		}

		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			service.bot.MessageReactionAdd(message.ChannelID, message.ID, "❌"),
		)

		saves, err := service.saves.GetSaves(ctx, settings, message.Author.ID)
		if err != nil {
			utils.Logger.Errorw(
				"game: add number: get saves failed",
				"error",
				err,
				"guildID",
				guildID,
				"userID",
				message.Author.ID,
				"messageID",
				message.ID,
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
					"game: add number: deduct player save failed",
					"error",
					err,
					"guildID",
					guildID,
					"userID",
					message.Author.ID,
				)
				return
			}

			go func() {
				_, sendErr := service.bot.ChannelMessageSendReply(
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
					"game: add number: deduct guild save failed",
					"error",
					err,
					"guildID",
					guildID,
					"userID",
					message.Author.ID,
				)
				return
			}

			go func() {
				_, sendErr := service.bot.ChannelMessageSendReply(
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

		isHighscore, _, err := service.checkStreak(ctx, settings, game, number)
		if err != nil {
			utils.Logger.Errorw(
				"game: add number: check streak failed",
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
		pointsRemoved := int(history.Number / 10)

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
			_, sendErr := service.bot.ChannelMessageSendReply(
				message.ChannelID,
				fmt.Sprintf(
					`%s
**The game has ended on a streak of %d!**%s%s

**Want to save the game?** Make sure to **/vote** for Kazu and earn yourself saves to save the game!`,
					failReason,
					number,
					highScoreText,
					pointsRemovedText,
				),
				message.Reference(),
			)
			utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
		}()

		shame := ShameOptions{
			message:  message,
			settings: settings,
		}
		if _, _, startErr := service.Start(
			ctx,
			guildID,
			db.GameTypeNormal,
			1,
			true,
			&shame,
		); startErr != nil {
			utils.Logger.Warnw(
				"game: add number: restart failed",
				"error", startErr,
				"guildID", guildID,
				"userID", message.Author.ID,
			)
		}

		return
	}

	cooldown, err := service.checkCooldown(ctx,
		message.Author.ID,
		game.ID,
		settings.Cooldown,
	)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		utils.Logger.Errorw(
			"game: add number: check cooldown failed",
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
		db.History.Number.Set(number),
		db.History.MessageID.Set(message.ID),
	).Exec(ctx)
	if err != nil {
		utils.Logger.Errorw(
			"game: add number: create history failed",
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
	isHighscore, isGameHighscored, err := service.checkStreak(
		ctx,
		settings,
		game,
		number,
	)
	if err != nil {
		utils.Logger.Errorw(
			"game: add number: check streak failed after history",
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
			service.bot.MessageReactionAdd(message.ChannelID, message.ID, "🎉"),
		)
		go func() {
			utils.LogIfErr(
				utils.Logger,
				"reset-shame",
				service.settings.ResetShame(ctx, guildID),
			)
		}()
	}

	emoji := "✅"
	if isHighscore {
		emoji = "☑️"
	}

	utils.LogIfErr(
		utils.Logger,
		"message-reaction-add",
		service.bot.MessageReactionAdd(message.ChannelID, message.ID, emoji),
	)
	service.checkSpecialReactions(message, number)
}

func (service *GameService) IsEqualToLast(
	ctx context.Context,
	message *discordgo.Message,
	settings *db.SettingsModel,
	isDelete bool,
) (ok bool, number int) {
	ok = true
	number = -1

	game, exists, err := service.GetCurrentGame(ctx, message.GuildID)
	if err != nil || !exists {
		utils.Logger.Info("Couldnt find game", err)
		return ok, number
	}

	history, _, err := service.GetLastHistory(ctx, game)
	if err != nil {
		utils.Logger.Info("Couldnt find last history", err)
		return ok, number
	}

	messageID, messageIDOk := history.MessageID()
	if !messageIDOk {
		return ok, number
	}

	if messageID != message.ID {
		return ok, number
	}

	number = history.Number

	if isDelete {
		ok = false
		return ok, number
	}

	parsedNumber, err := service.ParseNumber(ctx, message, settings.Math)
	if err != nil {
		ok = false
		return ok, number
	}

	utils.Logger.Info("Checking is equal", message.Content)

	if parsedNumber != number {
		ok = false
	}

	return ok, number
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

func (service *GameService) checkStreak(
	ctx context.Context,
	settings *db.SettingsModel,
	game *db.GameModel,
	number int,
) (isHighscore bool, isGameHighscored bool, err error) {
	if number <= settings.Highscore {
		return false, false, nil
	}

	isHighscore = true

	go service.settings.SetHighscoreByGuildID(ctx, settings.GuildID, number)

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
		cooldown = time.Now().Add(-time.Second * 600)
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
	if len(emoji) > 0 {
		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			service.bot.MessageReactionAdd(
				message.ChannelID,
				message.ID,
				emoji,
			),
		)
	}

	sentMessage, err := service.bot.ChannelMessageSendReply(
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
				service.bot.ChannelMessageDelete(
					sentMessage.ChannelID,
					sentMessage.ID,
				),
			)
		})
	}
}

// specialEmojisForNumber returns the extra emoji reactions a number earns.
// Pure function — no side effects; tested directly.
func specialEmojisForNumber(number int) []string {
	var emojis []string

	if number > 10 && utils.IsPalindrome(strconv.Itoa(number)) {
		emojis = append(emojis, "🪞")
	}

	switch number {
	case 4:
		emojis = append(emojis, "🍀")
	case 69:
		emojis = append(emojis, "niceone:1260697303224815696")
	case 100:
		emojis = append(emojis, "💯")
	case 360:
		emojis = append(emojis, "⚪")
	case 420:
		emojis = append(emojis, "🍃")
	case 666:
		emojis = append(emojis, "🤘")
	case 777:
		emojis = append(emojis, "🎰")
	case 1000:
		emojis = append(emojis, "1000:1262411624019525684")
	case 10_000:
		emojis = append(emojis, "10000:1262411765996851200")
	case 100_000:
		emojis = append(emojis, "100000:1262411649407647904")
	}

	return emojis
}

func (service *GameService) checkSpecialReactions(
	message *discordgo.Message,
	number int,
) {
	for _, emoji := range specialEmojisForNumber(number) {
		emoji := emoji
		go func() {
			utils.LogIfErr(
				utils.Logger,
				"message-reaction-add",
				service.bot.MessageReactionAdd(
					message.ChannelID,
					message.ID,
					emoji,
				),
			)
		}()
	}
}
