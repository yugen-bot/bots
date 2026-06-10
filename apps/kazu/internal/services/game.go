package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/expr-lang/expr"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"

	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/kazu/internal/ent/game"
	"jurien.dev/yugen/kazu/internal/ent/history"
	localStatic "jurien.dev/yugen/kazu/internal/static"
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
	client   *disgoplus.Bot
	cfg      *config.Config
	database *ent.Client
	settings *SettingsService
	saves    *SavesService
	points   *PointsService
}

func CreateGameService(container *di.Container) *GameService {
	utils.Logger.Info("Creating Game Service")

	return &GameService{
		client:   container.Get(static.DiBot).(*disgoplus.Bot),
		cfg:      container.Get(static.DiConfig).(*config.Config),
		database: container.Get(static.DiDatabase).(*ent.Client),
		settings: container.Get(static.DiSettings).(*SettingsService),
		saves:    container.Get(localStatic.DiSaves).(*SavesService),
		points:   container.Get(localStatic.DiPoints).(*PointsService),
	}
}

type ShameOptions struct {
	message  discord.Message
	settings *ent.Settings
}

func (s *GameService) Start(
	ctx context.Context,
	guildID string,
	gameType game.Type,
	startingNumber int,
	recreate bool,
	shame ...*ShameOptions,
) (g *ent.Game, started bool, err error) {
	utils.Logger.Infof("Trying to start a game for %s", guildID)

	currentGame, exists, err := s.GetCurrentGame(ctx, guildID)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Errorw(
			"game: start: get current game failed",
			"error", err,
			"guildID", guildID,
		)
		return g, started, err
	}

	guildSettings, err := s.settings.GetByGuildID(ctx, guildID)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: get settings failed",
			"error", err,
			"guildID", guildID,
		)
		return g, started, fmt.Errorf("game: start: get settings: %w", err)
	}

	channelID := guildSettings.ChannelID
	if channelID == nil {
		err = ErrNoChannelIDConfigured
		utils.Logger.Errorw(
			"game: start: no channel id configured",
			"error", err,
			"guildID", guildID,
		)
		return g, started, err
	}

	channelSnowflake, err := snowflake.Parse(*channelID)
	if err != nil {
		return g, started, fmt.Errorf("game: start: parse channel id: %w", err)
	}

	channel, err := s.client.Client().Rest.GetChannel(channelSnowflake)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: get channel failed",
			"error", err,
			"guildID", guildID,
			"channelID", *channelID,
		)
		return g, started, fmt.Errorf("game: start: get channel: %w", err)
	}

	if exists && !recreate {
		return g, started, err
	}

	if (exists && recreate) || (exists && currentGame.Type != gameType) {
		if _, endErr := s.End(
			ctx,
			currentGame.ID,
			game.StatusFAILED,
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

	g, err = s.database.Game.Create().
		SetGuildID(guildID).
		SetType(gameType).
		Save(ctx)
	if err != nil {
		utils.Logger.Errorw(
			"game: start: create game failed",
			"error", err,
			"guildID", guildID,
		)
		return g, started, fmt.Errorf("game: start: create game: %w", err)
	}

	self, _ := s.client.Client().Caches.SelfUser()

	if number < 0 {
		number = 0
	}

	_, err = s.database.History.Create().
		SetUserID(self.ID.String()).
		SetGameID(g.ID).
		SetNumber(number).
		Save(ctx)
	if err != nil {
		return g, started, fmt.Errorf("game: start: create history: %w", err)
	}

	channelType := channel.Type()
	if channelType == discord.ChannelTypeGuildText ||
		channelType == discord.ChannelTypeGuildPublicThread ||
		channelType == discord.ChannelTypeGuildPrivateThread {
		go func() {
			_, sendErr := s.client.Client().Rest.CreateMessage(
				channelSnowflake,
				discord.MessageCreate{
					Content: fmt.Sprintf(`**A new game has started!**
Start the count from **%d**`, number+1),
				},
			)
			utils.LogIfErr(utils.Logger, "channel-message-send", sendErr)
		}()
	}

	return g, started, err
}

func (s *GameService) End(
	ctx context.Context,
	gameID int,
	status game.Status,
	shame ...*ShameOptions,
) (g *ent.Game, err error) {
	hasShame := len(shame) > 0

	g, err = s.database.Game.UpdateOneID(gameID).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return g, fmt.Errorf("game: end: update game: %w", err)
	}

	if hasShame {
		shameOpts := shame[0]
		roleID := shameOpts.settings.ShameRoleID
		okRoleID := roleID != nil

		lastShameUserID := shameOpts.settings.LastShameUserID
		okLastShameUserID := lastShameUserID != nil

		guildSnowflake, parseErr := snowflake.Parse(shameOpts.settings.GuildID)
		if parseErr != nil {
			utils.Logger.Warnw("game: end: parse guild id failed", "error", parseErr, "guildID", shameOpts.settings.GuildID)
		} else {
			if okLastShameUserID && okRoleID {
				lastUserSnowflake, lastUserErr := snowflake.Parse(*lastShameUserID)
				roleSnowflake, roleErr := snowflake.Parse(*roleID)
				if lastUserErr == nil && roleErr == nil {
					go func() {
						utils.LogIfErr(
							utils.Logger,
							"guild-member-role-remove",
							s.client.Client().Rest.RemoveMemberRole(guildSnowflake, lastUserSnowflake, roleSnowflake),
						)
					}()
				}
			}

			if okRoleID {
				authorSnowflake, authorErr := snowflake.Parse(shameOpts.message.Author.ID.String())
				roleSnowflake, roleErr := snowflake.Parse(*roleID)
				if authorErr == nil && roleErr == nil {
					go func() {
						utils.LogIfErr(
							utils.Logger,
							"guild-member-role-add",
							s.client.Client().Rest.AddMemberRole(guildSnowflake, authorSnowflake, roleSnowflake),
						)
					}()
				}
			}
		}

		_, err = s.settings.Update(
			ctx,
			shameOpts.settings.ID,
			func(u *ent.SettingsUpdateOne) {
				u.SetLastShameUserID(shameOpts.message.Author.ID.String())
			},
		)
		if err != nil {
			utils.Logger.Errorw(
				"game: end: update shame settings failed",
				"error", err,
				"guildID", shameOpts.settings.GuildID,
				"gameID", gameID,
				"userID", shameOpts.message.Author.ID.String(),
			)
			return g, fmt.Errorf("game: end: update shame settings: %w", err)
		}
	}

	return g, err
}

