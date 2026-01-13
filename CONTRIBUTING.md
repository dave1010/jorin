# Contributing

Thanks for considering contributing to jorin! This guide covers expectations
for changes and links to the deeper development guide.

## Development workflow

See the [development and architecture guide](docs/development.md) for build,
formatting, linting, and test commands.

Before submitting a change, run:

```bash
make fmt
make test
make lint
```

## Documentation and releases

- Update the README or other documentation if you change behavior or add new
  commands.
- For release-related changes, update CHANGELOG.md (see release workflow
  guidance in README).
- Include or update usage examples when changing CLI flags or behavior.

## Pull requests

1. Fork the repository and create a branch for your change.
2. Make changes, run the checks above, and ensure formatting is correct.
3. Open a pull request describing your changes and linking any related issues.
