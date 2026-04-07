// build_lexicons fetches upstream NLP data files (VADER, NLTK stopwords,
// NRC EmoLex / Affect Intensity / VAD) and converts them into the compact
// JSON files that internal/lexicon embeds at build time.
//
// This is a dev-time tool. The Chinwag binary never runs it; runtime is
// fully offline against the embedded JSON.
//
// Default behaviour fetches VADER + NLTK stopwords from GitHub. The NRC
// lexicons are gated behind a form on saifmohammad.com, so they have no
// canonical raw URL — pass either a mirror URL or a local file path for
// each via the --nrc-* flags.
//
// Usage:
//
//	# fetch VADER + stopwords only
//	go run ./scripts/build_lexicons
//
//	# fetch everything (NRC from a mirror you trust, or local files)
//	go run ./scripts/build_lexicons \
//	    --nrc-emotion   ~/Downloads/NRC-Emotion-Lexicon-Wordlevel-v0.92.txt \
//	    --nrc-intensity ~/Downloads/NRC-AffectIntensity-Lexicon.txt \
//	    --nrc-vad       ~/Downloads/NRC-VAD-Lexicon.txt
//
//	# skip the network entirely and use local copies of all five
//	go run ./scripts/build_lexicons \
//	    --vader     ~/Downloads/vader_lexicon.txt \
//	    --stopwords ~/Downloads/stopwords-english.txt \
//	    --nrc-emotion   ... --nrc-intensity ... --nrc-vad ...
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	urlVaderLexicon  = "https://raw.githubusercontent.com/cjhutto/vaderSentiment/master/vaderSentiment/vader_lexicon.txt"
	urlNLTKStopZip   = "https://raw.githubusercontent.com/nltk/nltk_data/gh-pages/packages/corpora/stopwords.zip"
	urlNRCEmotion    = "https://saifmohammad.com/WebDocs/Lexicons/NRC-Emotion-Lexicon.zip"
	urlNRCIntensity  = "https://saifmohammad.com/WebDocs/Lexicons/NRC-Emotion-Intensity-Lexicon.zip"
	urlNRCVAD        = "https://saifmohammad.com/WebDocs/Lexicons/NRC-VAD-Lexicon-v2.1.zip"
)

func main() {
	out := flag.String("out", "internal/lexicon/data", "output directory for JSON files")
	vaderSrc := flag.String("vader", urlVaderLexicon, "URL or local path for VADER lexicon")
	stopSrc := flag.String("stopwords", urlNLTKStopZip, "URL or local path for NLTK stopwords (the corpora .zip, or a plain english file)")
	nrcEmoSrc := flag.String("nrc-emotion", urlNRCEmotion, "URL or local path for NRC EmoLex (zip or .txt)")
	nrcIntSrc := flag.String("nrc-intensity", urlNRCIntensity, "URL or local path for NRC Emotion Intensity (zip or .txt)")
	nrcVADSrc := flag.String("nrc-vad", urlNRCVAD, "URL or local path for NRC VAD (zip or .txt)")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}

	steps := []struct {
		name    string
		src     string
		zipHint string // substring used to pick the right file inside a zip
		fn      func(io.Reader, string) error
		dst     string
	}{
		{"vader.json", *vaderSrc, "", buildVader, filepath.Join(*out, "vader.json")},
		{"stopwords.json", *stopSrc, "english", buildStopwords, filepath.Join(*out, "stopwords.json")},
		{"nrc_emotion.json", *nrcEmoSrc, "NRC-Emotion-Lexicon-Wordlevel-v0.92.txt", buildNRCEmotion, filepath.Join(*out, "nrc_emotion.json")},
		{"nrc_intensity.json", *nrcIntSrc, "NRC-Emotion-Intensity-Lexicon-v1.txt", buildNRCIntensity, filepath.Join(*out, "nrc_intensity.json")},
		{"nrc_vad.json", *nrcVADSrc, "NRC-VAD-Lexicon-v2.1.txt", buildNRCVAD, filepath.Join(*out, "nrc_vad.json")},
	}

	for _, s := range steps {
		if s.src == "" {
			fmt.Printf("skip  %s (no source)\n", s.name)
			continue
		}
		r, err := fetch(s.src, s.zipHint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch %s: %v\n", s.name, err)
			os.Exit(1)
		}
		if err := s.fn(r, s.dst); err != nil {
			fmt.Fprintf(os.Stderr, "build %s: %v\n", s.name, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", s.dst)
	}
}

// fetch reads src (URL or local path) into memory. If the result is a
// ZIP, it extracts the first member whose name contains zipHint (and ends
// in .txt for safety) and returns a reader over that member's bytes.
// Plain inputs are returned as-is.
func fetch(src, zipHint string) (io.Reader, error) {
	var body []byte
	var err error
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		fmt.Printf("fetch %s\n", src)
		client := &http.Client{Timeout: 120 * time.Second}
		req, _ := http.NewRequest("GET", src, nil)
		req.Header.Set("User-Agent", "chinwag-build_lexicons/1.0")
		resp, err2 := client.Do(req)
		if err2 != nil {
			return nil, err2
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("GET %s: %s", src, resp.Status)
		}
		body, err = io.ReadAll(resp.Body)
	} else {
		body, err = os.ReadFile(src)
	}
	if err != nil {
		return nil, err
	}

	// Detect ZIP via PK\x03\x04 magic.
	if len(body) >= 4 && body[0] == 'P' && body[1] == 'K' {
		zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			return nil, fmt.Errorf("unzip: %w", err)
		}
		// We match by exact basename so directory-prefix substrings
		// (which would match many sibling files in NRC archives) don't
		// fool us into picking a multilingual or per-emotion variant.
		var picked *zip.File
		for _, f := range zr.File {
			if f.FileInfo().IsDir() {
				continue
			}
			if zipHint == "" {
				picked = f
				break
			}
			if filepath.Base(f.Name) == zipHint {
				picked = f
				break
			}
		}
		if picked == nil {
			var names []string
			for _, f := range zr.File {
				names = append(names, f.Name)
			}
			return nil, fmt.Errorf("no zip member matched hint %q (have: %s)", zipHint, strings.Join(names, ", "))
		}
		fmt.Printf("  unzip %s\n", picked.Name)
		rc, err := picked.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		inner, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(inner), nil
	}

	return bytes.NewReader(body), nil
}

