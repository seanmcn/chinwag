// Package lexicon owns the embedded NLP data files (VADER, NLTK stopwords,
// NRC emotion / intensity / VAD) and provides a single Load() entry point
// that the analyse package consumes.
//
// Everything ships as compact JSON under data/. The shipped lexicons are
// pragmatic seeds — large enough to "feel right" on chat data without
// bloating the binary. To rebuild from upstream sources (full VADER, full
// NRC), run scripts/build_lexicons against a directory of raw files;
// the script overwrites these JSONs.
package lexicon

import (
	"encoding/json"
	"sync"
)

// Lexicons holds all NLP data tables. Maps are read-only after Load.
type Lexicons struct {
	// VADER
	Vader     map[string]float64 // word -> valence in [-4, +4]
	Boosters  map[string]float64 // word -> ±0.293-ish shift
	Negations map[string]struct{}

	// Stopwords (NLTK english + chat extras)
	Stopwords map[string]struct{}

	// NRC categories list, indexed by bit position in EmotionMask values.
	EmotionCategories []string
	// EmotionMask: word -> bitmask over EmotionCategories.
	EmotionMask map[string]uint16
	// Intensity: word -> per-category strength in [0,1]. Sparse.
	Intensity map[string]map[string]float32
	// VAD: word -> [valence, arousal, dominance], each in [0,1].
	VAD map[string][3]float32
}

var (
	loadOnce sync.Once
	loaded   *Lexicons
	loadErr  error
)

// Load returns the singleton lexicon set. The first call decodes the
// embedded JSON; subsequent calls are free.
func Load() (*Lexicons, error) {
	loadOnce.Do(func() {
		loaded, loadErr = decodeAll()
	})
	return loaded, loadErr
}

// MustLoad panics on error. Used by callers that have no sensible fallback.
func MustLoad() *Lexicons {
	lx, err := Load()
	if err != nil {
		panic("lexicon: " + err.Error())
	}
	return lx
}

func decodeAll() (*Lexicons, error) {
	lx := &Lexicons{
		Negations: map[string]struct{}{},
		Stopwords: map[string]struct{}{},
	}

	// VADER
	var vraw struct {
		Lexicon   map[string]float64 `json:"lexicon"`
		Boosters  map[string]float64 `json:"boosters"`
		Negations []string           `json:"negations"`
	}
	if err := readJSON("data/vader.json", &vraw); err != nil {
		return nil, err
	}
	lx.Vader = vraw.Lexicon
	lx.Boosters = vraw.Boosters
	for _, n := range vraw.Negations {
		lx.Negations[n] = struct{}{}
	}

	// Stopwords
	var sraw struct {
		Words []string `json:"words"`
	}
	if err := readJSON("data/stopwords.json", &sraw); err != nil {
		return nil, err
	}
	for _, w := range sraw.Words {
		lx.Stopwords[w] = struct{}{}
	}

	// NRC emotion (bitmask)
	var eraw struct {
		Categories []string          `json:"categories"`
		Words      map[string]uint16 `json:"words"`
	}
	if err := readJSON("data/nrc_emotion.json", &eraw); err != nil {
		return nil, err
	}
	lx.EmotionCategories = eraw.Categories
	lx.EmotionMask = eraw.Words

	// NRC intensity
	var iraw struct {
		Words map[string]map[string]float32 `json:"words"`
	}
	if err := readJSON("data/nrc_intensity.json", &iraw); err != nil {
		return nil, err
	}
	lx.Intensity = iraw.Words

	// NRC VAD
	var vadRaw struct {
		Words map[string][3]float32 `json:"words"`
	}
	if err := readJSON("data/nrc_vad.json", &vadRaw); err != nil {
		return nil, err
	}
	lx.VAD = vadRaw.Words

	return lx, nil
}

func readJSON(name string, v any) error {
	b, err := files.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
