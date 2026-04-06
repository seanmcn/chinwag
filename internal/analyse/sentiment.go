package analyse

import "strings"

// Tiny embedded sentiment lexicon. Not a research-grade model — a heuristic
// that's good enough to surface tone trends in a personal chat.
var positiveWords = map[string]struct{}{
	"love": {}, "loved": {}, "loving": {}, "lovely": {}, "happy": {}, "happiness": {},
	"great": {}, "amazing": {}, "awesome": {}, "wonderful": {}, "fantastic": {},
	"good": {}, "nice": {}, "cool": {}, "thanks": {}, "thank": {}, "thx": {},
	"yay": {}, "yes": {}, "yeah": {}, "perfect": {}, "best": {}, "beautiful": {},
	"excited": {}, "fun": {}, "cute": {}, "sweet": {}, "cheers": {}, "haha": {},
	"miss": {}, "hug": {}, "hugs": {}, "kiss": {}, "smile": {}, "proud": {},
	"win": {}, "won": {}, "well": {}, "glad": {}, "enjoy": {}, "enjoyed": {},
	"congrats": {}, "congratulations": {}, "brilliant": {}, "delighted": {},
}

var negativeWords = map[string]struct{}{
	"sad": {}, "angry": {}, "mad": {}, "hate": {}, "hated": {}, "annoyed": {},
	"annoying": {}, "tired": {}, "exhausted": {}, "sick": {}, "ill": {},
	"bad": {}, "worst": {}, "awful": {}, "terrible": {}, "horrible": {},
	"upset": {}, "cry": {}, "crying": {}, "lonely": {}, "alone": {}, "stress": {},
	"stressed": {}, "anxious": {}, "worried": {}, "worry": {}, "scared": {},
	"afraid": {}, "fail": {}, "failed": {}, "lost": {}, "broken": {},
	"hurt": {}, "pain": {}, "stupid": {}, "dumb": {}, "ugh": {},
	"fight": {}, "argue": {}, "wrong": {}, "problem": {}, "issue": {},
}

// scoreSentiment returns +1 / -1 / 0 for a single message body.
func scoreSentiment(body string) int {
	pos, neg := 0, 0
	for _, w := range splitWords(strings.ToLower(body)) {
		if _, ok := positiveWords[w]; ok {
			pos++
		}
		if _, ok := negativeWords[w]; ok {
			neg++
		}
	}
	switch {
	case pos > neg:
		return 1
	case neg > pos:
		return -1
	}
	return 0
}
