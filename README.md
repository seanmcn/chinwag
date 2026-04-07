# Chinwag

A native desktop app (macOS / Windows) that turns a WhatsApp chat export into
a dashboard about your relationship with one other person — message volume,
response rhythm, top emojis, conversation flow, sentiment, emotion and a few
cheeky observations. You can export the results as high-quality JPEGs to
share. A small CLI is also included for terminal use.

**Everything runs locally. Your chats never leave your machine.**

## Install

Grab the latest build from the [Releases page](https://github.com/seanmcn/chinwag/releases):

- **macOS** — `chinwag-macos.zip` (universal, signed & notarised — opens like any other Mac app)
- **Windows** — `chinwag-windows.zip` (unsigned; on first launch SmartScreen says *"Windows protected your PC"* → **More info → Run anyway**)
- **CLI** — `go install github.com/seanmcn/chinwag/cmd/chinwag@latest`, or use the Docker image (see below)

## Use it

1. In WhatsApp open a chat → ⋯ menu → **Export chat** → *Without media*. You get a `.txt` (or a `.zip` containing one). Both work.
2. Open Chinwag, click **Open chat export…**, pick the file, choose which person is *you*, hit **Analyse**.
3. Browse the **Overview**, **Conversation**, **Tone** and **Activity** tabs.
4. Use the **Export** tab to save sections as JPEGs — rename participants, pick what to include, choose a single tall image or one file per section.

### CLI

```sh
chinwag --me Alice --them "Bob Smith" data/chat.zip
chinwag --format json data/chat.zip | jq .Messages
```

| Flag | Default | Notes |
|---|---|---|
| `--me` | top author | Your name as it appears in the chat |
| `--them` | 2nd author | The other person |
| `--gap` | `6h` | Silence threshold for splitting conversations |
| `--format` | `text` | `text` or `json` |

Or via Docker:

```sh
docker run --rm -v "$PWD/data:/data" ghcr.io/seanmcn/chinwag /data/chat.zip
```

## Contributing

Bug reports, ideas and PRs welcome — see [CONTRIBUTING.md](CONTRIBUTING.md)
for setup, project layout and dev workflow.

## License

MIT.
