# Chinwag

A small Go CLI that turns a WhatsApp chat export into a single-page dashboard
about your relationship with one other person — message volume, response
rhythm, top emojis, conversation flow, and a few cheeky observations.

Everything runs locally. Your chats never leave your machine.

## Install

```sh
go install github.com/seanmcn/whatsapp-analyse/cmd/whatsapp-analyse@latest
```

Or build from source:

```sh
git clone https://github.com/seanmcn/whatsapp-analyse
cd whatsapp-analyse
go build ./cmd/whatsapp-analyse
```

## Get a chat export

In WhatsApp, open a chat → ⋯ menu → **Export chat** → *Without media*. You'll
get a `.txt` file (or a `.zip` containing one). Both work.

## Run

```sh
whatsapp-analyse path/to/_chat.zip
```

Then open <http://127.0.0.1:8080>.

By default the two most-frequent authors are picked automatically. To override:

```sh
whatsapp-analyse --me Sean --them "Harry Young" data/Harry\ Young.zip
```

### Flags

| Flag | Default | Notes |
|---|---|---|
| `--me` | top author | Your name as it appears in the chat |
| `--them` | 2nd author | The other person |
| `--addr` | `127.0.0.1:8080` | Listen address |
| `--gap` | `6h` | Silence threshold for splitting conversations |

## What you get

- **Top bar** — chat points, time period, total messages and conversations.
- **Relationship growth** — monthly message volume per person.
- **Chat rating** — a 0–100 score from balance, response speed, and reciprocity.
- **Key insights** — observations like *"You laugh more than your contact"*.
- **Language analysis** — emojis, laughs, apologies, questions, encouragement,
  and each person's top 5 emojis.
- **Balance** — how evenly the conversation is shared.
- **Message & media analysis** — words, unique words, characters, images,
  videos, audios, GIFs, stickers, links.
- **Responding** — rapid first replies, average first response, average reply
  time.
- **Conversation analysis** — convos started, closed, missed, reconnects,
  double messages.
- **Messaging times** — weekday × hour heatmap of when you actually talk.
- **Conversation flow** — sankey of how chats start, who carries them, and
  who has the last word.
- **Daily chat activity** — GitHub-style grid for the last 500 days.

## Project layout

```
cmd/whatsapp-analyse/   CLI entrypoint
internal/parser/        WhatsApp export parser (iOS, Android, .zip)
internal/analyse/       Stats, conversations, insights, rating
internal/render/        html/template + embedded CSS/JS
internal/server/        Local http server
testdata/               Tiny sample chat for tests
data/                   Drop your real exports here (gitignored)
```

## Development

```sh
go test ./...
go run ./cmd/whatsapp-analyse testdata/sample_chat.txt
```

## License

MIT.
