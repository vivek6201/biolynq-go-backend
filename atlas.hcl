locals {
  # Read and parse the .env file if it exists, otherwise fallback to empty string
  env_content = try(file(".env"), "")
  env_lines   = split("\n", local.env_content)
  env_vars = {
    for line in local.env_lines :
    split("=", line)[0] => trim(regex("=(.*)", line)[0], "\"'\r")
    if !startswith(line, "#") && length(split("=", line)) > 1
  }

  # Extract DB_URL with fallback options
  db_url = try(
    local.env_vars["DB_URL"],
    getenv("DB_URL") != "" ? getenv("DB_URL") : "postgres://postgres:postgres@localhost:5432/biolynq_db?sslmode=disable"
  )
}

data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/models",
    "--dialect", "postgres"
  ]
}

env "local" {
  src = data.external_schema.gorm.url
  url = local.db_url
  dev = "docker://postgres/15/dev?search_path=public"
  
  migration {
    dir = "file://migrations"
  }
  
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
