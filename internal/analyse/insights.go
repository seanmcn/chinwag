package analyse

import "fmt"

func buildInsights(per map[string]*UserStats, me, them string) []string {
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
