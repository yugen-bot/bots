package services

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/ent"
	"jurien.dev/yugen/hoshi/internal/ent/starboardlog"
	"jurien.dev/yugen/hoshi/internal/ent/starboards"
	"jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/config"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type StarboardService struct {
	database *ent.Client
	settings *SettingsService
	client   *bot.Client
	cfg      *config.Config
}

func CreateStarboardService(container *di.Container) *StarboardService {
	utils.Logger.Info("Creating Starboard Service")

	return &StarboardService{
		database: container.Get(sharedStatic.DiDatabase).(*ent.Client),
		settings: container.Get(sharedStatic.DiSettings).(*SettingsService),
		client:   container.Get(sharedStatic.DiBot).(*disgoplus.Bot).Client(),
		cfg:      container.Get(sharedStatic.DiConfig).(*config.Config),
	}
}

func (s *StarboardService) CheckReaction(
	ctx context.Context,
	channelID, messageID, guildID, emojiID, emojiName string,
) {
	sett, err := s.settings.GetByGuildID(ctx, guildID)
	if err != nil {
		if !ent.IsNotFound(err) {
			utils.Logger.Errorw("starboard: check reaction: get settings failed",
				"error", err, "guildID", guildID)
		}

		return
	}

	chID, _ := snowflake.Parse(channelID)

	// Resolve parent channel when reacting inside a thread.
	sourceChannelID := channelID

	ch, chErr := s.client.Rest.GetChannel(chID)
	if chErr == nil && ch != nil &&
		(ch.Type() == discord.ChannelTypeGuildPublicThread ||
			ch.Type() == discord.ChannelTypeGuildPrivateThread) {
		if thread, ok := ch.(interface{ ParentID() *snowflake.ID }); ok {
			if pid := thread.ParentID(); pid != nil {
				sourceChannelID = pid.String()
			}
		}
	}

	if slices.Contains([]string(sett.IgnoredChannelIds), sourceChannelID) {
		return
	}

	msgID, _ := snowflake.Parse(messageID)

	isTarget, err := s.getLogByMessageID(ctx, messageID)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Errorw("starboard: check reaction: get log by message id failed",
			"error", err, "guildID", guildID, "channelID", channelID, "messageID", messageID)

		return
	}

	if isTarget != nil {
		return
	}

	reactionEmoji := emojiID
	if reactionEmoji == "" {
		reactionEmoji = emojiName
	}

	configurations, err := s.database.Starboards.Query().Where(
		starboards.GuildIDEQ(guildID),
		starboards.SourceEmojiEQ(reactionEmoji),
		starboards.Or(
			starboards.SourceChannelIDEQ(sourceChannelID),
			starboards.SourceChannelIDIsNil(),
		),
	).All(ctx)

	if err != nil || len(configurations) == 0 {
		return
	}

	cfg := configurations[0]
	for _, c := range configurations {
		if c.SourceChannelID != nil && *c.SourceChannelID == sourceChannelID {
			cfg = c
			break
		}
	}

	emojiAPIName := emojiName
	if emojiID != "" {
		emojiAPIName = fmt.Sprintf("%s:%s", emojiName, emojiID)
	}

	users, err := s.client.Rest.GetReactions(chID, msgID, emojiAPIName, discord.MessageReactionTypeNormal, 0, 100)
	if err != nil {
		utils.Logger.Errorw("starboard: check reaction: get message reactions failed",
			"error", err, "guildID", guildID, "channelID", channelID, "messageID", messageID)

		return
	}

	msg, err := s.client.Rest.GetMessage(chID, msgID)
	if err != nil || msg == nil {
		utils.Logger.Errorw("starboard: check reaction: get channel message failed",
			"error", err, "guildID", guildID, "channelID", channelID, "messageID", messageID)

		return
	}

	allowSelf := sett.Self

	count := 0

	for _, u := range users {
		if u.Bot {
			continue
		}

		if !allowSelf && u.ID == msg.Author.ID {
			continue
		}

		count++
	}

	log, err := s.getLogByOriginalID(ctx, messageID)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Errorw("starboard: check reaction: get log by original id failed",
			"error", err, "guildID", guildID, "channelID", channelID, "messageID", messageID)

		return
	}

	if count == 0 && log != nil {
		s.deleteStarboard(ctx, log)
		return
	}

	if count < sett.Treshold {
		return
	}

	if msg.GuildID == nil {
		gid, _ := snowflake.Parse(guildID)
		msg.GuildID = &gid
	}

	embeds := s.createEmbeds(msg, guildID)
	if len(embeds) == 0 {
		return
	}

	if log != nil {
		s.updateStarboard(count, embeds, msg, guildID, emojiName, emojiID, log)
		return
	}

	s.createStarboard(
		ctx, count, embeds, channelID, messageID, msg,
		guildID, cfg.TargetChannelID, emojiName, emojiID,
	)
}

