# WhatsApp Analyse

A native desktop app (macOS / Windows) that turns a WhatsApp chat export into
a dashboard about your relationship with one other person — message volume,
response rhythm, top emojis, conversation flow, sentiment, and a few cheeky
observations. A small CLI is also included for terminal/scripted use.

Everything runs locally. Your chats never leave your machine.

## Desktop app

Grab the latest release from the [Releases page](https://github.com/seanmcn/whatsapp-analyse/releases):

- **macOS**: `WhatsApp-Analyse-macOS.zip` (universal binary)
- **Windows**: `WhatsApp-Analyse-Windows.zip`

macOS builds are signed and notarised, so they open like any other Mac app.
Windows builds are unsigned for now — on first launch SmartScreen will say
*"Windows protected your PC"*; click **More info → Run anyway**. After that
Windows remembers and won't ask again.

Once open, click **Open chat export…**, pick your `.txt` or `.zip` export,
choose which person is "you", and hit **Analyse**.

### Get a chat export

In WhatsApp, open a chat → ⋯ menu → **Export chat** → *Without media*. You'll
get a `.txt` file (or a `.zip` containing one). Both work.

## CLI

For scripts, terminals, or piping into `jq`:

```sh
go install github.com/seanmcn/whatsapp-analyse/cmd/whatsapp-analyse@latest
whatsapp-analyse --me Sean --them "Harry Young" data/chat.zip
whatsapp-analyse --format json data/chat.zip | jq .Messages
```

| Flag | Default | Notes |
|---|---|---|
| `--me` | top author | Your name as it appears in the chat |
| `--them` | 2nd author | The other person |
| `--gap` | `6h` | Silence threshold for splitting conversations |
| `--format` | `text` | `text` or `json` |

A Docker image is available for the CLI:

```sh
docker build -t whatsapp-analyse .
docker run --rm -v "$PWD/data:/data" whatsapp-analyse /data/chat.zip
```

## Development

The repo is a Go workspace: the root module holds the parser/analyse engine
and CLI; `app/` is the Wails desktop app.

```sh
# Run the CLI against the bundled sample
go run ./cmd/whatsapp-analyse testdata/sample_chat.txt

# Run the desktop app in dev mode (requires Wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest)
cd app && wails dev

# Build the desktop app
cd app && wails build

# Tests
go test ./...
```

### Project layout

```
cmd/whatsapp-analyse/   CLI entrypoint (text/JSON output)
internal/parser/        WhatsApp export parser (iOS, Android, .zip)
internal/analyse/       Stats, conversations, insights, rating, sentiment
app/                    Wails desktop app (Go backend + React/TS frontend)
app/frontend/src/       UI components, charts, format helpers
.github/workflows/      CI and tagged release pipelines
testdata/               Tiny sample chat for tests
data/                   Drop your real exports here (gitignored)
```

## License

MIT.
