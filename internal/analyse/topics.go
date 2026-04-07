package analyse

import (
	"math"
	"sort"
	"strings"

	"github.com/seanmcn/chinwag/internal/lexicon"
	"github.com/seanmcn/chinwag/internal/parser"
)

// topicFiller is a topic-extraction-only stopword extension. These are
// high-frequency conversational filler words that are never useful as
// "topics" but are deliberately NOT in the general lexicon.Stopwords list
// because other parts of the NLP pipeline (sentiment, emotion) may need
// them. Keep emotionally-salient verbs like feel/want/need/love/hate OUT.
var topicFiller = map[string]struct{}{
	"about": {}, "back": {}, "nice": {}, "probably": {}, "though": {},
	"cool": {}, "something": {}, "work": {}, "really": {}, "actually": {},
	"thing": {}, "things": {}, "stuff": {}, "sure": {}, "fine": {},
	"well": {}, "much": {}, "lots": {}, "ways": {}, "also": {},
	"even": {}, "still": {}, "already": {}, "maybe": {}, "perhaps": {},
	"definitely": {}, "basically": {}, "literally": {}, "honestly": {},
	"totally": {}, "pretty": {}, "kind": {}, "sort": {}, "little": {},
	"good": {}, "right": {}, "wrong": {}, "same": {}, "different": {},
	"early": {}, "late": {}, "soon": {}, "today": {}, "tomorrow": {},
	"yesterday": {}, "tonight": {}, "morning": {}, "evening": {}, "night": {},
	"week": {}, "month": {}, "year": {}, "years": {}, "days": {},
	"weeks": {}, "months": {}, "hours": {}, "minutes": {}, "later": {},
	"before": {}, "after": {}, "everyone": {}, "anyone": {}, "someone": {},
	"nobody": {}, "everything": {}, "anything": {}, "nothing": {},
	"going": {}, "gonna": {}, "getting": {}, "making": {}, "taking": {},
	"coming": {}, "saying": {}, "telling": {}, "asking": {}, "trying": {},
	"using": {}, "calling": {}, "helping": {}, "working": {}, "seeming": {},
	"turning": {}, "leaving": {}, "keeping": {}, "starting": {}, "stopping": {},
	"showing": {}, "running": {}, "playing": {}, "moving": {}, "living": {},
	"believing": {}, "happening": {}, "meaning": {}, "letting": {},
	// Grammatical hedges and reaction filler — these are real words but
	// rarely actual topics in chat data.
	"would": {}, "could": {}, "might": {}, "should": {}, "shall": {},
	"think": {}, "thinks": {}, "thinking": {}, "guess": {}, "guessing": {},
	"imagine": {}, "suppose": {}, "reckon": {}, "wonder": {},
	"sounds": {}, "looks": {}, "seems": {}, "feels": {},
	"great": {}, "either": {}, "neither": {}, "certainly": {},
	"quite": {}, "sweet": {}, "alright": {}, "classic": {}, "completely": {},
	"grand": {}, "obviously": {}, "clearly": {}, "exactly": {}, "absolutely": {},
	"hilarious": {}, "amazing": {}, "awesome": {}, "brilliant": {},
	"point": {}, "bunch": {}, "outside": {}, "inside": {},
	"people": {}, "least": {}, "around": {}, "thought": {},
	"better": {}, "start": {}, "first": {},
	// Tech-protocol words that show up as real content in dev chats but
	// are never useful as "topics".
	"https": {}, "http": {},
}

// computeTopTerms produces top distinctive single-word terms per participant
// using a simple TF-IDF-style score: term's share for the user divided by its
// share across the whole chat. Result is up to 8 terms per user.
func computeTopTerms(msgs []parser.Message, me, them string, lex *lexicon.Lexicons) (map[string][]string, []string) {
	userCounts := map[string]map[string]int{me: {}, them: {}}
	totalCounts := map[string]int{}
	userTotals := map[string]int{me: 0, them: 0}

	for _, m := range msgs {
		if m.Author != me && m.Author != them {
			continue
		}
		// Only real text contributes to vocabulary. KindImage etc. carry
		// placeholder strings like "<Media omitted>" that pollute results.
		if m.Kind != parser.KindText {
			continue
		}
		body := stripURLs(stripWhatsAppMarkers(strings.ToLower(m.Body)))
		for _, w := range splitWords(body) {
			if len(w) < 5 {
				continue
			}
			if lex != nil {
				if _, skip := lex.Stopwords[w]; skip {
					continue
				}
			}
			if _, skip := topicFiller[w]; skip {
				continue
			}
			if isLaughToken(w) {
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
			if c < 8 { // ignore noise
				continue
			}
			userShare := float64(c) / float64(ut)
			globalShare := float64(totalCounts[w]) / float64(totalAll)
			// Distinctiveness floor: only keep terms this user uses
			// meaningfully more than the chat baseline.
			if userShare < 1.5*globalShare {
				continue
			}
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

	// Shared vocabulary: top frequencies, but minus anything we already
	// flagged as distinctive to one user (a word can't be both).
	claimed := map[string]struct{}{}
	for _, list := range perUser {
		for _, w := range list {
			claimed[w] = struct{}{}
		}
	}
	type kv struct {
		k string
		v int
	}
	var all []kv
	for w, c := range totalCounts {
		if c < 15 {
			continue
		}
		if _, taken := claimed[w]; taken {
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

// stripWhatsAppMarkers removes WhatsApp's inline annotations that are
// part of the export format, not part of what the user actually typed.
// Expects lowercase input. Only used for topic extraction; the rest of
// the analyser works against the raw body.
//
// Mixed messages like "Something to do with this image omitted" leave
// the placeholder text inline next to real content, so per-message Kind
// filtering isn't enough — we have to scrub the markers themselves.
func stripWhatsAppMarkers(s string) string {
	markers := []string{
		"<this message was edited>",
		"<media omitted>",
		"image omitted",
		"video omitted",
		"audio omitted",
		"sticker omitted",
		"gif omitted",
		"document omitted",
		"contact card omitted",
	}
	for _, m := range markers {
		s = strings.ReplaceAll(s, m, " ")
	}
	return s
}

// isLaughToken collapses every variant of haha/hehe/lol/loooool/lmao/rofl
// into a single "drop this" check. The trick: strings.Trim(w, "ha")
// removes any leading/trailing chars in the cutset {h,a}, and if the
// result is empty the whole word was made of those chars (so haha,
// hahah, hahahah, ahha, ahahaha, etc. all qualify).
func isLaughToken(w string) bool {
	if len(w) < 3 {
		return false
	}
	if strings.Trim(w, "ha") == "" {
		return true
	}
	if strings.Trim(w, "he") == "" {
		return true
	}
	if strings.Trim(w, "lo") == "" {
		return true
	}
	switch w {
	case "lmao", "lmfao", "rofl", "roflmao":
		return true
	}
	return false
}

// stripURLs removes anything that looks like a URL from text. Cheap
// token-level pass — no regex. We replace each match with a space so the
// surrounding words still tokenize cleanly.
func stripURLs(s string) string {
	if !strings.ContainsAny(s, ":w") {
		return s
	}
	fields := strings.Fields(s)
	for i, f := range fields {
		lower := strings.ToLower(f)
		if strings.HasPrefix(lower, "http://") ||
			strings.HasPrefix(lower, "https://") ||
			strings.HasPrefix(lower, "www.") ||
			strings.Contains(lower, "://") {
			fields[i] = " "
		}
	}
	return strings.Join(fields, " ")
}
