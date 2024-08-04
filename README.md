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

## Usage

### Basic Usage

After installation, you can use Commitron directly:

```bash
commitron comment -ak YOUR_ACCESS_KEY -sk YOUR_SECRET_KEY -diff "$(git diff --cached)"
```

### Installing Git Alias

For seamless integration with your Git workflow, install the Commitron alias:

```bash
commitron install_alias
```

This will add a `cz` alias to your Git configuration. Now you can use Commitron by simply typing:

```bash
git cz
```

### Dry Run / Preview Mode

You can generate a commit message without actually committing changes. 
This is useful for previewing the message or using Commitron with different inputs:

1. Generate message from staged changes:
```bash
   commitron comment --diff "$(git diff --cached)"
   #or
   commitron comment -ak YOUR_ACCESS_KEY -sk YOUR_SECRET_KEY -diff "$(git diff --cached)"
```

2. Generate message from unstaged changes:
```bash
   commitron comment -ak YOUR_ACCESS_KEY -sk YOUR_SECRET_KEY -diff "$(git diff)"
```

3. Generate message for a specific file:
```bash
   commitron comment -ak YOUR_ACCESS_KEY -sk YOUR_SECRET_KEY -diff "$(git diff HEAD -- path/to/your/file)"
```

4. Generate message based on git blame:
```bash
   commitron comment -ak YOUR_ACCESS_KEY -sk YOUR_SECRET_KEY -diff "$(git blame path/to/your/file)"
```

### Configuration

Commitron can be configured using command-line flags or environment variables:

- Access Key: `-ak` flag or `VOLC_ACCESSKEY` environment variable
- Secret Key: `-sk` flag or `VOLC_SECRETKEY` environment variable
- API Endpoint: `-endpoint` flag or `DOUBAO_ENDPOINT` environment variable

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