func (s *StarboardService) getLogByOriginalID(ctx context.Context, id string) (*ent.StarboardLog, error) {
	result, err := s.database.StarboardLog.Query().
		Where(starboardlog.OriginalMessageIDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("starboard: get log by original id: %w", err)
	}

	return result, nil
}

func (s *StarboardService) getLogByMessageID(ctx context.Context, id string) (*ent.StarboardLog, error) {
	result, err := s.database.StarboardLog.Query().
		Where(starboardlog.MessageIDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("starboard: get log by message id: %w", err)
	}

	return result, nil
}

func (s *StarboardService) GetStarboardBySourceIDAndEmoji(
	ctx context.Context,
	guildID, sourceEmoji string,
	sourceChannelID *string,
) (*ent.Starboards, error) {
	q := s.database.Starboards.Query().Where(
		starboards.GuildIDEQ(guildID),
		starboards.SourceEmojiEQ(sourceEmoji),
	)

	if sourceChannelID != nil {
		q = q.Where(starboards.SourceChannelIDEQ(*sourceChannelID))
	} else {
		q = q.Where(starboards.SourceChannelIDIsNil())
	}

	result, err := q.First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("starboard: get by source and emoji: %w", err)
	}

	return result, nil
}

func (s *StarboardService) GetStarboards(
	ctx context.Context,
	guildID string,
	page int,
) ([]*ent.Starboards, int, error) {
	where := starboards.GuildIDEQ(guildID)
	skip := (page - 1) * 10

	items, err := s.database.Starboards.Query().
		Where(where).
		Offset(skip).
		Limit(10).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("starboard: get starboards page: %w", err)
	}

	total, err := s.database.Starboards.Query().Where(where).Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("starboard: get starboards total: %w", err)
	}

	return items, total, nil
}

func (s *StarboardService) AddStarboard(
	ctx context.Context,
	guildID, sourceEmoji string,
	sourceChannelID *string,
	targetChannelID string,
) (*ent.Starboards, error) {
	c := s.database.Starboards.Create().
		SetGuildID(guildID).
		SetTargetChannelID(targetChannelID).
		SetSourceEmoji(sourceEmoji)

	if sourceChannelID != nil {
		c = c.SetSourceChannelID(*sourceChannelID)
	}

	result, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("starboard: add starboard: %w", err)
	}

	return result, nil
}

func (s *StarboardService) RemoveStarboardByID(
	ctx context.Context,
	guildID string,
	id int,
) (*ent.Starboards, error) {
	sb, err := s.database.Starboards.Query().Where(
		starboards.GuildIDEQ(guildID),
		starboards.IDEQ(id),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("starboard: remove by id find: %w", err)
	}

	if err := s.database.Starboards.DeleteOneID(sb.ID).Exec(ctx); err != nil {
		return nil, fmt.Errorf("starboard: remove by id delete: %w", err)
	}

	return sb, nil
}