func (s *GameService) ParseNumber(
	ctx context.Context,
	message discord.Message,
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

func (s *GameService) AddNumber(
	ctx context.Context,
	guildID string,
	number int,
	message discord.Message,
	guildSettings *ent.Settings,
) {
	g, exists, err := s.GetCurrentGame(ctx, guildID)
	if err != nil {
		utils.Logger.Errorw(
			"game: add number: get current game failed",
			"error", err,
			"guildID", guildID,
		)
		return
	}

	if !exists {
		return
	}

	h, _, err := s.GetLastHistory(ctx, g)
	if err != nil {
		utils.Logger.Errorw(
			"game: add number: get last history failed",
			"error", err,
			"guildID", guildID,
			"gameID", g.ID,
		)
		return
	}

	isNextNumber := number == h.Number+1
	isSameUser := message.Author.ID.String() == h.UserID &&
		s.cfg.Env != "development"

	if !isNextNumber || isSameUser {
		// Build failure reason
		failReason := fmt.Sprintf(
			"<@%s> counted twice in a row!",
			message.Author.ID.String(),
		)
		if !isNextNumber {
			failReason = fmt.Sprintf("%d is not the next number!", number)
		}

		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			s.client.Client().Rest.AddReaction(message.ChannelID, message.ID, "❌"),
		)

		saves, err := s.saves.GetSaves(ctx, guildSettings, message.Author.ID.String())
		if err != nil {
			utils.Logger.Errorw(
				"game: add number: get saves failed",
				"error", err,
				"guildID", guildID,
				"userID", message.Author.ID.String(),
				"messageID", message.ID.String(),
			)
			return
		}

		if saves.player >= 1 {
			leftoverSaves, maxSaves, err := s.saves.DeductSaveFromPlayer(
				ctx,
				message.Author.ID.String(),
				1,
			)
			if err != nil {
				utils.Logger.Errorw(
					"game: add number: deduct player save failed",
					"error", err,
					"guildID", guildID,
					"userID", message.Author.ID.String(),
				)
				return
			}

			msgID := message.ID
			channelID := message.ChannelID
			guildIDPtr := message.GuildID

			go func() {
				_, sendErr := s.client.Client().Rest.CreateMessage(
					channelID,
					discord.MessageCreate{
						Content: fmt.Sprintf(
							`%s
Used **1 of your own** saves, You have **%s/%s** saves left.`,
							failReason,
							strconv.FormatFloat(leftoverSaves, 'f', -1, 64),
							strconv.FormatFloat(maxSaves, 'f', -1, 64),
						),
						MessageReference: &discord.MessageReference{
							MessageID: &msgID,
							ChannelID: &channelID,
							GuildID:   guildIDPtr,
						},
					},
				)
				utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
			}()

			return
		}

		if saves.guild >= 1 {
			leftoverSaves, maxSaves, err := s.saves.DeductSaveFromGuild(
				ctx,
				message.GuildID.String(),
				guildSettings,
				1,
			)
			if err != nil {
				utils.Logger.Errorw(
					"game: add number: deduct guild save failed",
					"error", err,
					"guildID", guildID,
					"userID", message.Author.ID.String(),
				)
				return
			}

			msgID := message.ID
			channelID := message.ChannelID
			guildIDPtr := message.GuildID

			go func() {
				_, sendErr := s.client.Client().Rest.CreateMessage(
					channelID,
					discord.MessageCreate{
						Content: fmt.Sprintf(
							`%s
Used **1 server** save, There are **%s/%s** server saves left.`,
							failReason,
							strconv.FormatFloat(leftoverSaves, 'f', -1, 64),
							strconv.FormatFloat(maxSaves, 'f', -1, 64),
						),
						MessageReference: &discord.MessageReference{
							MessageID: &msgID,
							ChannelID: &channelID,
							GuildID:   guildIDPtr,
						},
					},
				)
				utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
			}()

			return
		}

		isHighscore, _, err := s.checkStreak(ctx, guildSettings, g, number)
		if err != nil {
			utils.Logger.Errorw(
				"game: add number: check streak failed",
				"error", err,
				"guildID", guildID,
				"gameID", g.ID,
			)
			return
		}

		highScoreText := ""
		if isHighscore {
			highScoreText = "\n**A new highscore has been set! 🎉**"
		}

		// Deduct points from the player who broke the chain
		pointsRemoved := int(h.Number / 10)

		go func() {
			utils.LogIfErr(
				utils.Logger,
				"remove-game-points",
				s.points.RemoveGamePoints(
					ctx,
					guildID,
					message.Author.ID.String(),
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

		msgID := message.ID
		channelID := message.ChannelID
		guildIDPtr := message.GuildID

		go func() {
			_, sendErr := s.client.Client().Rest.CreateMessage(
				channelID,
				discord.MessageCreate{
					Content: fmt.Sprintf(
						`%s
**The game has ended on a streak of %d!**%s%s

**Want to save the game?** Make sure to **/vote** for Kazu and earn yourself saves to save the game!`,
						failReason,
						number,
						highScoreText,
						pointsRemovedText,
					),
					MessageReference: &discord.MessageReference{
						MessageID: &msgID,
						ChannelID: &channelID,
						GuildID:   guildIDPtr,
					},
				},
			)
			utils.LogIfErr(utils.Logger, "channel-message-send-reply", sendErr)
		}()

		shameOpts := ShameOptions{
			message:  message,
			settings: guildSettings,
		}
		if _, _, startErr := s.Start(
			ctx,
			guildID,
			game.TypeNORMAL,
			1,
			true,
			&shameOpts,
		); startErr != nil {
			utils.Logger.Warnw(
				"game: add number: restart failed",
				"error", startErr,
				"guildID", guildID,
				"userID", message.Author.ID.String(),
			)
		}

		return
	}

	cooldown, err := s.checkCooldown(ctx,
		message.Author.ID.String(),
		g.ID,
		guildSettings.Cooldown,
	)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Errorw(
			"game: add number: check cooldown failed",
			"error", err,
			"guildID", guildID,
			"gameID", g.ID,
			"userID", message.Author.ID.String(),
		)
		return
	}

	if cooldown.After(time.Now()) {
		go s.replyAndDelete(
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
			s.points.AddGamePoints(ctx, guildID, message.Author.ID.String(), 1),
		)
	}()

	msgIDStr := message.ID.String()
	_, err = s.database.History.Create().
		SetUserID(message.Author.ID.String()).
		SetGameID(g.ID).
		SetNumber(number).
		SetNillableMessageID(&msgIDStr).
		Save(ctx)
	if err != nil {
		utils.Logger.Errorw(
			"game: add number: create history failed",
			"error", err,
			"guildID", guildID,
			"gameID", g.ID,
			"userID", message.Author.ID.String(),
			"messageID", message.ID.String(),
		)
		return
	}

	// Check streak and react
	isHighscore, isGameHighscored, err := s.checkStreak(
		ctx,
		guildSettings,
		g,
		number,
	)
	if err != nil {
		utils.Logger.Errorw(
			"game: add number: check streak failed after history",
			"error", err,
			"guildID", guildID,
			"gameID", g.ID,
		)
	}

	if isGameHighscored {
		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			s.client.Client().Rest.AddReaction(message.ChannelID, message.ID, "🎉"),
		)
		go func() {
			utils.LogIfErr(
				utils.Logger,
				"reset-shame",
				s.settings.ResetShame(ctx, guildID),
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
		s.client.Client().Rest.AddReaction(message.ChannelID, message.ID, emoji),
	)
	s.checkSpecialReactions(message, number)
}

func (s *GameService) IsEqualToLast(
	ctx context.Context,
	message discord.Message,
	guildSettings *ent.Settings,
	isDelete bool,
) (ok bool, number int) {
	ok = true
	number = -1

	// nil guard: zero ID means no message
	if message.ID == 0 {
		return ok, number
	}

	guildID := ""
	if message.GuildID != nil {
		guildID = message.GuildID.String()
	}

	g, exists, err := s.GetCurrentGame(ctx, guildID)
	if err != nil || !exists {
		utils.Logger.Info("Couldnt find game", err)
		return ok, number
	}

	h, _, err := s.GetLastHistory(ctx, g)
	if err != nil {
		utils.Logger.Info("Couldnt find last history", err)
		return ok, number
	}

	if h.MessageID == nil {
		return ok, number
	}

	if *h.MessageID != message.ID.String() {
		return ok, number
	}

	number = h.Number

	if isDelete {
		ok = false
		return ok, number
	}

	parsedNumber, err := s.ParseNumber(ctx, message, guildSettings.Math)
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

func (s *GameService) checkStreak(
	ctx context.Context,
	guildSettings *ent.Settings,
	g *ent.Game,
	number int,
) (isHighscore bool, isGameHighscored bool, err error) {
	if number <= guildSettings.Highscore {
		return false, false, nil
	}

	isHighscore = true

	go s.settings.SetHighscoreByGuildID(ctx, guildSettings.GuildID, number) //nolint:errcheck

	if g.IsHighscored {
		return isHighscore, false, nil
	}

	isGameHighscored = true

	go s.database.Game.UpdateOneID(g.ID). //nolint:errcheck
						SetIsHighscored(true).
						Save(ctx)

	return isHighscore, isGameHighscored, nil
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
	if len(emoji) > 0 {
		utils.LogIfErr(
			utils.Logger,
			"message-reaction-add",
			s.client.Client().Rest.AddReaction(
				message.ChannelID,
				message.ID,
				emoji,
			),
		)
	}

	msgID := message.ID
	channelID := message.ChannelID
	guildIDPtr := message.GuildID

	sentMessage, sendErr := s.client.Client().Rest.CreateMessage(
		message.ChannelID,
		discord.MessageCreate{
			Content: messageToSend,
			MessageReference: &discord.MessageReference{
				MessageID: &msgID,
				ChannelID: &channelID,
				GuildID:   guildIDPtr,
			},
		},
	)
	if sendErr != nil {
		utils.Logger.Errorw(
			"game: reply and delete: send reply failed",
			"error", sendErr,
			"channelID", message.ChannelID.String(),
			"messageID", message.ID.String(),
		)
		return
	}

	if deleteAfter {
		sentMsgID := sentMessage.ID
		sentChannelID := sentMessage.ChannelID
		time.AfterFunc(time.Second*5, func() {
			utils.LogIfErr(
				utils.Logger,
				"channel-message-delete",
				s.client.Client().Rest.DeleteMessage(
					sentChannelID,
					sentMsgID,
				),
			)
		})
	}
}

func (s *GameService) CountByGuildIDs(ctx context.Context, guildIDs []string) (games int, h int, err error) {
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

func (s *GameService) DeleteByGuildIDs(ctx context.Context, guildIDs []string) (games int, h int, err error) {
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

func (s *GameService) FindAllGuildIDs(ctx context.Context) ([]string, error) {
	return s.database.Game.Query().
		Unique(true).
		Select(game.FieldGuildID).
		Strings(ctx)
}

// specialEmojisForNumber returns the extra emoji reactions a number earns.
// Pure function — no side effects; tested directly.
func (s *GameService) checkSpecialReactions(
	message discord.Message,
	number int,
) {
	for _, emoji := range specialEmojisForNumber(number) {
		emoji := emoji
		go func() {
			utils.LogIfErr(
				utils.Logger,
				"message-reaction-add",
				s.client.Client().Rest.AddReaction(
					message.ChannelID,
					message.ID,
					emoji,
				),
			)
		}()
	}
}

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
