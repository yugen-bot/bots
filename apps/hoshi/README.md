<p align="center">
  <a href="https://discord.gg/UttZbEd9zn" target="blank"><img src="https://raw.githubusercontent.com/jurienhamaker/Yugen/main/assets/hoshi%20sticker.png" width="200" alt="Hoshi logo" /></a>
</p>

  <p align="center">A starboard for <a href="http://discord.com" target="_blank">Discord</a> bot.</p>
    <p align="center">
      <img src="https://img.shields.io/github/license/jurrienhamaker/yugen" alt="Package License" />
      <img src="https://img.shields.io/github/actions/workflow/status/jurienhamaker/yugen/hoshi.yml" alt="CircleCI" />
      <a href="https://discord.gg/UttZbEd9zn" target="_blank"><img src="https://img.shields.io/badge/discord-online-brightgreen.svg" alt="Discord"/></a>
    </p>
  <!--[![Backers on Open Collective](https://opencollective.com/nest/backers/badge.svg)](https://opencollective.com/nest#backer)
  [![Sponsors on Open Collective](https://opencollective.com/nest/sponsors/badge.svg)](https://opencollective.com/nest#sponsor)-->

## Running Hoshi

### Getting started

```bash
git clone git@github.com:jurienhamaker/yugen.git
```

**Copy the `.env.example` to `.env` and change the values in the `.env` file**
**Copy the `apps/hoshi/.env.example` to `apps/hoshi/.env` and change the values in the `.env` file**

---

### Docker (Recommended)

#### Prerequisite

- [Docker](https://www.docker.com/)

### Running the app

```bash
docker-compose up -d db
docker-compose up hoshi
```

### Running migrations

```bash
docker-compose exec -it koto make hoshi-migrate
```

---

### NodeJS

#### Prerequisite

- [go 1.25](https://go.dev/doc/install)
- [PostgresDB](https://www.postgresql.org/)

### Building the bot & running the bot

```bash
# watch mode (recommended)
$ make hoshi

# production mode
$ make hoshi-build
$ ./dist/hoshi
```

### Running migrations (Development)

```bash
# development
$ make hoshi-migrate
```

---

## Stay in touch

- Author - [Jurien Hamaker](https://jurien.dev)
- Website - [jurien.dev](https://jurien.dev/)
- Ko-Fi - [ko-fi.com/jurienhamaker](https://ko-fi.com/jurienhamaker)

## License

Hoshi is [GPL licensed](../../LICENSE).
