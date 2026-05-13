package services

import (
	"context"
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/hoshi/prisma/db"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type StarboardService struct {
	database *db.PrismaClient
	settings *SettingsService
	bot      *discordgoplus.Bot
}

func CreateStarboardService(container *di.Container) *StarboardService {
	utils.Logger.Info("Creating Starboard Service")
	return &StarboardService{
		database: container.Get(sharedStatic.DiDatabase).(*db.PrismaClient),
		settings: container.Get(sharedStatic.DiSettings).(*SettingsService),
		bot:      container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

func (s *StarboardService) CheckReaction(channelID, messageID, guildID, emojiID, emojiName string) {
	settings, err := s.settings.GetByGuildID(guildID)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	sourceChannelID := channelID
	ch, err := s.bot.Channel(channelID)
	if err == nil && ch != nil {
		if ch.Type == discordgo.ChannelTypeGuildPublicThread || ch.Type == discordgo.ChannelTypeGuildPrivateThread {
			sourceChannelID = ch.ParentID
		}
	}

	if slices.Contains(settings.IgnoredChannelIds, sourceChannelID) {
		return
	}

	isTarget, _ := s.getLogByMessageID(messageID)
	if isTarget != nil {
		return
	}

	reactionEmoji := emojiID
	if reactionEmoji == "" {
		reactionEmoji = emojiName
	}

	ctx := context.Background()
	configurations, err := s.database.Starboards.FindMany(
		db.Starboards.GuildID.Equals(guildID),
		db.Starboards.SourceEmoji.Equals(reactionEmoji),
		db.Starboards.Or(
			db.Starboards.And(
				db.Starboards.SourceChannelID.Equals(sourceChannelID),
			),
			db.Starboards.And(
				db.Starboards.SourceChannelID.IsNull(),
			),
		),
	).Exec(ctx)

	if err != nil || len(configurations) == 0 {
		return
	}

	config := configurations[0]
	for _, c := range configurations {
		if sID, ok := c.SourceChannelID(); ok && sID == sourceChannelID {
			config = c
			break
		}
	}

	emojiAPIName := emojiName
	if emojiID != "" {
		emojiAPIName = fmt.Sprintf("%s:%s", emojiName, emojiID)
	}

	users, err := s.bot.MessageReactions(channelID, messageID, emojiAPIName, 100, "", "")
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	msg, err := s.bot.ChannelMessage(channelID, messageID)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	allowSelf := settings.Self
	authorID := ""
	if msg.Author != nil {
		authorID = msg.Author.ID
	}

	count := 0
	for _, u := range users {
		if u.Bot {
			continue
		}
		if !allowSelf && u.ID == authorID {
			continue
		}
		count++
	}

	log, _ := s.getLogByOriginalID(messageID)

	if count == 0 && log != nil {
		s.deleteStarboard(log)
		return
	}

	if count < settings.Treshold {
		return
	}

	embed := s.createEmbed(msg)
	if embed == nil {
		return
	}

	displayEmoji := reactionEmoji

	if log != nil {
		s.updateStarboard(count, embed, msg, displayEmoji, log)
		return
	}

	s.createStarboard(count, embed, msg, config.TargetChannelID, displayEmoji)
}

func (s *StarboardService) getLogByOriginalID(id string) (*db.LogModel, error) {
	result, err := s.database.Log.FindUnique(
		db.Log.OriginalMessageID.Equals(id),
	).Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("starboard: get log by original id: %w", err)
	}
	return result, nil
}

func (s *StarboardService) getLogByMessageID(id string) (*db.LogModel, error) {
	result, err := s.database.Log.FindUnique(
		db.Log.MessageID.Equals(id),
	).Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("starboard: get log by message id: %w", err)
	}
	return result, nil
}

func (s *StarboardService) GetStarboardBySourceIDAndEmoji(guildID, sourceEmoji string, sourceChannelID *string) (*db.StarboardsModel, error) {
	var params []db.StarboardsWhereParam
	params = append(params,
		db.Starboards.GuildID.Equals(guildID),
		db.Starboards.SourceEmoji.Equals(sourceEmoji),
	)
	if sourceChannelID != nil {
		params = append(params, db.Starboards.SourceChannelID.Equals(*sourceChannelID))
	} else {
		params = append(params, db.Starboards.SourceChannelID.IsNull())
	}

	result, err := s.database.Starboards.FindFirst(params...).Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("starboard: get by source and emoji: %w", err)
	}
	return result, nil
}

