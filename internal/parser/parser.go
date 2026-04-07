// Package parser turns a WhatsApp chat export into a slice of Messages.
package parser

import (
	"archive/zip"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Kind int

const (
	KindText Kind = iota
	KindImage
	KindVideo
	KindAudio
	KindGIF
	KindSticker
	KindLink
	KindSystem
)

type Message struct {
	Timestamp time.Time
	Author    string
	Body      string
	Kind      Kind
}

// iOS:     [07/12/2013, 21:14:03] Bob: hello
// Android: 07/12/13, 21:14 - Bob: hello
var (
	reIOS     = regexp.MustCompile(`^\x{200e}?\[(\d{1,2}[\/\.\-]\d{1,2}[\/\.\-]\d{2,4}),\s+(\d{1,2}:\d{2}(?::\d{2})?)\s*([APap][Mm])?\]\s+([^:]+?):\s?(.*)$`)
	reAndroid = regexp.MustCompile(`^(\d{1,2}[\/\.\-]\d{1,2}[\/\.\-]\d{2,4}),\s+(\d{1,2}:\d{2}(?::\d{2})?)\s*([APap][Mm])?\s+-\s+([^:]+?):\s?(.*)$`)
	reURL     = regexp.MustCompile(`https?://\S+`)
)

// ParseFile reads a chat export from a path. Accepts .txt or .zip.
func ParseFile(path string) ([]Message, error) {
	if strings.EqualFold(filepath.Ext(path), ".zip") {
		return parseZip(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func parseZip(path string) ([]Message, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return parseZipFiles(r.File)
}

// ParseZipReader parses a WhatsApp .zip export from an in-memory reader.
func ParseZipReader(r io.ReaderAt, size int64) ([]Message, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	return parseZipFiles(zr.File)
}

func parseZipFiles(files []*zip.File) ([]Message, error) {
	for _, zf := range files {
		if strings.HasSuffix(zf.Name, ".txt") {
			rc, err := zf.Open()
			if err != nil {
				return nil, err
			}
			msgs, err := Parse(rc)
			rc.Close()
			return msgs, err
		}
	}
	return nil, os.ErrNotExist
}

// Parse reads from r and returns messages.
func Parse(r io.Reader) ([]Message, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1024*1024), 8*1024*1024)
	var out []Message
	var cur *Message
	flush := func() {
		if cur != nil {
			classify(cur)
			out = append(out, *cur)
			cur = nil
		}
	}
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		line = strings.TrimPrefix(line, "\u200e")
		if m := reIOS.FindStringSubmatch(line); m != nil {
			flush()
			ts, ok := parseTime(m[1], m[2], m[3])
			if !ok {
				continue
			}
			cur = &Message{Timestamp: ts, Author: strings.TrimSpace(m[4]), Body: m[5]}
		} else if m := reAndroid.FindStringSubmatch(line); m != nil {
			flush()
			ts, ok := parseTime(m[1], m[2], m[3])
			if !ok {
				continue
			}
			cur = &Message{Timestamp: ts, Author: strings.TrimSpace(m[4]), Body: m[5]}
		} else if cur != nil {
			cur.Body += "\n" + line
		}
	}
	flush()
	// Filter system messages (lines without author colon are still parsed but
	// some "system" notices come through with no author colon and were skipped).
	filtered := out[:0]
	for _, m := range out {
		if m.Author == "" {
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered, sc.Err()
}

func parseTime(date, clock, ampm string) (time.Time, bool) {
	date = strings.ReplaceAll(date, ".", "/")
	date = strings.ReplaceAll(date, "-", "/")
	dateLayouts := []string{"2/1/2006", "2/1/06", "1/2/2006", "1/2/06"}
	timeLayouts := []string{"15:04:05", "15:04", "3:04:05 PM", "3:04 PM"}
	clockFull := clock
	if ampm != "" {
		clockFull = clock + " " + strings.ToUpper(ampm)
	}
	for _, dl := range dateLayouts {
		for _, tl := range timeLayouts {
			if t, err := time.ParseInLocation(dl+" "+tl, date+" "+clockFull, time.Local); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func classify(m *Message) {
	body := strings.ToLower(strings.TrimSpace(m.Body))
	switch {
	case strings.Contains(body, "image omitted"), strings.HasSuffix(body, ".jpg (file attached)"), strings.HasSuffix(body, ".jpeg (file attached)"), strings.HasSuffix(body, ".png (file attached)"):
		m.Kind = KindImage
	case strings.Contains(body, "video omitted"), strings.HasSuffix(body, ".mp4 (file attached)"):
		m.Kind = KindVideo
	case strings.Contains(body, "audio omitted"), strings.Contains(body, "voice message"), strings.HasSuffix(body, ".opus (file attached)"), strings.HasSuffix(body, ".m4a (file attached)"):
		m.Kind = KindAudio
	case strings.Contains(body, "gif omitted"):
		m.Kind = KindGIF
	case strings.Contains(body, "sticker omitted"), strings.HasSuffix(body, ".webp (file attached)"):
		m.Kind = KindSticker
	case strings.Contains(body, "<media omitted>"):
		m.Kind = KindImage
	case reURL.MatchString(m.Body):
		m.Kind = KindLink
	default:
		m.Kind = KindText
	}
}
