package analyse

import "fmt"

// Insight is a structured observation about the relationship that the
// frontend can render with category-aware styling. Tone identifies which
// participant the comparison favours: "me" / "them" / "even" / "info".
type Insight struct {
	Category string `json:"Category"`
	Icon     string `json:"Icon"`
	Title    string `json:"Title"`
	Detail   string `json:"Detail"`
	Tone     string `json:"Tone"`
}

const (
	catInitiative = "initiative"
	catBalance    = "balance"
	catTone       = "tone"
	catEnergy     = "energy"
	catRhythm     = "rhythm"
)

// ratioStr renders a multiplier like "1.6×" given two non-negative counts.
func ratioStr(hi, lo int) string {
	if lo <= 0 {
		lo = 1
	}
	r := float64(hi+1) / float64(lo+1)
	return fmt.Sprintf("%.1f×", r)
}

func buildInsights(per map[string]*UserStats, me, them string, s Stats) []Insight {
	a, b := per[me], per[them]
	var out []Insight
	add := func(in Insight) { out = append(out, in) }

	// cmp emits one insight from a left/right comparison. The variant texts
	// describe the "me does more", "them does more", and "even" cases.
	cmp := func(left, right int, cat, icon,
		moreT, moreD, lessT, lessD, evenT, evenD string) {
		if left == 0 && right == 0 {
			return
		}
		ratio := float64(left+1) / float64(right+1)
		switch {
		case ratio > 1.3:
			add(Insight{cat, icon, moreT, moreD, "me"})
		case ratio < 0.77:
			add(Insight{cat, icon, lessT, lessD, "them"})
		default:
			if evenT != "" {
				add(Insight{cat, icon, evenT, evenD, "even"})
			}
		}
	}

	// ── Initiative ──────────────────────────────────────────────
	cmp(a.ConvosStarted, b.ConvosStarted, catInitiative, "🚀",
		"You start the chats", fmt.Sprintf("You open conversations %s more often than %s.", ratioStr(a.ConvosStarted, b.ConvosStarted), them),
		"They start the chats", fmt.Sprintf("%s opens conversations %s more often than you.", them, ratioStr(b.ConvosStarted, a.ConvosStarted)),
		"", "")
	cmp(a.Questions, b.Questions, catInitiative, "❓",
		"You ask more questions", fmt.Sprintf("You ask %s as many questions as %s.", ratioStr(a.Questions, b.Questions), them),
		"They ask more questions", fmt.Sprintf("%s asks %s as many questions as you.", them, ratioStr(b.Questions, a.Questions)),
		"", "")
	cmp(a.ConvosClosed, b.ConvosClosed, catInitiative, "👋",
		"You close more conversations", fmt.Sprintf("You wrap up %s more chats than %s.", ratioStr(a.ConvosClosed, b.ConvosClosed), them),
		"They close more conversations", fmt.Sprintf("%s wraps up %s more chats than you.", them, ratioStr(b.ConvosClosed, a.ConvosClosed)),
		"Conversation endings are even", fmt.Sprintf("You and %s close out chats about as often.", them))
	cmp(a.Reconnects, b.Reconnects, catInitiative, "🔁",
		"You reconnect after silences", fmt.Sprintf("You restart things %s more often after long gaps.", ratioStr(a.Reconnects, b.Reconnects)),
		"They reconnect after silences", fmt.Sprintf("%s usually breaks long silences first.", them),
		"", "")

	// ── Balance ─────────────────────────────────────────────────
	cmp(a.Messages, b.Messages, catBalance, "⚖️",
		"You send more messages", fmt.Sprintf("You send %s as many messages as %s.", ratioStr(a.Messages, b.Messages), them),
		"They send more messages", fmt.Sprintf("%s sends %s as many messages as you.", them, ratioStr(b.Messages, a.Messages)),
		"Messages are evenly shared", fmt.Sprintf("You and %s contribute about the same volume.", them))
	cmp(a.ConvosMissed, b.ConvosMissed, catBalance, "📭",
		"Missed chats skew to you", fmt.Sprintf("More unanswered openings land on your side."),
		"Missed chats skew to them", fmt.Sprintf("More unanswered openings land on %s's side.", them),
		"Missed conversations are even", "Unanswered openings are evenly matched.")
	cmp(a.Apologies, b.Apologies, catBalance, "🙏",
		"You apologise more", fmt.Sprintf("You say sorry %s as often as %s.", ratioStr(a.Apologies, b.Apologies), them),
		"They apologise more", fmt.Sprintf("%s says sorry %s as often as you.", them, ratioStr(b.Apologies, a.Apologies)),
		"", "")
	cmp(a.Encouragement, b.Encouragement, catBalance, "💪",
		"You cheer them on more", fmt.Sprintf("You send encouragement %s more often.", ratioStr(a.Encouragement, b.Encouragement)),
		"They cheer you on more", fmt.Sprintf("%s sends encouragement %s more often than you.", them, ratioStr(b.Encouragement, a.Encouragement)),
		"", "")
	cmp(a.Laughs, b.Laughs, catBalance, "😂",
		"You laugh more", fmt.Sprintf("You laugh %s as much as %s.", ratioStr(a.Laughs, b.Laughs), them),
		"They laugh more", fmt.Sprintf("%s laughs %s as much as you.", them, ratioStr(b.Laughs, a.Laughs)),
		"", "")

	// ── Tone ────────────────────────────────────────────────────
	cmp(a.Positive, b.Positive, catTone, "🌞",
		"You send more upbeat notes", fmt.Sprintf("Your messages skew positive %s more often.", ratioStr(a.Positive, b.Positive)),
		"They send more upbeat notes", fmt.Sprintf("%s's messages skew positive %s more often.", them, ratioStr(b.Positive, a.Positive)),
		"", "")
	if a.Negative+b.Negative > 20 {
		cmp(a.Negative, b.Negative, catTone, "🌧️",
			"You vent more", fmt.Sprintf("You voice frustration %s more often than %s.", ratioStr(a.Negative, b.Negative), them),
			"They vent more", fmt.Sprintf("%s voices frustration %s more often than you.", them, ratioStr(b.Negative, a.Negative)),
			"", "")
	}
	// NRC emotion mix. Indices: 0 anger, 1 anticipation, 3 fear, 4 joy, 7 trust.
	cmp(a.Emotion[4], b.Emotion[4], catTone, "😄",
		"Your words lean joyful", fmt.Sprintf("You use joyful language %s more than %s.", ratioStr(a.Emotion[4], b.Emotion[4]), them),
		"Their words lean joyful", fmt.Sprintf("%s uses joyful language %s more than you.", them, ratioStr(b.Emotion[4], a.Emotion[4])),
		"", "")
	cmp(a.Emotion[0], b.Emotion[0], catTone, "😠",
		"You express more anger", fmt.Sprintf("Anger surfaces %s more often in your messages.", ratioStr(a.Emotion[0], b.Emotion[0])),
		"They express more anger", fmt.Sprintf("Anger surfaces %s more often in %s's messages.", ratioStr(b.Emotion[0], a.Emotion[0]), them),
		"", "")
	cmp(a.Emotion[7], b.Emotion[7], catTone, "🤝",
		"You express more trust", fmt.Sprintf("Trust language is %s more common from you.", ratioStr(a.Emotion[7], b.Emotion[7])),
		"They express more trust", fmt.Sprintf("Trust language is %s more common from %s.", ratioStr(b.Emotion[7], a.Emotion[7]), them),
		"", "")
	cmp(a.Emotion[3], b.Emotion[3], catTone, "😟",
		"You voice more worry", fmt.Sprintf("Fear or worry comes through %s more from you.", ratioStr(a.Emotion[3], b.Emotion[3])),
		"They voice more worry", fmt.Sprintf("Fear or worry comes through %s more from %s.", ratioStr(b.Emotion[3], a.Emotion[3]), them),
		"", "")
	cmp(a.Emotion[1], b.Emotion[1], catTone, "✨",
		"You look forward more", fmt.Sprintf("Anticipation language is %s more common from you.", ratioStr(a.Emotion[1], b.Emotion[1])),
		"They look forward more", fmt.Sprintf("Anticipation language is %s more common from %s.", ratioStr(b.Emotion[1], a.Emotion[1]), them),
		"", "")

	// Joy vs anger tilt — describes the chat as a whole.
	if a.Emotion[4]+a.Emotion[0] > 20 {
		ratio := float64(a.Emotion[4]+1) / float64(a.Emotion[0]+1)
		switch {
		case ratio > 2.0:
			add(Insight{catTone, "🌈", "Joy outweighs anger", "Your chats lean strongly toward joy over anger.", "info"})
		case ratio < 0.5:
			add(Insight{catTone, "⛈️", "Anger outweighs joy", "Your chats lean toward anger over joy.", "info"})
		}
	}

	// VADER compound: overall warmth.
	if a.ScoredMsgs > 20 && b.ScoredMsgs > 20 {
		if a.CompoundAvg > b.CompoundAvg+0.1 {
			add(Insight{catTone, "💗", "Your tone is warmer", fmt.Sprintf("Your overall tone reads warmer than %s's.", them), "me"})
		} else if b.CompoundAvg > a.CompoundAvg+0.1 {
			add(Insight{catTone, "💗", "Their tone is warmer", fmt.Sprintf("%s's overall tone reads warmer than yours.", them), "them"})
		}
	}

	// ── Energy ──────────────────────────────────────────────────
	if a.IntensityPeak > b.IntensityPeak*1.3 {
		add(Insight{catEnergy, "🔥", "You hit higher peaks", fmt.Sprintf("Your emotional peaks run hotter than %s's.", them), "me"})
	} else if b.IntensityPeak > a.IntensityPeak*1.3 {
		add(Insight{catEnergy, "🔥", "They hit higher peaks", fmt.Sprintf("%s's emotional peaks run hotter than yours.", them), "them"})
	}
	if a.VADMessages > 5 && b.VADMessages > 5 {
		if a.VAD[1] > b.VAD[1]+0.05 {
			add(Insight{catEnergy, "⚡", "You bring more energy", fmt.Sprintf("Your messages carry more arousal than %s's.", them), "me"})
		} else if b.VAD[1] > a.VAD[1]+0.05 {
			add(Insight{catEnergy, "⚡", "They bring more energy", fmt.Sprintf("%s's messages carry more arousal than yours.", them), "them"})
		}
		if a.VAD[2] > b.VAD[2]+0.05 {
			add(Insight{catEnergy, "🎯", "You sound more in control", fmt.Sprintf("You come across as more assertive than %s.", them), "me"})
		} else if b.VAD[2] > a.VAD[2]+0.05 {
			add(Insight{catEnergy, "🎯", "They sound more in control", fmt.Sprintf("%s comes across as more assertive than you.", them), "them"})
		}
	}

	// ── Rhythm ──────────────────────────────────────────────────
	if a.AvgResponse > 0 && b.AvgResponse > 0 {
		if a.AvgResponse*4 < b.AvgResponse*3 {
			add(Insight{catRhythm, "⏱️", "You reply faster", fmt.Sprintf("Your average reply lands well before %s's.", them), "me"})
		} else if b.AvgResponse*4 < a.AvgResponse*3 {
			add(Insight{catRhythm, "⏱️", "They reply faster", fmt.Sprintf("%s's average reply lands well before yours.", them), "them"})
		}
	}
	if a.P90Response > 0 && a.AvgResponse > 0 && a.P90Response > 6*a.AvgResponse {
		add(Insight{catRhythm, "📈", "Your replies are bursty", "Usually fast, occasionally very slow.", "info"})
	}
	if b.P90Response > 0 && b.AvgResponse > 0 && b.P90Response > 6*b.AvgResponse {
		add(Insight{catRhythm, "📈", "Their replies are bursty", fmt.Sprintf("%s is usually fast, but occasionally very slow.", them), "info"})
	}
	if a.Chronotype != "" && b.Chronotype != "" && a.Chronotype != b.Chronotype {
		add(Insight{catRhythm, "🌗", "Different rhythms", fmt.Sprintf("You're %s; %s is %s.", a.Chronotype, them, b.Chronotype), "info"})
	}
	if s.LongestStreak >= 14 {
		add(Insight{catRhythm, "🔥", "Longest streak", fmt.Sprintf("Your longest unbroken daily streak ran for %d days.", s.LongestStreak), "info"})
	}
	if s.CurrentStreak >= 7 {
		add(Insight{catRhythm, "📅", "On a streak", fmt.Sprintf("You're currently on a %d-day streak.", s.CurrentStreak), "info"})
	}
	return out
}

func computeRating(s Stats) (int, string) {
	// Score components 0-100
	balance := 100 - abs(50-s.BalancePct[s.Participants[0]])*2
	if balance < 0 {
		balance = 0
	}
	// Response speed: < 1 hour = 100, > 1 day = 0
	avg := (s.PerUser[s.Participants[0]].AvgResponse + s.PerUser[s.Participants[1]].AvgResponse) / 2
	speed := 100 - int(avg/864) // seconds to 0..100 over 24h
	if speed < 0 {
		speed = 0
	}
	if speed > 100 {
		speed = 100
	}
	// Reciprocity (both contribute messages)
	rec := 0
	if s.Messages > 0 {
		rec = 100 - abs(s.PerUser[s.Participants[0]].Messages-s.PerUser[s.Participants[1]].Messages)*100/s.Messages
	}
	score := (balance*40 + speed*30 + rec*30) / 100
	label := "A great relationship"
	switch {
	case score < 40:
		label = "Needs some love"
	case score < 60:
		label = "A solid connection"
	case score < 80:
		label = "A strong relationship"
	}
	return score, fmt.Sprintf("%s", label)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