// ---- VADER -----------------------------------------------------------------

// vaderBoosters / vaderNegations are baked in from the canonical VADER
// implementation (cjhutto/vaderSentiment, vader_utils.py). Upstream ships
// them as Python constants, not data files, so we mirror them here.
var vaderBoosters = map[string]float64{
	"absolutely": 0.293, "amazingly": 0.293, "completely": 0.293, "considerably": 0.293,
	"decidedly": 0.293, "deeply": 0.293, "effing": 0.293, "enormously": 0.293,
	"entirely": 0.293, "especially": 0.293, "exceptionally": 0.293, "extremely": 0.293,
	"fabulously": 0.293, "flipping": 0.293, "fricking": 0.293, "frigging": 0.293,
	"fucking": 0.293, "fully": 0.293, "greatly": 0.293, "hella": 0.293,
	"highly": 0.293, "hugely": 0.293, "incredibly": 0.293, "intensely": 0.293,
	"majorly": 0.293, "more": 0.293, "most": 0.293, "particularly": 0.293,
	"purely": 0.293, "quite": 0.293, "really": 0.293, "remarkably": 0.293,
	"so": 0.293, "substantially": 0.293, "thoroughly": 0.293, "totally": 0.293,
	"tremendously": 0.293, "uber": 0.293, "unbelievably": 0.293, "unusually": 0.293,
	"utterly": 0.293, "very": 0.293, "super": 0.293, "such": 0.293,

	"almost": -0.293, "barely": -0.293, "hardly": -0.293, "kinda": -0.293,
	"kindof": -0.293, "less": -0.293, "little": -0.293, "marginally": -0.293,
	"occasionally": -0.293, "partly": -0.293, "scarcely": -0.293, "slightly": -0.293,
	"somewhat": -0.293, "sorta": -0.293, "sortof": -0.293,
}

var vaderNegations = []string{
	"not", "no", "never", "none", "nobody", "nothing", "nowhere", "neither",
	"nor", "cannot", "cant", "couldnt", "didnt", "doesnt", "dont", "hadnt",
	"hasnt", "havent", "isnt", "wasnt", "werent", "wont", "wouldnt", "shouldnt",
	"aint", "without",
}

// vaderNegationOverlap lists words we exclude from the VADER lexicon even
// if upstream lists them, because we treat them as negations instead
// (otherwise the negation lookback short-circuits on "is sentiment word").
var vaderNegationOverlap = map[string]bool{
	"no": true, "not": true, "never": true, "nothing": true, "none": true,
	"nobody": true, "nowhere": true, "neither": true, "nor": true,
}

