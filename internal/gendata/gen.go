// Package gendata generates a deterministic, fictional WhatsApp-format chat
// between two people. The output matches the iOS bracket format accepted by
// internal/parser. It exists so the project can ship a richly populated sample
// chat for screenshots and demos without committing real conversations.
package gendata

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

// Options controls the generated chat. Zero values get sensible defaults via
// Generate.
type Options struct {
	Seed     int64
	Start    time.Time
	Days     int
	MeName   string
	ThemName string
}

// Generate writes a WhatsApp-format chat to w. Output is fully deterministic
// for a given Options value.
func Generate(w io.Writer, opts Options) error {
	if opts.Seed == 0 {
		opts.Seed = 42
	}
	if opts.Start.IsZero() {
		opts.Start = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	}
	if opts.Days == 0 {
		opts.Days = 400
	}
	if opts.MeName == "" {
		opts.MeName = "Alice"
	}
	if opts.ThemName == "" {
		opts.ThemName = "Bob"
	}

	rng := rand.New(rand.NewSource(opts.Seed))
	bw := &bufWriter{w: w}

	for d := 0; d < opts.Days; d++ {
		day := opts.Start.AddDate(0, 0, d)
		mood := moodFor(d, opts.Days)

		// Some days are silent: occasional 1–2 day gaps, plus rare 3–5 day silences
		// during the rough patch.
		if mood.silenceChance > 0 && rng.Float64() < mood.silenceChance {
			continue
		}

		nMsgs := mood.minMsgs + rng.Intn(mood.maxMsgs-mood.minMsgs+1)
		writeDay(bw, rng, day, nMsgs, mood, opts.MeName, opts.ThemName)
		if bw.err != nil {
			return bw.err
		}
	}
	return bw.err
}

type bufWriter struct {
	w   io.Writer
	err error
}

func (b *bufWriter) printf(format string, args ...interface{}) {
	if b.err != nil {
		return
	}
	_, b.err = fmt.Fprintf(b.w, format, args...)
}

// monthMood describes the emotional shape of a stretch of days.
type monthMood struct {
	label         string
	posWeight     float64 // 0..1 chance a message draws from positive pools
	negWeight     float64 // 0..1 chance from negative pools
	excitedWeight float64
	apologyChance float64
	emojiChance   float64
	mediaChance   float64
	linkChance    float64
	questionBias  float64
	minMsgs       int
	maxMsgs       int
	silenceChance float64
}

// moodFor returns a mood profile for a given day index. The arc is:
//
//	0–60   honeymoon: lots of messages, very warm
//	61–150 settled positive
//	151–200 rough patch: fewer messages, more apologies, occasional silences, negative tone
//	201–260 recovery
//	261+   strong positive finish
func moodFor(dayIdx, totalDays int) monthMood {
	switch {
	case dayIdx < 60:
		return monthMood{
			label: "honeymoon", posWeight: 0.75, excitedWeight: 0.4,
			emojiChance: 0.35, mediaChance: 0.06, linkChance: 0.04,
			questionBias: 0.25, apologyChance: 0.01,
			minMsgs: 8, maxMsgs: 22, silenceChance: 0.02,
		}
	case dayIdx < 150:
		return monthMood{
			label: "settled", posWeight: 0.55, excitedWeight: 0.2,
			emojiChance: 0.22, mediaChance: 0.05, linkChance: 0.05,
			questionBias: 0.2, apologyChance: 0.02,
			minMsgs: 4, maxMsgs: 16, silenceChance: 0.08,
		}
	case dayIdx < 200:
		return monthMood{
			label: "rough", posWeight: 0.18, negWeight: 0.45, excitedWeight: 0.05,
			emojiChance: 0.1, mediaChance: 0.02, linkChance: 0.02,
			questionBias: 0.3, apologyChance: 0.18,
			minMsgs: 2, maxMsgs: 9, silenceChance: 0.22,
		}
	case dayIdx < 260:
		return monthMood{
			label: "recovery", posWeight: 0.5, negWeight: 0.1, excitedWeight: 0.15,
			emojiChance: 0.18, mediaChance: 0.04, linkChance: 0.04,
			questionBias: 0.25, apologyChance: 0.05,
			minMsgs: 3, maxMsgs: 12, silenceChance: 0.1,
		}
	default:
		return monthMood{
			label: "finale", posWeight: 0.78, excitedWeight: 0.45,
			emojiChance: 0.4, mediaChance: 0.07, linkChance: 0.05,
			questionBias: 0.22, apologyChance: 0.01,
			minMsgs: 9, maxMsgs: 24, silenceChance: 0.02,
		}
	}
}

