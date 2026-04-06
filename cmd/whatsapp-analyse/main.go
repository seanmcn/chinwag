package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/seanmcn/whatsapp-analyse/internal/analyse"
	"github.com/seanmcn/whatsapp-analyse/internal/parser"
	"github.com/seanmcn/whatsapp-analyse/internal/server"
)

func main() {
	me := flag.String("me", "", "your name as it appears in the chat (default: top author)")
	them := flag.String("them", "", "the other person's name (default: 2nd author)")
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	gap := flag.Duration("gap", 6*time.Hour, "conversation gap threshold")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: whatsapp-analyse [flags] <chat.txt|chat.zip>")
		flag.PrintDefaults()
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
	fmt.Printf("parsed %d messages between %s and %s\n", stats.Messages, stats.Participants[0], stats.Participants[1])
	if err := server.Serve(*addr, stats); err != nil {
		log.Fatal(err)
	}
}