func (s *StarboardService) createEmbeds(msg *discord.Message, guildID string) []discord.Embed {
	if len(msg.Content) == 0 && len(msg.Attachments) == 0 {
		return nil
	}

	var footerIconURL string
	if ownerID, err := snowflake.Parse(s.cfg.OwnerID); err == nil {
		if owner, err := s.client.Rest.GetUser(ownerID); err == nil && owner != nil {
			footerIconURL = owner.EffectiveAvatarURL()
		}
	}

	botName := "me"
	if self, ok := s.client.Caches.SelfUser(); ok {
		botName = self.Username
	}

	footer := discord.EmbedFooter{
		Text:    fmt.Sprintf("Like %s? Please vote using /vote!", botName),
		IconURL: footerIconURL,
	}

	chunks := utils.SplitBySentence(msg.Content, static.MaxEmbedDescription)
	if len(chunks) == 0 {
		chunks = []string{""}
	}

	embeds := make([]discord.Embed, 0, len(chunks))

	for i, chunk := range chunks {
		e := discord.NewEmbed().WithColor(static.EmbedColor)

		if i == 0 && msg.Author.ID != 0 {
			e = e.WithAuthor(msg.Author.Username, msg.Author.EffectiveAvatarURL(), "")
		}

		if chunk != "" {
			e = e.WithDescription(chunk)
		}

		if i == len(chunks)-1 {
			e = e.WithEmbedFooter(&footer)

			if len(msg.Attachments) > 0 {
				e = e.WithImage(msg.Attachments[0].URL)
			}
		}

		embeds = append(embeds, e)
	}

	return embeds
}

func (s *StarboardService) createContentString(
	count int,
	guildID string,
	emojiName, emojiID string,
	msg *discord.Message,
) string {
	display := emojiName
	if emojiID != "" {
		display = fmt.Sprintf("<:%s:%s>", emojiName, emojiID)
	}

	return fmt.Sprintf("**%d %s** at https://discord.com/channels/%s/%s/%s",
		count, display, guildID, msg.ChannelID, msg.ID)
}

func emojiAPIFormat(emojiName, emojiID string) string {
	if emojiID != "" {
		return fmt.Sprintf("%s:%s", emojiName, emojiID)
	}

	return emojiName
}

func (s *StarboardService) createStarboard(
	ctx context.Context,
	count int,
	embeds []discord.Embed,
	sourceChannelID string,
	originalMessageID string,
	msg *discord.Message,
	guildID string,
	targetChannelID, emojiName, emojiID string,
) {
	targetChID, err := snowflake.Parse(targetChannelID)
	if err != nil {
		return
	}

	sourceChID, _ := snowflake.Parse(sourceChannelID)
	origMsgID, _ := snowflake.Parse(originalMessageID)

	sent, err := s.client.Rest.CreateMessage(targetChID, discord.MessageCreate{
		Content: s.createContentString(count, guildID, emojiName, emojiID, msg),
		Embeds:  embeds,
	})
	if err != nil {
		utils.Logger.Errorw("starboard: create starboard: send message failed",
			"error", err,
			"guildID", guildID,
			"channelID", msg.ChannelID,
			"messageID", msg.ID,
			"targetChannelID", targetChannelID,
		)

		var errType string
		if rest.IsJSONErrorCode(err, rest.JSONErrorCodeUnknownChannel) {
			errType = "unknown_channel"
		} else if rest.IsJSONErrorCode(err, rest.JSONErrorCodeLackPermissionsToPerformAction) {
			errType = "forbidden_channel"
		}

		if errType != "" {
			originalMessage, fetchErr := s.client.Rest.GetMessage(sourceChID, origMsgID)
			if fetchErr != nil || originalMessage == nil {
				return
			}

			message := "The starboard channel does not seem to exist: <#%s>.\nPlease inform a moderator of this server."
			if errType == "forbidden_channel" {
				message = "Hoshi does not have permissions to access the starboard channel: <#%s>.\nPlease inform a moderator of this server."
			}

			_, _ = s.client.Rest.CreateMessage(sourceChID, discord.MessageCreate{
				Content: fmt.Sprintf(message, targetChannelID),
				MessageReference: &discord.MessageReference{
					MessageID: &originalMessage.ID,
					ChannelID: &originalMessage.ChannelID,
					GuildID:   originalMessage.GuildID,
				},
			})
		}

		return
	}

	result, err := s.database.StarboardLog.Create().
		SetGuildID(guildIDString(msg.GuildID)).
		SetChannelID(targetChannelID).
		SetMessageID(sent.ID.String()).
		SetOriginalMessageID(msg.ID.String()).
		Save(ctx)
	if err != nil {
		utils.Logger.Errorw("starboard: create log entry failed",
			"error", fmt.Errorf("starboard: create log: %w", err),
			"guildID", guildID,
			"channelID", msg.ChannelID,
			"messageID", msg.ID,
			"targetChannelID", targetChannelID,
		)

		return
	}

	utils.LogIfErrNoRateLimit(utils.Logger, "message-reaction-add",
		s.client.Rest.AddReaction(targetChID, sent.ID, emojiAPIFormat(emojiName, emojiID)))
	utils.LogIfErrNoRateLimit(utils.Logger, "message-reaction-add",
		s.client.Rest.AddReaction(msg.ChannelID, msg.ID, "🌟"))

	utils.Logger.Infow("starboard: create starboard: created new starboard entry",
		"starboardID", result.ID,
		"guildID", guildID,
		"channelID", targetChannelID,
		"messageID", sent.ID,
	)
}

