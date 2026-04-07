package gendata

import (
	"bytes"
	"strings"
	"testing"

	"github.com/seanmcn/chinwag/internal/parser"
)

func TestGenerateRoundTripsThroughParser(t *testing.T) {
	var buf bytes.Buffer
	if err := Generate(&buf, Options{}); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	msgs, err := parser.Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(msgs) < 1500 {
		t.Fatalf("expected >=1500 messages, got %d", len(msgs))
	}

	authors := map[string]int{}
	for _, m := range msgs {
		authors[m.Author]++
	}
	if authors["Alice"] == 0 || authors["Bob"] == 0 {
		t.Fatalf("expected both Alice and Bob to appear, got %v", authors)
	}

	// Spot check: first line uses iOS bracket format.
	first := strings.SplitN(buf.String(), "\n", 2)[0]
	if !strings.HasPrefix(first, "[") {
		t.Fatalf("first line not iOS bracket format: %q", first)
	}
}

func TestGenerateDeterministic(t *testing.T) {
	var a, b bytes.Buffer
	if err := Generate(&a, Options{Seed: 7, Days: 30}); err != nil {
		t.Fatal(err)
	}
	if err := Generate(&b, Options{Seed: 7, Days: 30}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Fatalf("same seed produced different output")
	}
}
