package services

import (
	"cmp"
	"slices"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GuildsService struct {
	client *bot.Client
}

func CreateGuildsService(container *di.Container) *GuildsService {
	utils.Logger.Info("Creating Guilds Service")

	return &GuildsService{
		client: container.Get(sharedStatic.DiClient).(*disgoplus.Bot).Client(),
	}
}

func (s *GuildsService) GetData(page int) ([]discord.Guild, int) {
	var guilds []discord.Guild

	for g := range s.client.Caches.Guilds() {
		guilds = append(guilds, g)
	}

	slices.SortFunc(guilds, func(a, b discord.Guild) int {
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
