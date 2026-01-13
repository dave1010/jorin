# `jorin`

![jorin screenshot](docs/jorin-screenshot.png)

Jorin is a small coding agent that calls tools (shell, read_file, write_file,
http_get) and communicates with an OpenAI-compatible API. It is designed for
use as a command-line tool and REPL.

## Documentation

- [Usage guide](docs/usage.md)
- [Development and architecture](docs/development.md)
- [Security notes](docs/security.md)
- [Contributing](CONTRIBUTING.md)
- [Code of conduct](CODE_OF_CONDUCT.md)
- [Changelog](CHANGELOG.md)

## Install

One line install:

```bash
curl -fsSL https://get.jorin.ai | bash
```

Download the latest release for your platform from
[GitHub Releases](https://github.com/dave1010/jorin/releases).

Then add it to your $PATH.

## Configuration

Set your API key before running jorin:

```bash
export OPENAI_API_KEY="your-api-key"
```

To use a different OpenAI-compatible endpoint:

```bash
export OPENAI_BASE_URL="https://api.openai.com"
```

## Quick start

Show help:

```bash
jorin --help
```

Start the REPL (default when invoked with no args):

```bash
jorin
```

Send a single prompt:

```bash
jorin "Refactor function X to be smaller"
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) and follow the
[Code of Conduct](CODE_OF_CONDUCT.md).

## License

MIT
