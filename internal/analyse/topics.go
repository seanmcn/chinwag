package analyse

import (
	"math"
	"sort"
	"strings"

	"github.com/seanmcn/whatsapp-analyse/internal/parser"
)

// Common chat-noise words we never want to surface as a "topic".
var stopWords = map[string]struct{}{
	"the": {}, "a": {}, "an": {}, "and": {}, "or": {}, "but": {}, "if": {},
	"of": {}, "to": {}, "in": {}, "on": {}, "for": {}, "with": {}, "at": {},
	"by": {}, "from": {}, "as": {}, "is": {}, "am": {}, "are": {}, "was": {},
	"were": {}, "be": {}, "been": {}, "being": {}, "have": {}, "has": {}, "had": {},
	"do": {}, "does": {}, "did": {}, "will": {}, "would": {}, "should": {},
	"could": {}, "can": {}, "may": {}, "might": {}, "must": {}, "shall": {},
	"i": {}, "you": {}, "he": {}, "she": {}, "it": {}, "we": {}, "they": {},
	"me": {}, "him": {}, "her": {}, "us": {}, "them": {}, "my": {}, "your": {},
	"his": {}, "its": {}, "our": {}, "their": {}, "this": {}, "that": {},
	"these": {}, "those": {}, "what": {}, "which": {}, "who": {}, "whom": {},
	"so": {}, "no": {}, "not": {}, "yes": {}, "ok": {}, "okay": {}, "well": {},
	"just": {}, "now": {}, "then": {}, "here": {}, "there": {}, "when": {},
	"how": {}, "why": {}, "where": {}, "out": {}, "up": {}, "down": {}, "over": {},
	"all": {}, "any": {}, "some": {}, "more": {}, "most": {}, "much": {}, "very": {},
	"too": {}, "also": {}, "than": {}, "like": {}, "really": {}, "im": {},
	"dont": {}, "didnt": {}, "wont": {}, "cant": {}, "ive": {},
	"thats": {}, "youre": {}, "theyre": {}, "lol": {}, "haha": {}, "yeah": {},
	"yep": {}, "nope": {}, "u": {}, "ur": {}, "r": {}, "n": {}, "k": {},
	"omg": {}, "wtf": {}, "tbh": {}, "idk": {}, "got": {}, "get": {}, "go": {},
	"going": {}, "gonna": {}, "want": {}, "need": {}, "know": {}, "think": {},
	"thought": {}, "say": {}, "said": {}, "see": {}, "seen": {}, "one": {},
	"two": {}, "good": {}, "bad": {}, "today": {}, "tomorrow": {}, "yesterday": {},
}

// computeTopTerms produces top distinctive single-word terms per participant
// using a simple TF-IDF-style score: term's share for the user divided by its
// share across the whole chat. Result is up to 8 terms per user.
func computeTopTerms(msgs []parser.Message, me, them string) (map[string][]string, []string) {
	userCounts := map[string]map[string]int{me: {}, them: {}}
	totalCounts := map[string]int{}
	userTotals := map[string]int{me: 0, them: 0}

	for _, m := range msgs {
		if m.Author != me && m.Author != them {
			continue
		}
		for _, w := range splitWords(strings.ToLower(m.Body)) {
			if len(w) < 4 {
				continue
			}
			if _, skip := stopWords[w]; skip {
				continue
			}
			userCounts[m.Author][w]++
			userTotals[m.Author]++
			totalCounts[w]++
		}
	}

	totalAll := userTotals[me] + userTotals[them]
	if totalAll == 0 {
		return map[string][]string{me: nil, them: nil}, nil
	}

	type scored struct {
		term  string
		score float64
		count int
	}
	perUser := map[string][]string{}
	for u, counts := range userCounts {
		var ranked []scored
		ut := userTotals[u]
		if ut == 0 {
			perUser[u] = nil
			continue
		}
		for w, c := range counts {
			if c < 5 { // ignore noise
				continue
			}
			userShare := float64(c) / float64(ut)
			globalShare := float64(totalCounts[w]) / float64(totalAll)
			// Distinctiveness: how much more this user uses this term vs background.
			score := userShare * math.Log(1+userShare/(globalShare+1e-9))
			ranked = append(ranked, scored{w, score, c})
		}
		sort.Slice(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
		if len(ranked) > 8 {
			ranked = ranked[:8]
		}
		out := make([]string, 0, len(ranked))
		for _, r := range ranked {
			out = append(out, r.term)
		}
		perUser[u] = out
	}

	// Overall top terms by raw frequency (excluding stopwords already filtered).
	type kv struct {
		k string
		v int
	}
	var all []kv
	for w, c := range totalCounts {
		if c < 10 {
			continue
		}
		all = append(all, kv{w, c})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].v > all[j].v })
	if len(all) > 12 {
		all = all[:12]
	}
	overall := make([]string, 0, len(all))
	for _, x := range all {
		overall = append(overall, x.k)
	}

	return perUser, overall
}
