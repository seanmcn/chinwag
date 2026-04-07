package analyse

import "fmt"

func buildInsights(per map[string]*UserStats, me, them string, s Stats) []string {
	a, b := per[me], per[them]
	var out []string
	addCmp := func(left, right int, more, less, even string) {
		if left == 0 && right == 0 {
			return
		}
		ratio := float64(left+1) / float64(right+1)
		switch {
		case ratio > 1.3:
			out = append(out, more)
		case ratio < 0.77:
			out = append(out, less)
		default:
			if even != "" {
				out = append(out, even)
			}
		}
	}
	addCmp(a.ConvosStarted, b.ConvosStarted,
		"You start more conversations than your contact.",
		"They initiate far more conversations than you do.",
		"")
	addCmp(a.Laughs, b.Laughs, "You laugh more than your contact.", "Your contact laughs more than you.", "")
	addCmp(a.Apologies, b.Apologies, "You apologize more than your contact.", "Your contact apologizes more than you.", "")
	addCmp(a.Encouragement, b.Encouragement, "You send more encouragement than your contact.", "Your contact sends more encouragement than you.", "")
	addCmp(a.Questions, b.Questions, "You ask more questions than your contact.", "Your contact asks more questions than you.", "")
	if a.AvgResponse > 0 && b.AvgResponse > 0 {
		if a.AvgResponse*4 < b.AvgResponse*3 {
			out = append(out, "You respond much faster than your contact.")
		} else if b.AvgResponse*4 < a.AvgResponse*3 {
			out = append(out, "Your contact responds much faster than you.")
		}
	}
	addCmp(a.Messages, b.Messages, "You send slightly more messages than your contact.", "Your contact sends slightly more messages than you.", "Messages are evenly shared.")
	addCmp(a.ConvosClosed, b.ConvosClosed, "You close more conversations.", "Your contact closes more conversations.", "Conversation endings are evenly shared.")
	addCmp(a.ConvosMissed, b.ConvosMissed, "Your missed conversations skew toward you.", "Your missed conversations skew toward your contact.", "Your missed conversations are evenly matched.")
	addCmp(a.Reconnects, b.Reconnects, "You reconnect more often after long silences.", "You rarely reconnect after long silences.", "")

	// Sentiment
	addCmp(a.Positive, b.Positive, "You send more upbeat messages than your contact.", "Your contact sends more upbeat messages than you.", "")
	if a.Negative+b.Negative > 20 {
		addCmp(a.Negative, b.Negative, "You vent more often than your contact.", "Your contact vents more often than you.", "")
	}

	// Emotion mix (NRC). Indices: 0 anger, 1 anticipation, 2 disgust, 3 fear,
	// 4 joy, 5 sadness, 6 surprise, 7 trust, 8 negative, 9 positive.
	addCmp(a.Emotion[4], b.Emotion[4],
		"Your messages skew more toward joy than your contact's.",
		"Your contact's messages skew more toward joy than yours.", "")
	addCmp(a.Emotion[0], b.Emotion[0],
		"You express anger more often than your contact.",
		"Your contact expresses anger more often than you.", "")
	addCmp(a.Emotion[7], b.Emotion[7],
		"You express trust more often than your contact.",
		"Your contact expresses trust more often than you.", "")
	addCmp(a.Emotion[3], b.Emotion[3],
		"You voice fear or worry more than your contact.",
		"Your contact voices fear or worry more than you.", "")
	addCmp(a.Emotion[1], b.Emotion[1],
		"You look forward to things more than your contact.",
		"Your contact looks forward to things more than you.", "")

	// Joy vs anger ratio — internal tilt rather than person-vs-person.
	if a.Emotion[4]+a.Emotion[0] > 20 {
		ratio := float64(a.Emotion[4]+1) / float64(a.Emotion[0]+1)
		switch {
		case ratio > 2.0:
			out = append(out, "Your chats lean strongly toward joy over anger.")
		case ratio < 0.5:
			out = append(out, "Your chats lean toward anger over joy.")
		}
	}

	// Intensity peaks
	if a.IntensityPeak > b.IntensityPeak*1.3 {
		out = append(out, "You hit higher emotional peaks than your contact.")
	} else if b.IntensityPeak > a.IntensityPeak*1.3 {
		out = append(out, "Your contact hits higher emotional peaks than you.")
	}

	// VAD: arousal (energy) and dominance (control)
	if a.VADMessages > 5 && b.VADMessages > 5 {
		if a.VAD[1] > b.VAD[1]+0.05 {
			out = append(out, "You bring more energy to the chat than your contact.")
		} else if b.VAD[1] > a.VAD[1]+0.05 {
			out = append(out, "Your contact brings more energy to the chat than you.")
		}
		if a.VAD[2] > b.VAD[2]+0.05 {
			out = append(out, "You sound more in control than your contact.")
		} else if b.VAD[2] > a.VAD[2]+0.05 {
			out = append(out, "Your contact sounds more in control than you.")
		}
	}

	// VADER compound: overall tone
	if a.ScoredMsgs > 20 && b.ScoredMsgs > 20 {
		if a.CompoundAvg > b.CompoundAvg+0.1 {
			out = append(out, "Your overall tone reads warmer than your contact's.")
		} else if b.CompoundAvg > a.CompoundAvg+0.1 {
			out = append(out, "Your contact's overall tone reads warmer than yours.")
		}
	}

	// Latency spread (p90 vs avg) — flag bursty repliers
	if a.P90Response > 0 && a.AvgResponse > 0 && a.P90Response > 6*a.AvgResponse {
		out = append(out, "Your reply times are bursty — usually fast, occasionally very slow.")
	}
	if b.P90Response > 0 && b.AvgResponse > 0 && b.P90Response > 6*b.AvgResponse {
		out = append(out, "Your contact's reply times are bursty — usually fast, occasionally very slow.")
	}

	// Chronotype contrast
	if a.Chronotype != "" && b.Chronotype != "" && a.Chronotype != b.Chronotype {
		out = append(out, "You and your contact are on different rhythms — "+a.Chronotype+" vs "+b.Chronotype+".")
	}

	// Streaks
	if s.LongestStreak >= 14 {
		out = append(out, "Your longest unbroken daily streak ran for "+itoa(s.LongestStreak)+" days.")
	}
	if s.CurrentStreak >= 7 {
		out = append(out, "You're currently on a "+itoa(s.CurrentStreak)+"-day streak.")
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
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
