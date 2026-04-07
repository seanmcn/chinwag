// Command gendata writes a deterministic, fictional WhatsApp-format chat to a
// file. It's a dev utility used to produce examples/sample_chat_large.txt for
// screenshots and demos.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/seanmcn/chinwag/internal/gendata"
)

func main() {
	seed := flag.Int64("seed", 42, "random seed")
	out := flag.String("out", "examples/sample_chat_large.txt", "output file path")
	days := flag.Int("days", 400, "number of days to generate")
	start := flag.String("start", "2024-01-15", "start date YYYY-MM-DD")
	me := flag.String("me", "Alice", "first author name")
	them := flag.String("them", "Bob", "second author name")
	flag.Parse()

	startT, err := time.Parse("2006-01-02", *start)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --start: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := gendata.Generate(f, gendata.Options{
		Seed:     *seed,
		Start:    startT,
		Days:     *days,
		MeName:   *me,
		ThemName: *them,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", *out)
}