func writeDay(bw *bufWriter, rng *rand.Rand, day time.Time, n int, mood monthMood, me, them string) {
	// Pick a starting hour. Weekday vs weekend, plus chronotype bias for whoever
	// starts the conversation. Alice = early bird, Bob = night owl.
	weekend := day.Weekday() == time.Saturday || day.Weekday() == time.Sunday

	// Build a sequence of bursts. Most days have 1 burst; some have 2 or 3.
	bursts := 1
	if rng.Float64() < 0.35 {
		bursts = 2
	}
	if rng.Float64() < 0.1 {
		bursts = 3
	}

	for b := 0; b < bursts; b++ {
		// Who starts? Mornings tend to be Alice, evenings tend to be Bob, with noise.
		starterIsMe := rng.Float64() < 0.5
		var hour int
		switch b {
		case 0:
			if starterIsMe {
				hour = 7 + rng.Intn(4) // 7..10
			} else {
				hour = 18 + rng.Intn(5) // 18..22
			}
		case 1:
			hour = 12 + rng.Intn(4) // lunchtime burst
			starterIsMe = rng.Float64() < 0.55
		default:
			hour = 21 + rng.Intn(3) // late
			starterIsMe = rng.Float64() < 0.35 // bob more likely
		}
		if weekend {
			// Weekends spread later
			hour = clamp(hour+rng.Intn(3)-1, 8, 23)
		}
		minute := rng.Intn(60)
		second := rng.Intn(60)
		t := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, second, 0, day.Location())

		burstSize := n/bursts + rng.Intn(3)
		if burstSize < 2 {
			burstSize = 2
		}

		speaker := them
		if starterIsMe {
			speaker = me
		}

		for i := 0; i < burstSize; i++ {
			body := composeBody(rng, mood, speaker == me, i == 0, i == burstSize-1)
			bw.printf("[%s] %s: %s\n",
				t.Format("02/01/2006, 15:04:05"), speaker, body)

			// Reply gap: usually short, occasionally long.
			gap := replyGap(rng)
			t = t.Add(gap)
			// Don't roll past midnight too often.
			if t.Day() != day.Day() {
				break
			}

			// Alternate speaker most of the time, sometimes double-text.
			if rng.Float64() < 0.78 {
				if speaker == me {
					speaker = them
				} else {
					speaker = me
				}
			}
		}
	}
}

func replyGap(rng *rand.Rand) time.Duration {
	r := rng.Float64()
	switch {
	case r < 0.6:
		// snappy: 10s..3min
		return time.Duration(10+rng.Intn(170)) * time.Second
	case r < 0.85:
		// medium: 3..20min
		return time.Duration(3+rng.Intn(17)) * time.Minute
	case r < 0.97:
		// slow: 20min..2h
		return time.Duration(20+rng.Intn(100)) * time.Minute
	default:
		// long: 2..6h
		return time.Duration(120+rng.Intn(240)) * time.Minute
	}
}

