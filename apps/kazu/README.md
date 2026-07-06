<p align="center">
  <a href="https://discord.gg/UttZbEd9zn" target="blank"><img src="https://raw.githubusercontent.com/jurienhamaker/Yugen/main/assets/kazu%20sticker.png" width="200" alt="Kazu logo" /></a>
</p>

  <p align="center">A word-chain on <a href="http://discord.com" target="_blank">Discord</a> bot.</p>
    <p align="center">
      <img src="https://img.shields.io/github/license/yugen-bot/bots
" alt="Package License" />
      <img src="https://img.shields.io/github/actions/workflow/status/jurienhamaker/yugen/yugen.yml" alt="CircleCI" />
      <a href="https://discord.gg/UttZbEd9zn" target="_blank"><img src="https://img.shields.io/badge/discord-online-brightgreen.svg" alt="Discord"/></a>
    </p>
  <!--[![Backers on Open Collective](https://opencollective.com/nest/backers/badge.svg)](https://opencollective.com/nest#backer)
  [![Sponsors on Open Collective](https://opencollective.com/nest/sponsors/badge.svg)](https://opencollective.com/nest#sponsor)-->

## Running Kazu

### Getting started

```bash
git clone git@github.com:jurienhamaker/yugen.git
```

**Copy the `.env.example` to `.env` and change the values in the `.env` file**
**Copy the `apps/kazu/.env.example` to `apps/kazu/.env` and change the values in the `.env` file**

---

### Docker (Recommended)

#### Prerequisite

- [Docker](https://www.docker.com/)

### Running the app

Migrations will automatically run when the bot starts.

```bash
docker-compose up -d db
docker-compose up kazu
```

### Running migrations separately

```bash
docker-compose exec -it kazu make kazu-migrate
```

---

### Go

#### Prerequisite

- [go 1.25](https://go.dev/doc/install)
- [PostgresDB](https://www.postgresql.org/)
- [Atlas CLI](https://atlasgo.io/docs#installation)

### Building the bot & running the bot

```bash
# watch mode (recommended)
$ make kazu

# production mode
$ make kazu-build
$ ./dist/kazu
```

### Running migrations

Migrations use [Ent](https://entgo.io/) for the ORM and [Atlas](https://atlasgo.io/) for schema migrations.

```bash
# Apply pending migrations
$ make kazu-migrate

# Move to kazu directory
$ cd apps/kazu

# Generate a new migration after schema changes
$ make migrate-diff name=<migration_name>

# Validate migration checksums
$ make migrate-validate

# Regenerate Ent code after schema changes
$ make ent-generate
```

---

## Stay in touch

- Author - [Jurien Hamaker](https://jurien.dev)
- Website - [jurien.dev](https://jurien.dev/)
- Ko-Fi - [ko-fi.com/jurienhamaker](https://ko-fi.com/jurienhamaker)

## License

Kazu is [GPL licensed](../../LICENSE).
