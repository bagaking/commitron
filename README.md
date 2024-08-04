# Commitron

Commitron is an AI-powered command-line tool that automatically generates
meaningful Git commit messages based on your code changes. It analyzes your
diff information and uses advanced language models to create concise,
informative commit comments.

## Features

- 🤖 AI-powered commit message generation
- 🔗 Easy integration with Git workflow through custom alias
- 🎨 Customizable prompts for tailored commit styles
- 🔐 Secure API key management
- 🔧 Flexible configuration options
- 🔍 Dry run mode for previewing commit messages

## Installation

To install Commitron, run:

```bash
go install github.com/bagaking/commitron@latest
```

## Local validation

Run the project test suite before opening a pull request:

```bash
make test
```

The `make test` target runs:

```bash
go test ./...
```

For release or maintainer checks, run the broader local gate:

```bash
make check
```

The `make check` target runs the tests and verifies that all packages build.

## Usage

### Basic Usage

After installation, you can use Commitron directly:

```bash
commitron comment \
  --access_key YOUR_ACCESS_KEY \
  --secret_key YOUR_SECRET_KEY \
  --endpoint YOUR_MODEL_ENDPOINT \
  --diff "$(git diff --cached)"
```

`YOUR_ACCESS_KEY`, `YOUR_SECRET_KEY`, and `YOUR_MODEL_ENDPOINT` are example
placeholders. Replace them with your own credentials and endpoint, or omit the
flags when the matching environment variables are already configured.

### Installing Git Alias

For seamless integration with your Git workflow, install the Commitron alias:

```bash
commitron install_alias
```

This writes a `cz` alias to your global Git configuration. If you enter
credentials or an endpoint during `install_alias`, the resulting global alias
may include `-ak`, `-sk`, or `-endpoint` values in plain text. Prefer configuring
credentials with shell environment variables before installing the alias, and
decline the prompts for inline credentials unless you intentionally want those
values embedded in your global Git config.

Now you can use Commitron by simply typing:

```bash
git cz
```

### Dry Run / Preview Mode

You can generate a commit message without actually committing changes.
This is useful for previewing the message or using Commitron with different
inputs:

1. Generate message from staged changes:

```bash
commitron comment --diff "$(git diff --cached)"
# Or with explicit credentials:
commitron comment \
  --access_key YOUR_ACCESS_KEY \
  --secret_key YOUR_SECRET_KEY \
  --endpoint YOUR_MODEL_ENDPOINT \
  --diff "$(git diff --cached)"
```

2. Generate message from unstaged changes:

```bash
commitron comment \
  --access_key YOUR_ACCESS_KEY \
  --secret_key YOUR_SECRET_KEY \
  --endpoint YOUR_MODEL_ENDPOINT \
  --diff "$(git diff)"
```

3. Generate message for a specific file:

```bash
commitron comment \
  --access_key YOUR_ACCESS_KEY \
  --secret_key YOUR_SECRET_KEY \
  --endpoint YOUR_MODEL_ENDPOINT \
  --diff "$(git diff HEAD -- path/to/your/file)"
```

4. Generate message based on git blame:

```bash
commitron comment \
  --access_key YOUR_ACCESS_KEY \
  --secret_key YOUR_SECRET_KEY \
  --endpoint YOUR_MODEL_ENDPOINT \
  --diff "$(git blame path/to/your/file)"
```

### Configuration

Commitron can be configured using command-line flags or environment variables:

- Access Key: `-ak` flag or `VOLC_ACCESSKEY` environment variable
- Secret Key: `-sk` flag or `VOLC_SECRETKEY` environment variable
- API Endpoint: `-endpoint` flag or `DOUBAO_ENDPOINT` environment variable

Command-line flags take precedence over environment variables. The endpoint is
the model service or bot endpoint used to generate the commit message. Commitron
does not require a hard-coded vendor endpoint in the repository; point the
endpoint value at the service you operate or are authorized to use.

For local use, prefer exporting secrets in your shell session:

```bash
export VOLC_ACCESSKEY="YOUR_ACCESS_KEY"
export VOLC_SECRETKEY="YOUR_SECRET_KEY"
export DOUBAO_ENDPOINT="YOUR_MODEL_ENDPOINT"
```

If you keep credentials in a local config file, use one that is ignored by your
repository, loaded by your shell or secret manager, and never committed.

Model selection is currently owned by the configured endpoint or upstream bot
service. If your provider exposes model names separately, keep that selection in
the provider-side configuration or a local wrapper rather than committing
environment-specific model IDs to this repository.

### Security and secrets

Never commit real credentials or private endpoint values. Keep these values in
your shell environment, secret manager, ignored local configuration, or CI
secret store:

- `VOLC_ACCESSKEY`
- `VOLC_SECRETKEY`
- `DOUBAO_ENDPOINT` when it identifies a private or internal service
- provider-specific model names, deployment IDs, tenant IDs, or base URLs that
  should not be public

Use placeholders such as `YOUR_ACCESS_KEY`, `YOUR_SECRET_KEY`, and
`YOUR_MODEL_ENDPOINT` in documentation, tests, issues, and pull requests. Before
publishing a release, inspect staged changes for accidental secrets, private
URLs, local paths, and machine-specific configuration.

### Release readiness

Before tagging or publishing a release:

```bash
make check
git diff --check
```

Confirm the README examples still match `commitron comment --help`, the install
command uses the intended module path, and no credential, endpoint, or local
machine value is staged for commit.

### Custom Prompts

You can customize the AI prompt used for generating commit messages:

```bash
commitron comment -prompt "Your custom prompt here" ...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the [MIT License](LICENSE).

## Acknowledgements

- Thanks to all contributors who have helped shape Commitron.
- Special thanks to the AI models and APIs that power our commit message generation.