func composeBody(rng *rand.Rand, mood monthMood, isMe, first, last bool) string {
	// Media / link special cases first.
	if rng.Float64() < mood.mediaChance {
		return pick(rng, mediaLines)
	}
	if rng.Float64() < mood.linkChance {
		return pick(rng, linkLines)
	}

	// Per-user signature lines (hobbies, places, tells) so each side has
	// distinctive vocabulary the topic extractor can pick out. Needs to fire
	// often enough that each signature word lands ≥8 times across the chat.
	if rng.Float64() < 0.18 {
		if isMe {
			return pick(rng, aliceSignatures)
		}
		return pick(rng, bobSignatures)
	}

	// Per-user sentiment bias: Alice trends warmer/more excited, Bob cooler
	// and more frustrated. Applied as additive shifts on top of the day mood.
	posW, negW, excW := mood.posWeight, mood.negWeight, mood.excitedWeight
	if isMe {
		posW += 0.15
		excW += 0.1
		negW = maxF(0, negW-0.1)
	} else {
		negW += 0.15
		posW = maxF(0, posW-0.15)
	}

	var pool []string
	r := rng.Float64()
	switch {
	case first && rng.Float64() < 0.6:
		pool = greetings
	case rng.Float64() < mood.apologyChance:
		pool = apologies
	case r < negW:
		pool = pickPool(rng, sadnessPhrases, fearPhrases, frustrationPhrases)
	case r < negW+excW:
		pool = excitementPhrases
	case r < negW+excW+posW:
		pool = pickPool(rng, plansPositive, encouragement, joyPhrases, trustPhrases)
	default:
		pool = chitchat
	}

	body := pick(rng, pool)

	// Sometimes turn it into a question.
	if rng.Float64() < mood.questionBias && !strings.HasSuffix(body, "?") {
		body = pick(rng, questions)
	}

	// Laughs.
	if rng.Float64() < 0.08 {
		body = body + " " + pick(rng, laughs)
	}

	// Emojis. Per-user pools so the top-5 lists differ.
	if rng.Float64() < mood.emojiChance {
		var pool []string
		if isMe {
			pool = aliceEmojis
		} else {
			pool = bobEmojis
		}
		body = body + " " + pick(rng, pool)
	}

	// Closer flavour.
	if last && rng.Float64() < 0.25 {
		body = pick(rng, closers)
	}

	return body
}

func pick(rng *rand.Rand, pool []string) string {
	return pool[rng.Intn(len(pool))]
}

