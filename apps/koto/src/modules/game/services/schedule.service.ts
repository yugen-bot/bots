import { Injectable, Logger } from '@nestjs/common';
import { Cron } from '@nestjs/schedule';
import { GameStatus } from '@prisma/koto';
import { addMinutes, addSeconds, isAfter } from 'date-fns';
import { Client } from 'discord.js';

import { SettingsService } from '../../settings';

import { delay } from '@yugen/util';

import { PrismaService } from '@yugen/prisma/koto';

import { GameService } from './game.service';

@Injectable()
export class GameScheduleService {
	private readonly _logger = new Logger(GameScheduleService.name);

	constructor(
		private _prisma: PrismaService,
		private _client: Client,
		private _game: GameService,
		private _settings: SettingsService
	) {}

	@Cron(`0 * * * * *`)
	async check() {
		const stats = {
			outOfTimeGames: 0,
			checkedGuilds: 0,
			startedGames: 0,
		};

		const outOfTimeGames = await this._prisma.game.findMany({
			where: {
				status: GameStatus.IN_PROGRESS,
				endingAt: {
					lte: new Date(),
				},
			},
			select: {
				id: true,
				guildId: true,
			},
		});


		stats.outOfTimeGames = outOfTimeGames.length;
		const endPromises = [];
		for (const game of outOfTimeGames) {
			const guild = await this._client.guilds.fetch(game.guildId).catch(() => null);
			if(!guild) {
				continue;
			}

			endPromises.push(this._endGame(game.id, game.guildId));
		}
		const endGames = await Promise.allSettled(endPromises);
		const startedAfterEndgames = endGames.filter(
			r => r.status === 'fulfilled' && !!r.value
		);

		const guildsWithChannelId = await this._prisma.settings.findMany({
			where: {
				channelId: { not: null },
			},
			select: {
				guildId: true,
			},
		});

		stats.checkedGuilds = guildsWithChannelId.length;
		const promises = [];
		for (const { guildId } of guildsWithChannelId) {
			const guild = await this._client.guilds.fetch(guildId).catch(() => null);
			if (guild) {
				promises.push(this._checkGuild(guildId));
			}
		}

		const guildChecks = await Promise.allSettled(promises);
		const startedGames = guildChecks.filter(
			r => r.status === 'fulfilled' && !!r.value
		);

		stats.startedGames = startedGames.length + startedAfterEndgames.length;
		this._logger.log(
			`Ended ${stats.outOfTimeGames} games. Checked ${stats.checkedGuilds} guilds. Started ${stats.startedGames} games.`
		);
	}

	private async _endGame(id: number, guildId: string) {
		await this._game.endGame(id);

		const settings = await this._settings.getSettings(guildId);
		if (settings.autoStart) {
			await delay(500);
			await this._game.start(guildId, false);
			return true;
		}

		return false;
	}

	private async _checkGuild(guildId: string) {
		const settings = await this._settings.getSettings(guildId);
		const currentGame = await this._prisma.game.findFirst({
			where: {
				guildId,
				status: GameStatus.IN_PROGRESS,
				endingAt: { gt: new Date() },
			},
		});

		if (currentGame) {
			return false;
		}

		const lastGame = await this._prisma.game.findFirst({
			where: {
				guildId,
				status: {
					not: GameStatus.IN_PROGRESS,
				},
			},
			orderBy: {
				createdAt: 'desc',
			},
			include: {
				guesses: settings.startAfterFirstGuess
					? {
							orderBy: {
								createdAt: 'asc',
							},
							take: 1,
					  }
					: false,
			},
		});

		if (
			!lastGame ||
			isAfter(
				addMinutes(
					settings.startAfterFirstGuess && lastGame.guesses?.[0]?.createdAt
						? lastGame.guesses[0].createdAt
						: lastGame.createdAt,
					settings.frequency
				),
				addSeconds(new Date(), 30)
			)
		) {
			return false;
		}

		await this._game.start(guildId, true);
		return true;
	}
}
