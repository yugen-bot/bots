package services

import (
	"cmp"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GuildsService struct {
	bot *discordgoplus.Bot
}

func CreateGuildsService(container *di.Container) *GuildsService {
	utils.Logger.Info("Creating Guilds Service")

	return &GuildsService{
		bot: container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

func (s *GuildsService) GetData(page int) ([]*discordgo.Guild, int) {
	guilds := make([]*discordgo.Guild, len(s.bot.State.Guilds))
	copy(guilds, s.bot.State.Guilds)

	slices.SortFunc(guilds, func(a, b *discordgo.Guild) int {
		return cmp.Compare(b.MemberCount, a.MemberCount)
	})

	total := len(guilds)

	start := (page - 1) * 10
	if start >= total {
		return nil, total
	}

	end := start + 10
	if end > total {
		end = total
	}

	return guilds[start:end], total
}
