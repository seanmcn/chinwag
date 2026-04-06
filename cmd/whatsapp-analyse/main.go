package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/seanmcn/whatsapp-analyse/internal/analyse"
	"github.com/seanmcn/whatsapp-analyse/internal/parser"
)

func main() {
	me := flag.String("me", "", "your name as it appears in the chat (default: top author)")
	them := flag.String("them", "", "the other person's name (default: 2nd author)")
	gap := flag.Duration("gap", 6*time.Hour, "conversation gap threshold")
	format := flag.String("format", "text", "output format: text or json")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] FILE\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "FILE is a WhatsApp chat export (.txt or .zip).")
		fmt.Fprintln(os.Stderr, "For the full graphical experience, use the desktop app instead.")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	msgs, err := parser.ParseFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	if len(msgs) == 0 {
		log.Fatal("no messages parsed — check the file format")
	}
	stats := analyse.Run(msgs, *me, *them, *gap)

	switch *format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(stats); err != nil {
			log.Fatal(err)
		}
	case "text":
		printText(stats)
	default:
		log.Fatalf("unknown --format %q (want text or json)", *format)
	}
}

func printText(s analyse.Stats) {
	a, b := s.Participants[0], s.Participants[1]
	fmt.Printf("WhatsApp Analyse — %s & %s\n", a, b)
	fmt.Printf("Period:        %s → %s\n", s.Period.Start.Format("2006-01-02"), s.Period.End.Format("2006-01-02"))
	fmt.Printf("Messages:      %d\n", s.Messages)
	fmt.Printf("Conversations: %d\n", s.Conversations)
	if s.RatingLabel != "" {
		fmt.Printf("Rating:        %d/100 (%s)\n", s.Rating, s.RatingLabel)
	}
	fmt.Println()
	for _, name := range []string{a, b} {
		u := s.PerUser[name]
		if u == nil {
			continue
		}
		fmt.Printf("── %s ──\n", name)
		fmt.Printf("  messages: %d   words: %d   chars: %d\n", u.Messages, u.Words, u.Characters)
		fmt.Printf("  emojis: %d   laughs: %d   questions: %d\n", u.Emojis, u.Laughs, u.Questions)
		fmt.Printf("  median reply: %ds   p90: %ds   chronotype: %s\n", u.MedianResp, u.P90Response, u.Chronotype)
		if len(u.TopEmojis) > 0 {
			fmt.Printf("  top emojis:")
			for i, e := range u.TopEmojis {
				if i >= 5 {
					break
				}
				fmt.Printf(" %s×%d", e.Emoji, e.Count)
			}
			fmt.Println()
		}
		if len(u.TopTerms) > 0 {
			fmt.Printf("  top terms: %v\n", u.TopTerms[:min(8, len(u.TopTerms))])
		}
		fmt.Println()
	}
	if len(s.Insights) > 0 {
		fmt.Println("Insights:")
		for _, ins := range s.Insights {
			fmt.Printf("  • %s\n", ins)
		}
	}
}
