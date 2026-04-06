// Package render builds the HTML report from Stats.
package render

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/seanmcn/whatsapp-analyse/internal/analyse"
)

//go:embed templates/*.tmpl static/*
var assets embed.FS

// StaticFS exposes the embedded /static dir for the http server.
func StaticFS() fs.FS {
	sub, _ := fs.Sub(assets, "static")
	return sub
}

type viewData struct {
	Me, Them                          string
	S                                 analyse.Stats
	JSON                              template.JS
	PeriodStart, PeriodEnd, PeriodLen string
	RingOffset                        float64
	BalanceTag                        string
}

// HTML renders the report HTML for the given stats.
func HTML(s analyse.Stats) ([]byte, error) {
	funcs := template.FuncMap{
		"fmtSec": func(sec int64) string {
			d := time.Duration(sec) * time.Second
			if d >= time.Hour {
				return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
			}
			return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
		},
		"comma":   commaInt,
		"commaF":  func(f float64) string { return commaInt(int(f + 0.5)) },
		"initial": func(s string) string {
			for _, r := range s {
				if unicode.IsLetter(r) {
					return strings.ToUpper(string(r))
				}
			}
			return "?"
		},
		"cmpA": func(a, b int) string { return cmpClass(a, b) },
		"cmpB": func(a, b int) string { return cmpClass(b, a) },
		"cmpAS": func(a, b int64) string {
			// for response time: lower is better, so flip
			return cmpClass(int(b), int(a))
		},
		"cmpBS": func(a, b int64) string {
			return cmpClass(int(a), int(b))
		},
	}
	tmpl, err := template.New("report.html.tmpl").Funcs(funcs).ParseFS(assets, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	jsonBlob, err := json.Marshal(map[string]any{
		"me":       s.Participants[0],
		"them":     s.Participants[1],
		"timeline": s.Timeline,
		"heatmap":  s.Heatmap,
		"sankey":   s.Convos.Sankey,
		"daily":    s.DailyActivity,
	})
	if err != nil {
		return nil, err
	}

	d := viewData{
		Me:          s.Participants[0],
		Them:        s.Participants[1],
		S:           s,
		JSON:        template.JS(jsonBlob),
		PeriodStart: s.Period.Start.Format("2 Jan 2006"),
		PeriodEnd:   s.Period.End.Format("2 Jan 2006"),
		PeriodLen:   periodLen(s.Period.Start, s.Period.End),
		RingOffset:  326.7 * (1 - float64(s.Rating)/100),
		BalanceTag:  balanceTag(s),
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func commaInt(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.Itoa(n)
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func balanceTag(s analyse.Stats) string {
	p := s.BalancePct[s.Participants[0]]
	d := p - 50
	if d < 0 {
		d = -d
	}
	switch {
	case d <= 2:
		return "Contributions are perfectly balanced"
	case d <= 8:
		return "Contributions are well balanced"
	case d <= 18:
		return "Contributions lean to one side"
	default:
		return "Contributions are heavily skewed"
	}
}

// cmpClass returns "win" if a > b, "lose" if a < b, else "".
func cmpClass(a, b int) string {
	switch {
	case a > b:
		return "win"
	case a < b:
		return "lose"
	}
	return ""
}

func periodLen(a, b time.Time) string {
	if a.IsZero() || b.IsZero() {
		return "—"
	}
	years := b.Year() - a.Year()
	months := int(b.Month()) - int(a.Month())
	days := b.Day() - a.Day()
	if days < 0 {
		months--
		days += 30
	}
	if months < 0 {
		years--
		months += 12
	}
	return fmt.Sprintf("%dy %dm %dd", years, months, days)
}