func pickPool(rng *rand.Rand, pools ...[]string) []string {
	return pools[rng.Intn(len(pools))]
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// --- phrase pools ---------------------------------------------------------

var greetings = []string{
	"hey!", "morning ☀️", "hiya", "good morning", "yo", "heya", "hey you",
	"hellooo", "hey hey", "morning! how'd you sleep?", "afternoon!",
	"evening :)", "hi!", "hey, you around?",
}

var chitchat = []string{
	"just got back from the shops",
	"work is wild today",
	"making lunch in a bit",
	"the dog is being a menace again",
	"on the train",
	"finally sitting down",
	"phone was on silent sorry",
	"writing this from bed",
	"in a meeting, brb",
	"queue at the post office is unreal",
	"forgot my umbrella, classic",
	"running on coffee today",
}

var plansPositive = []string{
	"want to grab coffee tomorrow?",
	"dinner Friday? new place opened on Bridge St",
	"I booked the cinema for Saturday",
	"come over later? I'll cook",
	"weekend plans? was thinking a walk",
	"got tickets for the gig next month, you free?",
	"brunch Sunday is happening",
	"let's do something nice this weekend",
	"I picked up wine for tonight",
	"we should plan a little trip",
}

var encouragement = []string{
	"you've got this",
	"so proud of you",
	"that's brilliant news!",
	"honestly, well done",
	"you'll smash it",
	"don't worry, you're great at this",
	"that sounds amazing, go for it",
	"I believe in you, truly",
	"you handled that so well",
}

var joyPhrases = []string{
	"today was honestly perfect",
	"I'm in such a good mood right now",
	"that made me so happy",
	"loving life today",
	"feeling really lucky",
	"the sun is out and I'm beaming",
	"can't stop smiling tbh",
	"such a lovely morning",
}

var trustPhrases = []string{
	"thanks for always listening",
	"I really appreciate you",
	"means a lot, honestly",
	"you always know what to say",
	"glad I can talk to you about this",
	"thank you for being there",
}

var excitementPhrases = []string{
	"OMG yes!!",
	"no waaay",
	"I'm SO excited",
	"this is the best news!!",
	"I literally can't wait",
	"shut up that's amazing",
	"YES YES YES",
	"finally!! 🙌",
}

var sadnessPhrases = []string{
	"feeling pretty low today",
	"I'm just exhausted",
	"don't really want to talk about it",
	"having a rough one",
	"my head is everywhere",
	"I miss how things were",
	"just sad I guess",
	"crying a bit not gonna lie",
}

var fearPhrases = []string{
	"I'm worried about the meeting tomorrow",
	"this whole thing is making me anxious",
	"I'm scared I'm going to mess it up",
	"nervous about the results",
	"can't stop overthinking it",
}

var frustrationPhrases = []string{
	"I'm so annoyed right now",
	"this is ridiculous honestly",
	"why is everything so hard today",
	"I hate when this happens",
	"fed up with it tbh",
	"that really wound me up",
}

var apologies = []string{
	"sorry, that came out wrong",
	"I'm sorry, really",
	"sorry I snapped earlier",
	"my bad, I was being unfair",
	"sorry — long day, not an excuse",
	"I'm sorry I missed your call",
	"sorry for going quiet",
}

var questions = []string{
	"how was your day?",
	"what are you up to?",
	"did you eat?",
	"are you free later?",
	"what time works for you?",
	"how did the meeting go?",
	"have you heard back yet?",
	"want me to call?",
	"are you ok?",
	"shall I bring anything?",
	"did you see the news?",
}

var laughs = []string{
	"haha", "lol", "lmao", "hahaha", "ahah", "😂😂",
}

var closers = []string{
	"okay, sleep well x",
	"talk tomorrow ❤️",
	"night night",
	"catch you later",
	"gotta run, bye!",
	"speak soon x",
	"alright, ttyl",
}

var mediaLines = []string{
	"image omitted",
	"image omitted",
	"video omitted",
	"sticker omitted",
	"audio omitted",
	"GIF omitted",
}

var linkLines = []string{
	"check this https://example.com/article",
	"this is so us https://example.com/cats",
	"booked it https://example.com/restaurant",
	"saw this and thought of you https://example.org/post",
	"playlist for tonight https://example.com/playlist",
}

// Per-user signature lines. Each line carries one or more longer (>=5 char)
// nouns unique to that speaker so the topic extractor flags them as
// distinctive. Each signature word should land at least 8 times across the
// chat — with ~18% signature rate over ~4700 messages that's plenty.
var aliceSignatures = []string{
	"yoga class wiped me out today",
	"the bakery had fresh sourdough this morning",
	"watering the garden before it gets too hot",
	"finished another chapter of my novel",
	"pottery studio was packed tonight",
	"knitting another scarf, send help",
	"been journaling every morning this week",
	"painting the kitchen this weekend",
	"my succulent finally flowered!",
	"yoga teacher gave us a brutal sequence",
	"baking banana bread again",
	"reading at the bakery before work",
	"signed up for another pottery workshop",
	"garden is exploding with tomatoes",
	"started watercolour lessons on Tuesdays",
	"sewing a dress for the wedding",
	"the bakery espresso is unreal",
	"weekend pottery markets are my therapy",
}

var bobSignatures = []string{
	"football tonight, knees are wrecked",
	"new guitar pedal arrived in the post",
	"podcast episode dropped, listening now",
	"climbing gym was rammed this evening",
	"whisky tasting Friday, you in?",
	"watched the football with the lads",
	"jamming on guitar in the garage",
	"recording the podcast tomorrow morning",
	"climbing wall has new routes this week",
	"trying a peated whisky tonight",
	"football training was brutal",
	"guitar strings finally arrived",
	"podcast guest cancelled on me, ugh",
	"climbing chalk is everywhere in this house",
	"whisky shelf is getting out of hand",
	"watching the football match later",
	"new guitar tab I'm learning",
	"climbing trip to Wales next month?",
}

// Distinct top-5 emojis per user so the analyser's per-user emoji lists differ.
var aliceEmojis = []string{"😄", "☕", "❤️", "🙌", "✨"}
var bobEmojis = []string{"😂", "🔥", "👀", "🙃", "🎧"}
