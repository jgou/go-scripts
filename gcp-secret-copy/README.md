# gcp-secret-copy

A small Go CLI that copies the **latest version** of selected secrets from one GCP Secret Manager project to another, per customer and environment.

For each configured secret it:

1. Reads `projects/<source>/secrets/<name>/versions/latest`.
2. Adds that payload as a **new version** of `projects/<target>/secrets/<name>`.

The secret must already exist in the target project (only a new version is added; the secret itself is not created). The secret name is the same in both projects.

By default the tool runs as a **dry run** and only prints what it would copy. Nothing is read or written until you pass `-apply`.

## Prerequisites

- Go (see [go.mod](go.mod) for the version).
- Application Default Credentials, e.g. `gcloud auth application-default login`.
- **The secrets must already exist in every target project.** The tool only adds a new version to an existing secret; it does not create secrets. If a target secret is missing, that secret fails with a `NotFound` error and is skipped. Source secrets must exist too and have at least one enabled version.
- The identity must have:
  - `roles/secretmanager.secretAccessor` on the source secrets.
  - `roles/secretmanager.secretVersionAdder` on the target secrets.

## Setup

The real config file, `settings.json`, is **not part of the repo** (it is git-ignored) so that project IDs and secret names stay private. The repo only ships [settings.json.example](settings.json.example) as a reference. Create your own copy and edit it:

```bash
cp settings.json.example settings.json
```

### Configuration format

The config has three top-level maps, all keyed by **customer** and then **environment**:

```json
{
  "source_projects": {
    "Customer1": { "dev": "gcp-project-dev" }
  },
  "secrets": {
    "Customer1": { "dev": ["secret1", "secret2"] }
  },
  "target_projects": {
    "Customer1": { "dev": ["gcp-project-dev1"] }
  }
}
```

- `source_projects`: the project ID to read secrets from (a single string).
- `secrets`: the list of secret names to copy.
- `target_projects`: a list of project IDs; **only the first entry is currently used**.

Validation on load:

- Unknown JSON fields are rejected.
- Every customer/environment listed under `secrets` must have a non-empty source project and at least one target project, otherwise the tool exits with an error.

## Usage

Run from this folder. By default the config is read from `./settings.json`.

```bash
# Dry run for everything in the config
go run .

# Dry run limited to one customer / environment
go run . -customer Customer1 -env dev

# Actually copy the secrets
go run . -customer Customer1 -env dev -apply

# Use a config file in another location
go run . -config /path/to/other.json -apply
```

### Flags

| Flag        | Default         | Description                                              |
|-------------|-----------------|----------------------------------------------------------|
| `-apply`    | `false`         | Perform the copy. Without it, only a dry run is printed. |
| `-config`   | `settings.json` | Path to the JSON config file.                            |
| `-customer` | _(all)_         | Only process this customer key (e.g. `Customer1`).       |
| `-env`      | _(all)_         | Only process this environment key (e.g. `dev`).          |

## Notes

- A failure on one secret is logged and the tool continues with the next one; the exit code is still `0`, so check the output.
- Secret payloads are never printed, but they are held in memory while being copied.
- Each run adds a new version in the target, even if the value is unchanged.