func (s *StarboardService) updateStarboard(
	count int,
	embeds []discord.Embed,
	msg *discord.Message,
	guildID string,
	emojiName, emojiID string,
	log *ent.StarboardLog,
) {
	chID, err := snowflake.Parse(log.ChannelID)
	if err != nil {
		return
	}

	msgID, err := snowflake.Parse(log.MessageID)
	if err != nil {
		return
	}

	content := s.createContentString(count, guildID, emojiName, emojiID, msg)

	_, err = s.client.Rest.UpdateMessage(chID, msgID, discord.MessageUpdate{
		Content: &content,
		Embeds:  &embeds,
	})
	if err != nil {
		utils.Logger.Errorw("starboard: update starboard: edit message failed",
			"error", err,
			"guildID", guildIDString(msg.GuildID),
			"channelID", log.ChannelID,
			"messageID", log.MessageID,
		)
	}

	utils.Logger.Infow("starboard: update starboard: updated starboard entry",
		"starboardID", log.ID,
		"guildID", guildID,
		"channelID", log.ChannelID,
		"messageID", log.MessageID,
	)
}

func (s *StarboardService) deleteStarboard(ctx context.Context, log *ent.StarboardLog) {
	chID, err := snowflake.Parse(log.ChannelID)
	if err != nil {
		return
	}

	msgID, err := snowflake.Parse(log.MessageID)
	if err != nil {
		return
	}

	if err := s.client.Rest.DeleteMessage(chID, msgID); err != nil {
		utils.Logger.Errorw("starboard: delete starboard: delete message failed",
			"error", err,
			"guildID", log.GuildID,
			"channelID", log.ChannelID,
			"messageID", log.MessageID,
		)

		return
	}

	_, err = s.database.StarboardLog.Delete().
		Where(starboardlog.MessageIDEQ(log.MessageID)).
		Exec(ctx)
	if err != nil {
		utils.Logger.Errorw("starboard: delete log entry failed",
			"error", fmt.Errorf("starboard: delete log: %w", err),
			"guildID", log.GuildID,
			"channelID", log.ChannelID,
			"messageID", log.MessageID,
		)
	}
}

func (s *StarboardService) FindAllGuildIDs(ctx context.Context) ([]string, error) {
	guildIDs, err := s.database.Starboards.Query().
		Unique(true).
		Select(starboards.FieldGuildID).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("starboard: find distinct guild ids: %w", err)
	}

	return guildIDs, nil
}

func (s *StarboardService) FindByGuildIDs(ctx context.Context, guildIDs []string) ([]*ent.Starboards, error) {
	result, err := s.database.Starboards.Query().
		Where(starboards.GuildIDIn(guildIDs...)).
		All(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("starboard: find by guild ids: %w", err)
	}

	return result, nil
}

func (s *StarboardService) DeleteByGuildIDs(ctx context.Context, guildIDs []string) (int, error) {
	n, err := s.database.Starboards.Delete().
		Where(starboards.GuildIDIn(guildIDs...)).
		Exec(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, fmt.Errorf("starboard: delete by guild ids: %w", err)
	}

	return n, nil
}

func guildIDString(id *snowflake.ID) string {
	if id == nil {
		return ""
	}

	return id.String()
}

