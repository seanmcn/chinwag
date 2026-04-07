package lexicon

import "testing"

func TestLoadEmbeddedLexicons(t *testing.T) {
	lx, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(lx.Vader) == 0 {
		t.Error("vader lexicon empty")
	}
	if len(lx.Boosters) == 0 {
		t.Error("boosters empty")
	}
	if len(lx.Negations) == 0 {
		t.Error("negations empty")
	}
	if len(lx.Stopwords) == 0 {
		t.Error("stopwords empty")
	}
	if len(lx.EmotionCategories) != 10 {
		t.Errorf("want 10 NRC categories, got %d", len(lx.EmotionCategories))
	}
	if len(lx.EmotionMask) == 0 {
		t.Error("emotion mask empty")
	}
	if len(lx.Intensity) == 0 {
		t.Error("intensity empty")
	}
	if len(lx.VAD) == 0 {
		t.Error("vad empty")
	}
}
