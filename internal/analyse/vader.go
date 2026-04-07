package analyse

import (
	"math"
	"strings"
	"unicode"

	"github.com/seanmcn/chinwag/internal/lexicon"
)

// VaderScore is the per-message output of ScoreVader.
type VaderScore struct {
	Compound float64 // [-1, +1]
	Pos      float64
	Neu      float64
	Neg      float64
}

// Label classifies a compound score using VADER's standard cutoffs.
func (v VaderScore) Label() int {
	switch {
	case v.Compound >= 0.05:
		return 1
	case v.Compound <= -0.05:
		return -1
	}
	return 0
}

// VADER constants (from the canonical implementation).
const (
	vaderBIncr     = 0.293
	vaderCIncr     = 0.733 // ALL-CAPS emphasis shift
	vaderNScalar   = -0.74 // negation flips & damps valence
	vaderExclBoost = 0.292 // per '!', up to 4
	vaderQMarkB1   = 0.18  // 2 question marks
	vaderQMarkB2   = 0.96  // 3+ question marks (split across words)
)

// ScoreVader runs a pragmatic VADER-style sentiment scorer over text.
//
// Implements: lexicon lookup, ALL-CAPS emphasis, booster (degree) words,
// negation flip over a 3-token backwards window, and !/? punctuation
// boosting. Idiom phrase handling and the "but" clause re-weighting are
// deliberately omitted — they're a small contribution on chat data.
func ScoreVader(text string, lex *lexicon.Lexicons) VaderScore {
	if text == "" || lex == nil {
		return VaderScore{}
	}

	tokens := tokenizeKeepCase(text)
	if len(tokens) == 0 {
		return VaderScore{}
	}

	// Detect overall ALL-CAPS emphasis: if SOME tokens are all-caps and
	// SOME aren't, the caps ones get a boost. If everything is caps (or
	// nothing is) the boost is suppressed — same heuristic as upstream.
	capsDiff := hasMixedCaps(tokens)

	var sentiments []float64
	for i, tok := range tokens {
		lower := strings.ToLower(tok)

		// Boosters carry no valence on their own.
		if _, isBoost := lex.Boosters[lower]; isBoost {
			sentiments = append(sentiments, 0)
			continue
		}

		valence, ok := lex.Vader[lower]
		if !ok {
			sentiments = append(sentiments, 0)
			continue
		}

		// ALL-CAPS shift.
		if capsDiff && isAllCaps(tok) {
			if valence > 0 {
				valence += vaderCIncr
			} else {
				valence -= vaderCIncr
			}
		}

		// Look back up to 3 tokens for boosters and negations.
		for dist := 1; dist <= 3 && i-dist >= 0; dist++ {
			prev := strings.ToLower(tokens[i-dist])
			prevNorm := strings.ReplaceAll(prev, "'", "")
			if _, isVal := lex.Vader[prev]; isVal {
				// Don't double-count: stop at the previous sentiment word.
				break
			}
			if b, ok := lex.Boosters[prev]; ok {
				shift := b
				// Distance damping: 1.0 / 0.95 / 0.90.
				shift *= 1.0 - float64(dist-1)*0.05
				if valence < 0 {
					shift = -shift
				}
				valence += shift
			}
			if _, isNeg := lex.Negations[prev]; isNeg {
				valence *= vaderNScalar
			} else if _, isNeg := lex.Negations[prevNorm]; isNeg {
				valence *= vaderNScalar
			}
		}

		sentiments = append(sentiments, valence)
	}

	return aggregate(sentiments, text)
}

// aggregate sums per-token valences, applies punctuation boosts and
// computes the normalised compound score.
func aggregate(sentiments []float64, text string) VaderScore {
	var sum, posSum, negSum float64
	var neuCount int
	for _, v := range sentiments {
		sum += v
		switch {
		case v > 0:
			posSum += v + 1 // VADER's per-class +1 floor
		case v < 0:
			negSum += v - 1
		default:
			neuCount++
		}
	}

	// Punctuation amplifiers.
	punct := punctuationBoost(text)
	if sum > 0 {
		sum += punct
	} else if sum < 0 {
		sum -= punct
	}

	compound := sum / math.Sqrt(sum*sum+15.0)
	if compound > 1 {
		compound = 1
	} else if compound < -1 {
		compound = -1
	}

	// Per-class proportions.
	total := math.Abs(posSum) + math.Abs(negSum) + float64(neuCount)
	var pos, neg, neu float64
	if total > 0 {
		pos = math.Abs(posSum) / total
		neg = math.Abs(negSum) / total
		neu = float64(neuCount) / total
	}

	return VaderScore{Compound: compound, Pos: pos, Neu: neu, Neg: neg}
}

func punctuationBoost(text string) float64 {
	var boost float64
	if n := strings.Count(text, "!"); n > 0 {
		if n > 4 {
			n = 4
		}
		boost += float64(n) * vaderExclBoost
	}
	if n := strings.Count(text, "?"); n >= 2 {
		switch {
		case n <= 3:
			boost += vaderQMarkB1
		default:
			boost += vaderQMarkB2
		}
	}
	return boost
}

// tokenizeKeepCase splits on whitespace/punctuation but preserves casing.
func tokenizeKeepCase(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\''
	})
}

func isAllCaps(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func hasMixedCaps(tokens []string) bool {
	caps, lower := 0, 0
	for _, t := range tokens {
		if isAllCaps(t) && len([]rune(t)) > 1 {
			caps++
		} else {
			lower++
		}
	}
	return caps > 0 && lower > 0
}
