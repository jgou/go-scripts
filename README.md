# go-scripts

A collection of small, standalone Go command-line tools for day-to-day operations tasks. Each tool lives in its own folder with its own `go.mod`, so it can be built and run independently.

## Scripts

| Script | Description |
|--------|-------------|
| [gcp-secret-copy](gcp-secret-copy/) | Copies the latest version of selected GCP Secret Manager secrets from a source project to a target project, per customer and environment. Dry run by default. |

Each folder has its own README with prerequisites, configuration and usage details.

## Getting started

```bash
git clone <repo-url>
cd go-scripts/<script-folder>
go run . -h
```

## Conventions

- **One folder per tool**, each a separate Go module.
- **Safe by default:** tools that change anything should be a dry run unless an explicit flag such as `-apply` is passed.
- **Local config stays out of the repo:** real configuration files (project IDs, secret names, etc.) are git-ignored. Each tool ships a `*.example` file as a reference to copy and edit.

## Adding a new script

1. Create a new folder at the repo root and run `go mod init` inside it.
2. Add a `README.md` describing what it does and how to use it.
3. If it needs a local config file, commit a `*.example` version and add the real one to [.gitignore](.gitignore).
4. Add a row to the table above.
