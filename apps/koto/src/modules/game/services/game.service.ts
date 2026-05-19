import { Injectable, Logger } from '@nestjs/common';
import { Game, GameStatus, Guess, Settings } from '@prisma/koto';
import {
	addMinutes,
	addSeconds,
	isAfter,
	roundToNearestMinutes,
	setYear,
	subSeconds,
} from 'date-fns';
import { Message, Client } from 'discord.js';

import { SettingsService } from '../../settings';
import { WordsService } from '../../words/services/words.service';
import {
	GAME_TYPE,
	GameGuessMeta,
	GameMeta,
	GameWithMetaAndGuesses,
} from '../types/meta';

import { delay, getTimestamp } from '@yugen/util';

import { PrismaService } from '@yugen/prisma/koto';

import { GameMessageService } from './message.service';
import { GamePointsService } from './points.service';

@Injectable()
export class GameService {
	private readonly _logger = new Logger(GameService.name);

	constructor(
		private _prisma: PrismaService,
		private _settings: SettingsService,
		private _words: WordsService,
		private _message: GameMessageService,
		private _points: GamePointsService,
		private _client: Client
	) {}

	async start(
		guildId: string,
		schedule = false,
		recreate = false,
		word = undefined
	): Promise<boolean | void> {
		this._logger.log(`Trying to start a game for ${guildId}`);

		const guild = await this._client.guilds.fetch(guildId).catch(() => null);
		if(!guild) {
			return;
		}

		const currentGame = await this.getCurrentGame(guildId);

		if (currentGame && !recreate) {
			this._logger.debug(
				`Already a current game with id ${currentGame.id} (no recreate) for ${guildId}`
			);
			return;
		}

		if (currentGame && recreate) {
			await this.endGame(currentGame.id, GameStatus.FAILED);
			await delay(500);
		}

		const pastFiftyGames = await this._prisma.game.findMany({
			where: {
				guildId,
			},
			select: {
				word: true,
				number: true,
			},
			orderBy: {
				createdAt: 'desc',
			},
			take: 50,
		});

		const lastGame = pastFiftyGames.at(0);

		const ignoredWords = pastFiftyGames.map(g => g.word);
		word = word ?? this._words.getRandom(ignoredWords);

		if (!word) {
			return;
		}

		const settings = await this._settings.getSettings(guildId);
		const game = await this._prisma.game.create({
			data: {
				guildId,
				word,
				scheduleStarted: schedule,
				endingAt: settings.startAfterFirstGuess
					? setYear(new Date(), 3000)
					: roundToNearestMinutes(addMinutes(new Date(), settings.timeLimit)),
				meta: this._createBaseState(word) as never,
				number: lastGame ? lastGame.number + 1 : 1,
			},
			include: {
				guesses: true,
			},
		});

		await this._message.create(game as GameWithMetaAndGuesses, true);
		return true;
	}

