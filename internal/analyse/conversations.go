package analyse

import (
	"sort"
	"time"

	"github.com/seanmcn/chinwag/internal/parser"
)

type Conversation struct {
	Messages []parser.Message
}

func segment(msgs []parser.Message, gap time.Duration, me, them string) []Conversation {
	var out []Conversation
	var cur Conversation
	for _, m := range msgs {
		if m.Author != me && m.Author != them {
			continue
		}
		if len(cur.Messages) == 0 {
			cur.Messages = append(cur.Messages, m)
			continue
		}
		last := cur.Messages[len(cur.Messages)-1]
		if m.Timestamp.Sub(last.Timestamp) > gap {
			out = append(out, cur)
			cur = Conversation{Messages: []parser.Message{m}}
		} else {
			cur.Messages = append(cur.Messages, m)
		}
	}
	if len(cur.Messages) > 0 {
		out = append(out, cur)
	}
	return out
}

func computeResponses(convos []Conversation, per map[string]*UserStats) {
	type acc struct {
		first    []float64 // seconds
		all      []float64
		hourSum  [24]float64
		hourCnt  [24]int
	}
	accs := map[string]*acc{}
	for u := range per {
		accs[u] = &acc{}
	}
	for _, c := range convos {
		seenFirst := map[string]bool{}
		for i := 1; i < len(c.Messages); i++ {
			prev := c.Messages[i-1]
			cur := c.Messages[i]
			if cur.Author == prev.Author {
				continue
			}
			dt := cur.Timestamp.Sub(prev.Timestamp).Seconds()
			a := accs[cur.Author]
			if a == nil {
				continue
			}
			a.all = append(a.all, dt)
			h := cur.Timestamp.Hour()
			a.hourSum[h] += dt
			a.hourCnt[h]++
			if !seenFirst[cur.Author] {
				a.first = append(a.first, dt)
				seenFirst[cur.Author] = true
			}
		}
	}
	for u, a := range accs {
		if len(a.all) > 0 {
			var sum float64
			for _, v := range a.all {
				sum += v
			}
			per[u].AvgResponse = int64(sum / float64(len(a.all)))
			sorted := append([]float64(nil), a.all...)
			sort.Float64s(sorted)
			per[u].MedianResp = int64(sorted[len(sorted)/2])
			p90idx := (len(sorted) * 90) / 100
			if p90idx >= len(sorted) {
				p90idx = len(sorted) - 1
			}
			per[u].P90Response = int64(sorted[p90idx])
			var byHour [24]int64
			for h := 0; h < 24; h++ {
				if a.hourCnt[h] > 0 {
					byHour[h] = int64(a.hourSum[h] / float64(a.hourCnt[h]))
				}
			}
			per[u].ReplyByHour = byHour
		}
		if len(a.first) > 0 {
			var sum float64
			rapid := 0
			for _, v := range a.first {
				sum += v
				if v < 300 {
					rapid++
				}
			}
			per[u].AvgFirstResp = int64(sum / float64(len(a.first)))
			per[u].RapidFirstPct = rapid * 100 / len(a.first)
		}
	}
}