func (s *StarboardService) GetStarboards(guildID string, page int) ([]db.StarboardsModel, int, error) {
	ctx := context.Background()
	where := []db.StarboardsWhereParam{db.Starboards.GuildID.Equals(guildID)}
	skip := (page - 1) * 10

	items, err := s.database.Starboards.FindMany(where...).Skip(skip).Take(10).Exec(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("starboard: get starboards page: %w", err)
	}

	total, err := s.database.Starboards.FindMany(where...).Exec(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("starboard: get starboards total: %w", err)
	}

	return items, len(total), nil
}

func (s *StarboardService) AddStarboard(guildID, sourceEmoji string, sourceChannelID *string, targetChannelID string) (*db.StarboardsModel, error) {
	optional := []db.StarboardsSetParam{
		db.Starboards.SourceEmoji.Set(sourceEmoji),
	}
	if sourceChannelID != nil {
		optional = append(optional, db.Starboards.SourceChannelID.Set(*sourceChannelID))
	}

	result, err := s.database.Starboards.CreateOne(
		db.Starboards.GuildID.Set(guildID),
		db.Starboards.TargetChannelID.Set(targetChannelID),
		optional...,
	).Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("starboard: add starboard: %w", err)
	}
	return result, nil
}

func (s *StarboardService) RemoveStarboardByID(guildID string, id int) (*db.StarboardsModel, error) {
	ctx := context.Background()
	config, err := s.database.Starboards.FindFirst(
		db.Starboards.GuildID.Equals(guildID),
		db.Starboards.ID.Equals(id),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("starboard: remove by id find: %w", err)
	}
	if config == nil {
		return nil, nil
	}

	_, err = s.database.Starboards.FindUnique(
		db.Starboards.ID.Equals(config.ID),
	).Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("starboard: remove by id delete: %w", err)
	}
	return config, nil
}

func (s *StarboardService) createEmbed(msg *discordgo.Message) *discordgo.MessageEmbed {
	if len(msg.Content) == 0 && len(msg.Attachments) == 0 {
		return nil
	}

	embed := &discordgo.MessageEmbed{
		Color:     static.EmbedColor,
		Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	if msg.Author != nil {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    msg.Author.Username,
			IconURL: msg.Author.AvatarURL("64"),
		}
	}

	if len(msg.Content) > 0 {
		embed.Description = msg.Content
	}

	if len(msg.Attachments) > 0 {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: msg.Attachments[0].URL,
		}
	}

	return embed
}

func (s *StarboardService) createContentString(count int, emoji string, msg *discordgo.Message) string {
	return fmt.Sprintf("**%d %s** at https://discord.com/channels/%s/%s/%s",
		count, emoji, msg.GuildID, msg.ChannelID, msg.ID)
}

func (s *StarboardService) createStarboard(count int, embed *discordgo.MessageEmbed, msg *discordgo.Message, targetChannelID, emoji string) {
	sent, err := s.bot.ChannelMessageSendComplex(targetChannelID, &discordgo.MessageSend{
		Content: s.createContentString(count, emoji, msg),
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	_, err = s.database.Log.CreateOne(
		db.Log.GuildID.Set(msg.GuildID),
		db.Log.ChannelID.Set(targetChannelID),
		db.Log.MessageID.Set(sent.ID),
		db.Log.OriginalMessageID.Set(msg.ID),
	).Exec(context.Background())
	if err != nil {
		utils.Logger.Errorw("starboard: create log entry failed", "error", fmt.Errorf("starboard: create log: %w", err))
	}

	s.bot.MessageReactionAdd(targetChannelID, sent.ID, emoji)
	s.bot.MessageReactionAdd(msg.ChannelID, msg.ID, "🌟")
}

func (s *StarboardService) updateStarboard(count int, embed *discordgo.MessageEmbed, msg *discordgo.Message, emoji string, log *db.LogModel) {
	_, err := s.bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: log.ChannelID,
		ID:      log.MessageID,
		Content: func() *string { c := s.createContentString(count, emoji, msg); return &c }(),
		Embeds:  &[]*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (s *StarboardService) deleteStarboard(log *db.LogModel) {
	err := s.bot.ChannelMessageDelete(log.ChannelID, log.MessageID)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	_, err = s.database.Log.FindUnique(
		db.Log.MessageID.Equals(log.MessageID),
	).Delete().Exec(context.Background())
	if err != nil {
		utils.Logger.Errorw("starboard: delete log entry failed", "error", fmt.Errorf("starboard: delete log: %w", err))
	}
}
