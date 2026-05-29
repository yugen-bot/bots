data "external_schema" "ent" {
  program = [
    "go", "run", "-mod=mod",
    "ariga.io/atlas-provider-ent",
    "--path", "./internal/ent/schema",
    "--dialect", "postgres",
  ]
}

env "local" {
  src = data.external_schema.ent.url
  dev = "docker://postgres/15/dev"
  url = getenv("DATABASE_URL")
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "production" {
  url = getenv("DATABASE_URL")
  migration {
    dir = "file:///opt/app/migrations"
  }
}
