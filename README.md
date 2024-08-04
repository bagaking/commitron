# Commitron

Commitron is a small AI commit CLI. It takes a Git diff, sends that diff to the
bot/model endpoint you configure, and prints a proposed commit message.

Commitron does not replace review, staging, or the final commit decision. The
core command, `commitron comment`, only generates text. The optional Git alias,
`git cz`, wraps that text in `git commit -e -m` so you can inspect and edit the
message before the commit is created.

## What Commitron Owns

- Generate a commit message from an explicit `--diff` input.
- Use either command-line flags or environment variables for credentials and the
  generation endpoint.
- Let callers override the default prompt with `--prompt`.
- Install a convenience `cz` alias into global Git config when requested. The
  alias reads credentials and endpoint from environment variables at runtime.

## What You Still Own

- Stage the exact changes you want committed.
- Review the generated message before accepting it.
- Keep API keys and private endpoint values out of Git history, shell history,
  issue comments, and screenshots.
- Point `DOUBAO_ENDPOINT` or `--endpoint` at a service you operate or are
  authorized to use. The README intentionally does not name one internal
  endpoint as the only valid route.

## Installation

```bash
go install github.com/bagaking/commitron@latest
```

The module path is `github.com/bagaking/commitron`.

## Configuration

`commitron comment` requires a diff, an access key, a secret key, and an
endpoint. Flags take precedence over environment variables.

| Setting | Flag | Environment variable |
| --- | --- | --- |
| Access key | `--access_key` or `--ak` | `VOLC_ACCESSKEY` |
| Secret key | `--secret_key` or `--sk` | `VOLC_SECRETKEY` |
| Generation endpoint | `--endpoint` or `-e` | `DOUBAO_ENDPOINT` |
| Prompt override | `--prompt` or `-p` | none |

For day-to-day local use, prefer environment variables or a secret manager over
inline flags:

```bash
export VOLC_ACCESSKEY="YOUR_ACCESS_KEY"
export VOLC_SECRETKEY="YOUR_SECRET_KEY"
export DOUBAO_ENDPOINT="YOUR_MODEL_ENDPOINT"
```

`YOUR_MODEL_ENDPOINT` is a placeholder. Use the bot or model service endpoint
that is valid for your environment. If your provider handles model selection
outside the endpoint value, keep provider-specific model IDs, deployment IDs,
tenant IDs, and base URLs in local or provider-side configuration rather than in
this repository.

## Commands

Generate a message from staged changes:

```bash
commitron comment --diff "$(git diff --cached)"
```

Generate a message with explicit credentials and endpoint:

```bash
commitron comment \
  --access_key YOUR_ACCESS_KEY \
  --secret_key YOUR_SECRET_KEY \
  --endpoint YOUR_MODEL_ENDPOINT \
  --diff "$(git diff --cached)"
```

Generate from unstaged changes:

```bash
commitron comment --diff "$(git diff)"
```

Generate for one path:

```bash
commitron comment --diff "$(git diff HEAD -- path/to/file)"
```

Use a custom prompt:

```bash
commitron comment \
  --prompt "Write a concise Conventional Commit message." \
  --diff "$(git diff --cached)"
```

Check the current command surface:

```bash
go run . comment --help
```

Show local Git activity insight for one committer:

```bash
commitron insight --committer "Author Name"
```

`insight` reads local Git history and reports commit and line-change summaries
for the requested committer. The legacy misspelled `--commiter` flag remains
available as a compatibility alias. It does not call the model endpoint.

## Git Alias

Install the convenience alias:

```bash
commitron install_alias
```

This appends a `cz` alias to your global Git config. After installation:

```bash
git cz
```

The alias reads `git diff --cached`, calls `commitron comment`, and then runs
`git commit -e -m "$COMMIT_MSG_CONTENT"`. Git opens the message for editing
before creating the commit.

Credential handling: `install_alias` does not prompt for an access key, secret
key, or endpoint, and the generated alias does not embed `-ak`, `-sk`, or
`-endpoint` arguments. Set `VOLC_ACCESSKEY`, `VOLC_SECRETKEY`, and
`DOUBAO_ENDPOINT` in your shell or secret manager before running `git cz`.

Config permissions: installation rewrites the global Git config with `0600`
permissions so alias content is not world-readable.

Alias scope risk: the alias name is `cz`, and installation refuses to continue
if a global `cz` alias already exists. Review your global Git config before
installing if you already use that alias name.

## Credential Handling

Do not commit real credentials, private endpoint values, local config files, or
provider-specific deployment identifiers.

Keep these values in a shell session, ignored local config, a secret manager, or
CI secret storage:

- `VOLC_ACCESSKEY`
- `VOLC_SECRETKEY`
- `DOUBAO_ENDPOINT` when it identifies a private or internal service
- provider-specific model names, deployment IDs, tenant IDs, and base URLs

Use placeholders such as `YOUR_ACCESS_KEY`, `YOUR_SECRET_KEY`, and
`YOUR_MODEL_ENDPOINT` in docs, tests, issues, and pull requests.

## Local Validation

Run the command help check after changing CLI-facing documentation:

```bash
go run . comment --help
```

Run the local gate before opening a pull request:

```bash
make check
```

`make check` runs:

```bash
go test ./...
go build -o .build/ .
```

The build output belongs under `.build/`. Do not leave a root-level `commitron`
binary in the repository.

## Release Gate

Before tagging or publishing:

```bash
go run . comment --help
make check
test ! -e ./commitron
git diff --check
git diff --cached --check
```

Also inspect staged changes for real secrets, private URLs, internal endpoints,
local machine paths, and machine-specific configuration. The gate proves the
local command surface still builds; it does not prove that a generated commit
message is semantically correct for every diff.

## License

This project is licensed under the [MIT License](LICENSE).