	async guess(
		guildId: string,
		word: string,
		message: Message,
		settings: Settings
	) {
		let game = await this.getCurrentGame(guildId);

		await this._points.getPlayer(guildId, message.author.id);

		if (!game) {
			return;
		}

		const alreadyGuessedIndex = game.guesses.findIndex(
			(g: Guess) => g.word === word
		);

		if (alreadyGuessedIndex >= 0 && process.env['NODE_ENV'] === 'production') {
			await message.react('❌');
			return;
		}

		const cooldown = await this._checkCooldown(
			message.author.id,
			game,
			settings
		);

		if (cooldown.hit || cooldown.repeatHit) {
			message.react('🕒');
			let suffix = `you can guess again <t:${getTimestamp(cooldown.result)}:R>`;

			if (cooldown.hit && cooldown.repeatHit) {
				suffix = `you can guess again <t:${getTimestamp(
					cooldown.repeatResult
				)}:R> on your own or <t:${getTimestamp(
					cooldown.result
				)}:R> after a guess from another player.`;
			}

			if (!cooldown.hit && cooldown.repeatHit) {
				suffix = `you can guess again <t:${getTimestamp(
					cooldown.repeatResult
				)}:R> or immediatly after a guess from another player.`;
			}

			message.reply(`You're on a cooldown, ${suffix}`);
			return;
		}

		const { meta, guessed, points, gameMeta } = this._checkWord(
			game.word,
			word,
			game.meta as never
		);

		const createdGuess = await this._prisma.guess.create({
			data: {
				game: {
					connect: { id: game.id },
				},
				points,
				userId: message.author.id,
				word,
				meta: meta as never,
			},
		});

		let status = guessed ? GameStatus.COMPLETED : game.status;
		if (game.guesses.length + 1 === 9 && !guessed) {
			status = GameStatus.FAILED;
		}

		game = await this._prisma.game.update({
			where: {
				id: game.id,
			},
			data: {
				status,
				meta: gameMeta as never,
				endingAt:
					settings.startAfterFirstGuess && game.guesses.length === 0
						? roundToNearestMinutes(addMinutes(new Date(), settings.timeLimit))
						: game.endingAt,
			},
			include: {
				guesses: true,
			},
		});

		message
			.react(guessed ? '🎉' : '✅')
			.catch(error => this._logger.error(error));

		const promises = [];
		if (guessed) {
			promises.push(this._points.applyPoints(game, message.author.id));
		}

		if (status !== GameStatus.COMPLETED && settings.informCooldownAfterGuess) {
			const cooldownReply = message.reply(
				`You are now on a cooldown. You can guess again ${
					settings.enableBackToBackCooldown
						? `<t:${getTimestamp(
								addMinutes(createdGuess.createdAt, settings.backToBackCooldown)
						  )}:R> on your own or `
						: ''
				}<t:${getTimestamp(
					addMinutes(createdGuess.createdAt, settings.cooldown)
				)}:R>${
					settings.enableBackToBackCooldown
						? '  after a guess from another player'
						: ''
				}.`
			);
			promises.push(cooldownReply);
		}

		this._message.create(game as GameWithMetaAndGuesses, false);
		await Promise.allSettled(promises);

		if (status !== GameStatus.IN_PROGRESS) {
			// after the message has been send, we can delete the guesses so we don't keep any message content ever.
			await this._prisma.guess.deleteMany({
				where: {
					gameId: game.id,
				},
			});

			if (settings.autoStart) {
				await delay(500);
				return this.start(guildId);
			}
		}
	}

	async endGame(gameId: number, status: GameStatus = GameStatus.OUT_OF_TIME) {
		const game = await this._prisma.game.update({
			where: {
				id: gameId,
			},
			data: {
				status,
			},
			include: {
				guesses: true,
			},
		});

		await this._message.create(game as GameWithMetaAndGuesses, false);

		// after the message has been send, we can delete the guesses so we don't keep any message content ever.
		return this._prisma.guess.deleteMany({
			where: {
				gameId: game.id,
			},
		});
	}

	getCurrentGame(guildId: string): Promise<Game & { guesses: Guess[] }> {
		return this._prisma.game.findFirst({
			where: {
				guildId,
				endingAt: {
					gt: new Date(),
				},
				status: GameStatus.IN_PROGRESS,
			},
			include: {
				guesses: true,
			},
			orderBy: {
				createdAt: 'desc',
			},
		});
	}

