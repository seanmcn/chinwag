# build_lexicons

Dev-time tool that converts upstream NLP data files into the compact JSON
files embedded by `internal/lexicon`. The Chinwag binary itself never runs
this and never touches the network — this is a one-shot step you run when
you want to rebuild the embedded lexicons from canonical sources.

## Inputs

Download these manually into a directory of your choice and respect each
project's licence:

| File                                       | Source                                                                |
|--------------------------------------------|-----------------------------------------------------------------------|
| `vader_lexicon.txt`                        | https://github.com/cjhutto/vaderSentiment (`vaderSentiment/vader_lexicon.txt`) |
| `stopwords-english.txt`                    | NLTK `corpora/stopwords/english` (one word per line)                  |
| `NRC-Emotion-Lexicon-Wordlevel-v0.92.txt`  | https://saifmohammad.com/WebPages/NRC-Emotion-Lexicon.htm             |
| `NRC-AffectIntensity-Lexicon.txt`          | https://saifmohammad.com/WebPages/AffectIntensity.htm                 |
| `NRC-VAD-Lexicon.txt`                      | https://saifmohammad.com/WebPages/nrc-vad.html                        |

Any subset is fine — missing files just leave the corresponding output
JSON untouched.

## Run

```sh
go run ./scripts/build_lexicons \
    --src ~/Downloads/lexicons \
    --out internal/lexicon/data
```

Then rebuild Chinwag normally — `go build ./...` picks up the new JSON
through the existing `//go:embed` directive in `internal/lexicon/embed.go`.

## Notes

- VADER's booster and negation word lists live as Python constants in
  upstream `vader_utils.py`, not as data files. They're baked into this
  converter (`vaderBoosters`, `vaderNegations`) so the output JSON always
  contains them.
- Words that we treat as negations (`not`, `no`, `never`, …) are *excluded*
  from the VADER lexicon even if upstream lists them — otherwise the
  negation lookback in `internal/analyse/vader.go` short-circuits.
- The NRC EmoLex categories are packed into a 10-bit `uint16` mask in
  `nrc_emotion.json`. Bit order: anger, anticipation, disgust, fear, joy,
  sadness, surprise, trust, negative, positive.
- Chat-specific stopword extras (`lol`, `omg`, `tbh`, …) are appended on
  top of NLTK's english list — see `chatExtras` in `main.go`.
