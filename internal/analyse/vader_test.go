package analyse

import (
	"testing"

	"github.com/seanmcn/chinwag/internal/lexicon"
)

func TestScoreVader(t *testing.T) {
	lex, err := lexicon.Load()
	if err != nil {
		t.Fatalf("lexicon load: %v", err)
	}

	cases := []struct {
		text    string
		wantPos bool // true => expect compound > 0.05
		wantNeg bool // true => expect compound < -0.05
	}{
		{"The food is good", true, false},
		{"The food is not good", false, true},
		{"The food is VERY GOOD!!!", true, false},
		{"I love this so much", true, false},
		{"I hate this", false, true},
		{"I don't hate this", true, false}, // negation flips negative -> positive-ish
		{"meh", false, false},               // neutral
		{"", false, false},
	}

	for _, c := range cases {
		got := ScoreVader(c.text, lex)
		t.Logf("%-30q -> %+v", c.text, got)
		if c.wantPos && got.Compound <= 0.05 {
			t.Errorf("%q: expected positive, got compound=%.3f", c.text, got.Compound)
		}
		if c.wantNeg && got.Compound >= -0.05 {
			t.Errorf("%q: expected negative, got compound=%.3f", c.text, got.Compound)
		}
		if !c.wantPos && !c.wantNeg && (got.Compound > 0.5 || got.Compound < -0.5) {
			t.Errorf("%q: expected near-neutral, got compound=%.3f", c.text, got.Compound)
		}
	}
}

func TestScoreVaderEmphasis(t *testing.T) {
	lex, _ := lexicon.Load()
	plain := ScoreVader("the food is good", lex)
	emph := ScoreVader("the food is GOOD", lex)
	excl := ScoreVader("the food is good!!!", lex)
	if !(emph.Compound > plain.Compound) {
		t.Errorf("ALL CAPS should boost: plain=%.3f emph=%.3f", plain.Compound, emph.Compound)
	}
	if !(excl.Compound > plain.Compound) {
		t.Errorf("exclamation should boost: plain=%.3f excl=%.3f", plain.Compound, excl.Compound)
	}
}

func TestScoreMessageEmotion(t *testing.T) {
	lex, _ := lexicon.Load()
	ms := ScoreMessage("I am so happy and proud today", lex)
	// joy is bit index 4, positive is 9.
	if ms.Emotion[4] == 0 {
		t.Errorf("expected joy hits, got 0: %+v", ms.Emotion)
	}
	if ms.Emotion[9] == 0 {
		t.Errorf("expected positive hits, got 0: %+v", ms.Emotion)
	}
	if ms.Vader.Compound <= 0 {
		t.Errorf("expected positive compound, got %.3f", ms.Vader.Compound)
	}
}