func buildVader(r io.Reader, out string) error {
	lex := map[string]float64{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		mean, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		if vaderNegationOverlap[strings.ToLower(parts[0])] {
			continue
		}
		lex[parts[0]] = mean
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return writeJSON(out, struct {
		Lexicon   map[string]float64 `json:"lexicon"`
		Boosters  map[string]float64 `json:"boosters"`
		Negations []string           `json:"negations"`
	}{lex, vaderBoosters, vaderNegations})
}

// ---- Stopwords -------------------------------------------------------------

// chatExtras are merged on top of NLTK's english stopwords.
var chatExtras = []string{
	"lol", "lmao", "lmfao", "rofl", "haha", "hehe", "yeah", "yep", "yup", "nope",
	"ok", "okay", "omg", "idk", "tbh", "ngl", "rn", "fr", "smh", "tho", "thx",
	"ty", "pls", "plz", "u", "ur", "yall", "gonna", "wanna", "gotta", "kinda",
	"sorta", "dunno", "im", "ima", "btw", "fyi", "imo", "imho", "afaik", "iirc",
	"wtf", "stfu", "ffs", "nvm", "ikr", "tldr", "ish", "bruh", "fam", "yo",
}

func buildStopwords(r io.Reader, out string) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	set := map[string]struct{}{}
	for _, line := range strings.Split(string(raw), "\n") {
		w := strings.TrimSpace(line)
		if w == "" {
			continue
		}
		set[w] = struct{}{}
	}
	for _, w := range chatExtras {
		set[w] = struct{}{}
	}

	words := make([]string, 0, len(set))
	for w := range set {
		words = append(words, w)
	}
	sort.Strings(words)

	return writeJSON(out, struct {
		Words []string `json:"words"`
	}{words})
}

// ---- NRC Emotion (bitmask) -------------------------------------------------

// emotionOrder defines the bit positions used by the bitmask. Must match
// the order consumed by internal/lexicon at runtime.
var emotionOrder = []string{
	"anger", "anticipation", "disgust", "fear", "joy",
	"sadness", "surprise", "trust", "negative", "positive",
}

func buildNRCEmotion(r io.Reader, out string) error {
	idx := map[string]uint16{}
	for i, c := range emotionOrder {
		idx[c] = 1 << uint(i)
	}

	mask := map[string]uint16{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		if parts[2] != "1" {
			continue
		}
		bit, ok := idx[parts[1]]
		if !ok {
			continue
		}
		mask[parts[0]] |= bit
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return writeJSON(out, struct {
		Categories []string          `json:"categories"`
		Words      map[string]uint16 `json:"words"`
	}{emotionOrder, mask})
}

// ---- NRC Affect Intensity --------------------------------------------------

func buildNRCIntensity(r io.Reader, out string) error {
	words := map[string]map[string]float32{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	first := true
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if first {
			first = false
			low := strings.ToLower(line)
			if strings.HasPrefix(low, "term") || strings.HasPrefix(low, "word") || strings.HasPrefix(low, "english") {
				continue
			}
		}
		// NRC Emotion Intensity Lexicon v1 columns: word \t category \t score.
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		score, err := strconv.ParseFloat(parts[2], 32)
		if err != nil {
			continue
		}
		word, cat := parts[0], parts[1]
		if _, ok := words[word]; !ok {
			words[word] = map[string]float32{}
		}
		words[word][cat] = float32(score)
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return writeJSON(out, struct {
		Words map[string]map[string]float32 `json:"words"`
	}{words})
}

// ---- NRC VAD ---------------------------------------------------------------

func buildNRCVAD(r io.Reader, out string) error {
	words := map[string][3]float32{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	first := true
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if first {
			first = false
			low := strings.ToLower(line)
			if strings.HasPrefix(low, "word") || strings.HasPrefix(low, "term") {
				continue
			}
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 4 {
			continue
		}
		v, err1 := strconv.ParseFloat(parts[1], 32)
		a, err2 := strconv.ParseFloat(parts[2], 32)
		d, err3 := strconv.ParseFloat(parts[3], 32)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		// NRC VAD v2.1 ships values in [-1, +1]; the runtime expects
		// [0, 1] (matching v1 and the seeded data file). Linearly
		// rescale: y = (x + 1) / 2.
		words[parts[0]] = [3]float32{
			float32((v + 1) / 2),
			float32((a + 1) / 2),
			float32((d + 1) / 2),
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return writeJSON(out, struct {
		Words map[string][3]float32 `json:"words"`
	}{words})
}

// ---- helpers ---------------------------------------------------------------

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "build_lexicons:", err)
	os.Exit(1)
}
