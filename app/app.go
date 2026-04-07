package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/seanmcn/chinwag/internal/analyse"
	"github.com/seanmcn/chinwag/internal/parser"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application backend. It exposes a thin IPC surface
// over the existing parser + analyse packages.
type App struct {
	ctx context.Context
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// OpenFileDialog shows a native file picker and returns the chosen path,
// or an empty string if the user cancelled.
func (a *App) OpenFileDialog() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select WhatsApp chat export",
		Filters: []runtime.FileFilter{
			{DisplayName: "WhatsApp export (*.txt, *.zip)", Pattern: "*.txt;*.zip"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
}

// SaveJpeg shows a native save dialog seeded with defaultName and writes
// the given base64-encoded JPEG bytes to the chosen path. Returns the
// chosen path, or "" if the user cancelled.
func (a *App) SaveJpeg(defaultName string, b64 string) (string, error) {
	if i := strings.Index(b64, ","); i >= 0 && strings.HasPrefix(b64, "data:") {
		b64 = b64[i+1:]
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("decode jpeg: %w", err)
	}
	if !strings.HasSuffix(strings.ToLower(defaultName), ".jpg") &&
		!strings.HasSuffix(strings.ToLower(defaultName), ".jpeg") {
		defaultName += ".jpg"
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save export",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JPEG image (*.jpg)", Pattern: "*.jpg;*.jpeg"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// PickDirectory shows a native directory picker and returns the chosen
// path, or "" if the user cancelled.
func (a *App) PickDirectory(title string) (string, error) {
	if title == "" {
		title = "Choose folder"
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: title})
}

// SaveJpegTo writes the given base64-encoded JPEG to dir/filename without
// prompting. Returns the full path written.
func (a *App) SaveJpegTo(dir, filename, b64 string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("no directory")
	}
	if i := strings.Index(b64, ","); i >= 0 && strings.HasPrefix(b64, "data:") {
		b64 = b64[i+1:]
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("decode jpeg: %w", err)
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".jpg") &&
		!strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
		filename += ".jpg"
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// DistinctAuthors parses the file and returns the unique non-system author
// names in descending order of message count. Used to populate the
// "perspective" picker before running the full analysis.
func (a *App) DistinctAuthors(path string) ([]string, error) {
	msgs, err := parser.ParseFile(path)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, m := range msgs {
		if m.Kind == parser.KindSystem || m.Author == "" {
			continue
		}
		counts[m.Author]++
	}
	out := make([]string, 0, len(counts))
	for name := range counts {
		out = append(out, name)
	}
	sort.Slice(out, func(i, j int) bool { return counts[out[i]] > counts[out[j]] })
	return out, nil
}

// StatsDTO mirrors analyse.Stats but flattens the inline Period struct
// (which uses time.Time and an anonymous struct that the Wails binding
// generator can't introspect) into ISO-8601 strings the frontend can use.
type StatsDTO struct {
	Participants      [2]string                    `json:"Participants"`
	Period            PeriodDTO                    `json:"Period"`
	Messages          int                          `json:"Messages"`
	Conversations     int                          `json:"Conversations"`
	ChatPoints        int                          `json:"ChatPoints"`
	PerUser           map[string]*analyse.UserStats `json:"PerUser"`
	Timeline          []analyse.GrowthPoint        `json:"Timeline"`
	Heatmap           [7][24]int                   `json:"Heatmap"`
	DailyActivity     []analyse.DayCount           `json:"DailyActivity"`
	Convos            analyse.ConvoStats           `json:"Convos"`
	Insights          []analyse.Insight            `json:"Insights"`
	Rating            int                          `json:"Rating"`
	RatingLabel       string                       `json:"RatingLabel"`
	Balance           map[string]int               `json:"Balance"`
	BalancePct        map[string]int               `json:"BalancePct"`
	TopWeekdayHr      string                       `json:"TopWeekdayHr"`
	CharsTyped        int                          `json:"CharsTyped"`
	TimeTyping        string                       `json:"TimeTyping"`
	Direction         map[string]int               `json:"Direction"`
	LongestStreak     int                          `json:"LongestStreak"`
	CurrentStreak     int                          `json:"CurrentStreak"`
	TopTerms          []string                     `json:"TopTerms"`
	SentimentTimeline []analyse.SentimentPoint     `json:"SentimentTimeline"`
	TopDomains        []analyse.DomainCount        `json:"TopDomains"`
}

// PeriodDTO is a Wails-bindable replacement for analyse.Stats.Period.
type PeriodDTO struct {
	Start string `json:"Start"`
	End   string `json:"End"`
}

// Analyse parses the file and runs the full statistics computation.
// gapHours is the conversation-gap threshold in hours (defaults to 6 if <=0).
func (a *App) Analyse(path, me, them string, gapHours int) (StatsDTO, error) {
	msgs, err := parser.ParseFile(path)
	if err != nil {
		return StatsDTO{}, err
	}
	if len(msgs) == 0 {
		return StatsDTO{}, fmt.Errorf("no messages parsed — check the file format")
	}
	gap := time.Duration(gapHours) * time.Hour
	if gap <= 0 {
		gap = 6 * time.Hour
	}
	s := analyse.Run(msgs, me, them, gap)
	return StatsDTO{
		Participants:      s.Participants,
		Period:            PeriodDTO{Start: s.Period.Start.Format(time.RFC3339), End: s.Period.End.Format(time.RFC3339)},
		Messages:          s.Messages,
		Conversations:     s.Conversations,
		ChatPoints:        s.ChatPoints,
		PerUser:           s.PerUser,
		Timeline:          s.Timeline,
		Heatmap:           s.Heatmap,
		DailyActivity:     s.DailyActivity,
		Convos:            s.Convos,
		Insights:          s.Insights,
		Rating:            s.Rating,
		RatingLabel:       s.RatingLabel,
		Balance:           s.Balance,
		BalancePct:        s.BalancePct,
		TopWeekdayHr:      s.TopWeekdayHr,
		CharsTyped:        s.CharsTyped,
		TimeTyping:        s.TimeTyping,
		Direction:         s.Direction,
		LongestStreak:     s.LongestStreak,
		CurrentStreak:     s.CurrentStreak,
		TopTerms:          s.TopTerms,
		SentimentTimeline: s.SentimentTimeline,
		TopDomains:        s.TopDomains,
	}, nil
}

