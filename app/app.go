package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/seanmcn/whatsapp-analyse/internal/analyse"
	"github.com/seanmcn/whatsapp-analyse/internal/parser"
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

// Analyse parses the file and runs the full statistics computation.
// gapHours is the conversation-gap threshold in hours (defaults to 6 if <=0).
func (a *App) Analyse(path, me, them string, gapHours int) (analyse.Stats, error) {
	msgs, err := parser.ParseFile(path)
	if err != nil {
		return analyse.Stats{}, err
	}
	if len(msgs) == 0 {
		return analyse.Stats{}, fmt.Errorf("no messages parsed — check the file format")
	}
	gap := time.Duration(gapHours) * time.Hour
	if gap <= 0 {
		gap = 6 * time.Hour
	}
	return analyse.Run(msgs, me, them, gap), nil
}

