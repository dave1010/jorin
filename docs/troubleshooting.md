# Troubleshooting

This guide lists common failure modes and how to resolve them.

## API errors

### `API 401` or `API 403`

Likely causes:

- Missing or invalid `OPENAI_API_KEY`.
- An API key that is not authorized for the requested model.

Fix:

- Export a valid API key and retry.
- Verify your account has access to the selected model.

### `API 429`

Likely causes:

- Rate limits or quota exhaustion from the API provider.

Fix:

- Wait and retry, or switch to a different model/account with more quota.

### `API 5xx`

Likely causes:

- Temporary service outage.

Fix:

- Retry after a short delay, or use a different `OPENAI_BASE_URL`.

## Shell/tool issues

### Command blocked by policy

If a tool call returns `{"error":"denied by policy"}` or
`{"error":"not allowed by policy"}`, the `--allow` and `--deny` flags are
intervening.

Fix:

- Remove or loosen `--deny` substrings.
- Add an `--allow` substring that matches the full command.

### No file output

If a `write_file` tool call returns `{"error":"readonly session"}`, the CLI
was started with `--readonly`.

Fix:

- Remove the `--readonly` flag.

## Color or formatting issues

If you see garbled ANSI output or want plain text:

- Set `NO_COLOR=1`, or
- Use a terminal that does not report `TERM=dumb`.

