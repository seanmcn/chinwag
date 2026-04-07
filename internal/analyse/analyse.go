// Package analyse computes Stats from a slice of parsed messages.
package analyse

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/seanmcn/chinwag/internal/lexicon"
	"github.com/seanmcn/chinwag/internal/parser"
)

type UserStats struct {
	Messages      int
	Words         int
	UniqueWords   int
	Characters    int
	Emojis        int
	Laughs        int
	Apologies     int
	Questions     int
	Encouragement int
	Images        int
	Videos        int
	Audios        int
	GIFs          int
	Stickers      int
	Links         int
	TopEmojis     []EmojiCount
	ConvosStarted int
	ConvosClosed  int
	ConvosMissed  int
	TopContrib    int     // convos where this user was the top contributor
	AvgConvoPts   float64 // avg points contributed per convo
	Reconnects    int
	DoubleMsgs    int
	RapidFirstPct int   // % first replies < 5 min
	AvgFirstResp  int64 // seconds
	AvgResponse   int64 // seconds
	MedianResp    int64 // seconds (p50)
	P90Response   int64 // seconds (p90)
	Positive      int   // sentiment-positive messages (VADER compound >= 0.05)
	Negative      int   // sentiment-negative messages (VADER compound <= -0.05)
	SentimentAvg  float64
	// VADER compound average across this user's messages, in [-1, +1].
	CompoundAvg float64
	// Per-NRC-category counts (indexed by lexicon.EmotionCategories order:
	// anger, anticipation, disgust, fear, joy, sadness, surprise, trust,
	// negative, positive).
	Emotion [10]int
	// Summed NRC affect-intensity per category.
	IntensitySum [10]float64
	// Single peak intensity score across all messages (max of summed
	// intensities for any single message).
	IntensityPeak float64
	// Average VAD vector across messages that contained any VAD-known word.
	VAD          [3]float64
	VADMessages  int // denominator for VAD averaging
	ScoredMsgs   int // denominator for CompoundAvg
	PeakHour      int    // hour of day with most messages
	Chronotype    string // "Early bird" / "Daytime" / "Night owl"
	TopTerms      []string
	ReplyByHour   [24]int64 // avg seconds, indexed by hour-of-day of the reply
}

type EmojiCount struct {
	Emoji string
	Count int
}

type GrowthPoint struct {
	Month  string         `json:"month"`  // YYYY-MM
	Counts map[string]int `json:"counts"`
}

type SentimentPoint struct {
	Month string `json:"month"` // YYYY-MM
	Pos   int    `json:"pos"`
	Neg   int    `json:"neg"`
	Net   int    `json:"net"`
}

type DomainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type DayCount struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int    `json:"count"`
}

type SankeyLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

type ConvoStats struct {
	Total      int
	BySize     map[string]int // "Big moments" / "Everyday chats" / "No reply"
	Sankey     []SankeyLink
	GapMinutes int
}

type Stats struct {
	Participants  [2]string
	Period        struct{ Start, End time.Time }
	Messages      int
	Conversations int
	ChatPoints    int
	PerUser       map[string]*UserStats
	Timeline      []GrowthPoint
	Heatmap       [7][24]int
	DailyActivity []DayCount
	Convos        ConvoStats
	Insights      []string
	Rating        int
	RatingLabel   string
	Balance       map[string]int
	BalancePct    map[string]int
	TopWeekdayHr  string
	CharsTyped    int
	TimeTyping    string
	Direction     map[string]int // user / other (always 16% other for now derived)
	LongestStreak     int            // consecutive days with at least one message
	CurrentStreak     int
	TopTerms          []string         // overall distinctive terms across the whole chat
	SentimentTimeline []SentimentPoint // per-month sentiment net
	TopDomains        []DomainCount    // most-shared link hosts
}