	private _checkWord(word: string, guess: string, state: GameMeta) {
		const meta: GameGuessMeta[] = Array.from({ length: guess.length });
		const unmatched = {}; // unmatched word letters
		const letterCount = {};

		// color matched guess letters as correct-spot,
		// and count unmatched word letters
		for (const [index, letter] of [...word].entries()) {
			if (letter === guess[index]) {
				letterCount[letter] = (letterCount[letter] || 0) + 1;

				let points = 0;

				if (state.discovery.correct[letter] < letterCount[letter]) {
					points = 2;
					state.discovery.correct[letter] += 1;

					if (state.discovery.almost[letter] >= letterCount[letter]) {
						points = 1;
					}
				}

				if (state.discovery.almost[letter] < letterCount[letter]) {
					state.discovery.almost[letter] += 1;
				}

				meta[index] = {
					type: GAME_TYPE.CORRECT,
					points,
					letter: letter,
				};

				if (
					!state.keyboard[letter] ||
					state.keyboard[letter] !== GAME_TYPE.CORRECT
				) {
					state.keyboard[letter] = GAME_TYPE.CORRECT;
				}

				state.word[index].type = GAME_TYPE.CORRECT;
				continue;
			}

			unmatched[letter] = (unmatched[letter] || 0) + 1;
		}

		// color unmatched guess letters right-to-left,
		// allocating remaining word letters as wrong-spot,
		// otherwise, as not-any-spot
		for (const [index, element] of [...word].entries()) {
			const letter = guess[index];
			if (letter !== element) {
				if (unmatched[letter]) {
					letterCount[letter] = (letterCount[letter] || 0) + 1;

					let points = 0;

					if (state.discovery.almost[letter] < letterCount[letter]) {
						points = 1;
						state.discovery.almost[letter] += 1;
					}

					meta[index] = {
						type: GAME_TYPE.ALMOST,
						points,
						letter: letter,
					};
					unmatched[letter]--;

					if (
						!state.keyboard[letter] ||
						state.keyboard[letter] !== GAME_TYPE.CORRECT
					) {
						state.keyboard[letter] = GAME_TYPE.ALMOST;
					}

					continue;
				}

				meta[index] = {
					type: GAME_TYPE.WRONG,
					points: 0,
					letter: letter,
				};

				if (!state.keyboard[letter]) {
					state.keyboard[letter] = GAME_TYPE.WRONG;
				}
			}
		}

		return {
			meta,
			gameMeta: state,
			guessed: word === guess,
			points: Object.keys(meta).reduce((pts, key) => pts + meta[key].points, 0),
		};
	}

	private async _checkCooldown(
		userId: string,
		{ guesses }: Game & { guesses: Guess[] },
		{ enableBackToBackCooldown, cooldown, backToBackCooldown }: Settings
	): Promise<{
		hit: boolean;
		repeatHit: boolean;
		result?: Date;
		repeatResult?: Date;
	}> {
		// if (process.env['NODE_ENV'] !== 'production') {
		// 	return { hit: false, repeatHit: false };
		// }
		//
		if (guesses.length === 0) {
			return { hit: false, repeatHit: false };
		}

		const guessesSorted = guesses.sort(
			(a, b) => b.createdAt.getTime() - a.createdAt.getTime()
		);
		const lastGuessByUser = guessesSorted.find(g => g.userId === userId);
		const lastGuess = guessesSorted[0];

		if (!lastGuess || !lastGuessByUser) {
			return { hit: false, repeatHit: false };
		}

		const backToBackCooldownHit =
			enableBackToBackCooldown &&
			isAfter(
				lastGuessByUser.createdAt,
				subSeconds(new Date(), backToBackCooldown)
			) &&
			userId === lastGuess.userId;
		const cooldownHit = isAfter(
			lastGuessByUser.createdAt,
			subSeconds(new Date(), cooldown)
		);

		if (backToBackCooldownHit || cooldownHit) {
			return {
				hit: cooldownHit,
				repeatHit: backToBackCooldownHit,
				result: addSeconds(lastGuessByUser.createdAt, cooldown),
				repeatResult: addSeconds(lastGuessByUser.createdAt, backToBackCooldown),
			};
		}

		return { hit: false, repeatHit: false };
	}

	private _createBaseState(word: string): GameMeta {
		const data = {
			almost: {},
			correct: {},
		};

		for (const letter of word) {
			data.almost[letter] = 0;
			data.correct[letter] = 0;
		}

		return {
			keyboard: {},
			word: [...word].map((letter, index) => ({
				letter,
				index,
				type: GAME_TYPE.WRONG,
			})),
			discovery: data,
		};
	}
}
