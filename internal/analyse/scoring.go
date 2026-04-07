package analyse

import (
	"strings"

	"github.com/seanmcn/chinwag/internal/lexicon"
)

// MessageScore is the per-message NLP output: VADER sentiment plus the
// NRC-derived emotion / intensity / VAD signals.
type MessageScore struct {
	Vader     VaderScore
	Emotion   [10]float64 // counts per category, indexed by lexicon.EmotionCategories
	Intensity [10]float64 // summed intensity per category
	VAD       [3]float64  // summed valence/arousal/dominance
	VADWords  int         // denominator for VAD averaging
}

// ScoreMessage runs all per-message NLP signals over body in one pass.
func ScoreMessage(body string, lex *lexicon.Lexicons) MessageScore {
	var ms MessageScore
	if lex == nil || body == "" {
		return ms
	}
	ms.Vader = ScoreVader(body, lex)

	for _, w := range splitWords(strings.ToLower(body)) {
		if mask, ok := lex.EmotionMask[w]; ok {
			for bit := 0; bit < len(lex.EmotionCategories) && bit < 10; bit++ {
				if mask&(1<<uint(bit)) != 0 {
					ms.Emotion[bit]++
				}
			}
		}
		if cats, ok := lex.Intensity[w]; ok {
			for catName, strength := range cats {
				if idx := emotionIndex(lex, catName); idx >= 0 {
					ms.Intensity[idx] += float64(strength)
				}
			}
		}
		if vad, ok := lex.VAD[w]; ok {
			ms.VAD[0] += float64(vad[0])
			ms.VAD[1] += float64(vad[1])
			ms.VAD[2] += float64(vad[2])
			ms.VADWords++
		}
	}
	return ms
}

// emotionIndex maps a category name to its bit position. Linear scan over
// 10 entries — cheaper than a map for this size.
func emotionIndex(lex *lexicon.Lexicons, name string) int {
	for i, c := range lex.EmotionCategories {
		if c == name {
			return i
		}
	}
	return -1
}