// Run computes all stats. participants[0] is "me", [1] is the other.
func Run(msgs []parser.Message, me, them string, gap time.Duration) Stats {
	var s Stats
	lex, _ := lexicon.Load() // best-effort; nil-safe in scoring
	if me == "" || them == "" {
		// auto-detect top 2 authors
		counts := map[string]int{}
		for _, m := range msgs {
			counts[m.Author]++
		}
		type kv struct {
			k string
			v int
		}
		var ks []kv
		for k, v := range counts {
			ks = append(ks, kv{k, v})
		}
		sort.Slice(ks, func(i, j int) bool { return ks[i].v > ks[j].v })
		pick := func(skip string) string {
			for _, x := range ks {
				if x.k != skip {
					return x.k
				}
			}
			return ""
		}
		if me == "" {
			me = pick(them)
		}
		if them == "" {
			them = pick(me)
		}
	}
	s.Participants = [2]string{me, them}
	s.PerUser = map[string]*UserStats{me: {}, them: {}}
	if len(msgs) == 0 {
		return s
	}
	s.Period.Start = msgs[0].Timestamp
	s.Period.End = msgs[len(msgs)-1].Timestamp

	// Per-message stats
	emojiCounts := map[string]map[string]int{me: {}, them: {}}
	uniqueWords := map[string]map[string]struct{}{me: {}, them: {}}
	dailyMap := map[string]int{}
	growthMap := map[string]map[string]int{}
	hourHist := map[string][24]int{me: {}, them: {}}
	sentMap := map[string]*SentimentPoint{} // month -> running totals
	var prevAuthor string
	for _, m := range msgs {
		us, ok := s.PerUser[m.Author]
		if !ok {
			continue // ignore third parties
		}
		s.Messages++
		us.Messages++
		us.Characters += len([]rune(m.Body))
		s.CharsTyped += len([]rune(m.Body))
		words := splitWords(m.Body)
		us.Words += len(words)
		for _, w := range words {
			uniqueWords[m.Author][strings.ToLower(w)] = struct{}{}
		}
		// Counts
		us.Questions += strings.Count(m.Body, "?")
		lower := strings.ToLower(m.Body)
		if containsAny(lower, []string{"sorry", "apolog", "my bad"}) {
			us.Apologies++
		}
		if containsLaugh(lower) {
			us.Laughs++
		}
		if containsAny(lower, []string{"well done", "proud of you", "you got this", "amazing", "great job", "good luck", "you can do"}) {
			us.Encouragement++
		}
		// Sentiment
		mk := m.Timestamp.Format("2006-01")
		sp := sentMap[mk]
		if sp == nil {
			sp = &SentimentPoint{Month: mk}
			sentMap[mk] = sp
		}
		// Sentiment + emotion (VADER + NRC)
		score := ScoreMessage(m.Body, lex)
		us.CompoundAvg += score.Vader.Compound
		us.ScoredMsgs++
		switch score.Vader.Label() {
		case 1:
			us.Positive++
			sp.Pos++
		case -1:
			us.Negative++
			sp.Neg++
		}
		// NRC emotion / intensity / VAD
		var msgIntensity float64
		for i := 0; i < 10; i++ {
			us.Emotion[i] += int(score.Emotion[i])
			us.IntensitySum[i] += score.Intensity[i]
			msgIntensity += score.Intensity[i]
		}
		if msgIntensity > us.IntensityPeak {
			us.IntensityPeak = msgIntensity
		}
		if score.VADWords > 0 {
			inv := 1.0 / float64(score.VADWords)
			us.VAD[0] += score.VAD[0] * inv
			us.VAD[1] += score.VAD[1] * inv
			us.VAD[2] += score.VAD[2] * inv
			us.VADMessages++
		}
		// Emoji
		for _, r := range m.Body {
			if isEmoji(r) {
				us.Emojis++
				emojiCounts[m.Author][string(r)]++
			}
		}
		// Media
		switch m.Kind {
		case parser.KindImage:
			us.Images++
		case parser.KindVideo:
			us.Videos++
		case parser.KindAudio:
			us.Audios++
		case parser.KindGIF:
			us.GIFs++
		case parser.KindSticker:
			us.Stickers++
		case parser.KindLink:
			us.Links++
		}
		// Heatmap & daily
		s.Heatmap[int(m.Timestamp.Weekday())][m.Timestamp.Hour()]++
		if h, ok := hourHist[m.Author]; ok {
			h[m.Timestamp.Hour()]++
			hourHist[m.Author] = h
		}
		dailyMap[m.Timestamp.Format("2006-01-02")]++
		monthKey := m.Timestamp.Format("2006-01")
		if _, ok := growthMap[monthKey]; !ok {
			growthMap[monthKey] = map[string]int{}
		}
		growthMap[monthKey][m.Author]++
		// Double messages (same author back-to-back)
		if m.Author == prevAuthor {
			us.DoubleMsgs++
		}
		prevAuthor = m.Author
	}
	for u, ec := range emojiCounts {
		s.PerUser[u].UniqueWords = len(uniqueWords[u])
		var list []EmojiCount
		for k, v := range ec {
			list = append(list, EmojiCount{k, v})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].Count > list[j].Count })
		if len(list) > 5 {
			list = list[:5]
		}
		s.PerUser[u].TopEmojis = list
	}
	// Timeline
	var months []string
	for k := range growthMap {
		months = append(months, k)
	}
	sort.Strings(months)
	for _, k := range months {
		s.Timeline = append(s.Timeline, GrowthPoint{Month: k, Counts: growthMap[k]})
		if sp := sentMap[k]; sp != nil {
			sp.Net = sp.Pos - sp.Neg
			s.SentimentTimeline = append(s.SentimentTimeline, *sp)
		}
	}
	// Daily activity (last 500 days up to End)
	end := s.Period.End
	for i := 499; i >= 0; i-- {
		d := end.AddDate(0, 0, -i).Format("2006-01-02")
		s.DailyActivity = append(s.DailyActivity, DayCount{Date: d, Count: dailyMap[d]})
	}
	// Top weekday/hour
	bestWD, bestHR, best := 0, 0, 0
	for wd := 0; wd < 7; wd++ {
		for hr := 0; hr < 24; hr++ {
			if s.Heatmap[wd][hr] > best {
				best = s.Heatmap[wd][hr]
				bestWD = wd
				bestHR = hr
			}
		}
	}
	s.TopWeekdayHr = fmt.Sprintf("%s around %02d:00", time.Weekday(bestWD).String(), bestHR)
	// Time typing: assume ~5 chars/sec
	secs := s.CharsTyped / 5
	d := time.Duration(secs) * time.Second
	days := int(d.Hours()) / 24
	hrs := int(d.Hours()) % 24
	s.TimeTyping = fmt.Sprintf("%d days %d hours", days, hrs)

	// Per-user peak hour & chronotype
	for u, h := range hourHist {
		peak, peakV := 0, -1
		for hr, v := range h {
			if v > peakV {
				peakV, peak = v, hr
			}
		}
		s.PerUser[u].PeakHour = peak
		switch {
		case peak < 10:
			s.PerUser[u].Chronotype = "Early bird"
		case peak >= 22 || peak < 4:
			s.PerUser[u].Chronotype = "Night owl"
		default:
			s.PerUser[u].Chronotype = "Daytime"
		}
		// Sentiment average
		if s.PerUser[u].Messages > 0 {
			s.PerUser[u].SentimentAvg = float64(s.PerUser[u].Positive-s.PerUser[u].Negative) / float64(s.PerUser[u].Messages)
		}
		// VADER compound average + VAD averages
		us := s.PerUser[u]
		if us.ScoredMsgs > 0 {
			us.CompoundAvg /= float64(us.ScoredMsgs)
		}
		if us.VADMessages > 0 {
			inv := 1.0 / float64(us.VADMessages)
			us.VAD[0] *= inv
			us.VAD[1] *= inv
			us.VAD[2] *= inv
		}
	}

	// Streaks (consecutive days with at least one message), based on full daily map
	var dayKeys []string
	for k := range dailyMap {
		dayKeys = append(dayKeys, k)
	}
	sort.Strings(dayKeys)
	longest, current := 0, 0
	var prevDay time.Time
	for i, k := range dayKeys {
		d, _ := time.Parse("2006-01-02", k)
		if i == 0 || d.Sub(prevDay) == 24*time.Hour {
			current++
		} else {
			current = 1
		}
		if current > longest {
			longest = current
		}
		prevDay = d
	}
	s.LongestStreak = longest
	// Current streak: only counts if it runs up to (or including) the last day in the data
	if len(dayKeys) > 0 {
		last, _ := time.Parse("2006-01-02", dayKeys[len(dayKeys)-1])
		if s.Period.End.Sub(last) <= 24*time.Hour {
			s.CurrentStreak = current
		}
	}

	// Conversations
	convos := segment(msgs, gap, me, them)
	s.Conversations = len(convos)
	s.Convos.GapMinutes = int(gap.Minutes())
	s.Convos.BySize = map[string]int{"Big moments": 0, "Everyday chats": 0, "No reply": 0}
	sankey := map[string]int{} // "src->tgt" -> v
	addLink := func(src, tgt string) { sankey[src+"->"+tgt]++ }

	for _, c := range convos {
		starter := c.Messages[0].Author
		closer := c.Messages[len(c.Messages)-1].Author
		s.PerUser[starter].ConvosStarted++
		s.PerUser[closer].ConvosClosed++
		// missed: only one author
		oneAuthor := true
		for _, m := range c.Messages {
			if m.Author != starter {
				oneAuthor = false
				break
			}
		}
		if oneAuthor {
			s.PerUser[starter].ConvosMissed++
		}
		// classify size
		var category string
		switch {
		case oneAuthor:
			category = "No reply"
		case len(c.Messages) >= 30:
			category = "Big moments"
		default:
			category = "Everyday chats"
		}
		s.Convos.BySize[category]++

		// top contributor
		ucounts := map[string]int{}
		for _, m := range c.Messages {
			ucounts[m.Author]++
		}
		var top string
		topV := -1
		for u, v := range ucounts {
			if v > topV {
				topV, top = v, u
			}
		}
		s.PerUser[top].TopContrib++
		s.PerUser[top].AvgConvoPts += float64(len(c.Messages))

		addLink("Started by "+starter, category)
		addLink(category, "Major: "+top)
		addLink("Major: "+top, "Final: "+closer)
	}
	for u, us := range s.PerUser {
		if us.TopContrib > 0 {
			s.PerUser[u].AvgConvoPts = us.AvgConvoPts / float64(us.TopContrib)
		}
	}
	// Sankey links
	for k, v := range sankey {
		parts := strings.SplitN(k, "->", 2)
		s.Convos.Sankey = append(s.Convos.Sankey, SankeyLink{Source: parts[0], Target: parts[1], Value: v})
	}
	sort.Slice(s.Convos.Sankey, func(i, j int) bool { return s.Convos.Sankey[i].Source < s.Convos.Sankey[j].Source })

	// Response times
	computeResponses(convos, s.PerUser)

	// Reconnects: convos where the previous convo's last author == this convo's starter and gap > 7 days
	for i := 1; i < len(convos); i++ {
		prev := convos[i-1].Messages[len(convos[i-1].Messages)-1]
		next := convos[i].Messages[0]
		if next.Timestamp.Sub(prev.Timestamp) > 7*24*time.Hour && prev.Author != next.Author {
			s.PerUser[next.Author].Reconnects++
		}
	}

	// Balance
	s.Balance = map[string]int{me: 0, them: 0}
	for u, us := range s.PerUser {
		s.Balance[u] = us.Messages*5 + us.Words + us.Emojis*2 + us.Images*10 + us.Videos*15
	}
	s.ChatPoints = s.Balance[me] + s.Balance[them]
	s.BalancePct = map[string]int{}
	if s.ChatPoints > 0 {
		s.BalancePct[me] = s.Balance[me] * 100 / s.ChatPoints
		s.BalancePct[them] = 100 - s.BalancePct[me]
	}

	// Top link domains
	s.TopDomains = computeTopDomains(msgs)

	// Topic / distinctive term extraction
	perTerms, overallTerms := computeTopTerms(msgs, me, them, lex)
	s.TopTerms = overallTerms
	for u, terms := range perTerms {
		if us, ok := s.PerUser[u]; ok {
			us.TopTerms = terms
		}
	}

	s.Insights = buildInsights(s.PerUser, me, them, s)
	s.Rating, s.RatingLabel = computeRating(s)
	s.Direction = map[string]int{me: 47, them: 37, "Other people": 16} // placeholder split
	return s
}

func splitWords(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func containsLaugh(s string) bool {
	return strings.Contains(s, "haha") || strings.Contains(s, "lol") || strings.Contains(s, "lmao") || strings.Contains(s, "rofl") || strings.Contains(s, "😂") || strings.Contains(s, "🤣")
}

func isEmoji(r rune) bool {
	return (r >= 0x1F300 && r <= 0x1FAFF) ||
		(r >= 0x2600 && r <= 0x27BF) ||
		(r >= 0x1F1E6 && r <= 0x1F1FF)
}
