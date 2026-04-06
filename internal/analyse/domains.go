package analyse

import (
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/seanmcn/whatsapp-analyse/internal/parser"
)

var reURL = regexp.MustCompile(`https?://[^\s<>"']+`)

// computeTopDomains scans KindLink messages, extracts hosts, and returns the
// most-shared domains across the whole chat (top 10).
func computeTopDomains(msgs []parser.Message) []DomainCount {
	counts := map[string]int{}
	for _, m := range msgs {
		if m.Kind != parser.KindLink {
			continue
		}
		for _, raw := range reURL.FindAllString(m.Body, -1) {
			u, err := url.Parse(raw)
			if err != nil || u.Host == "" {
				continue
			}
			host := strings.ToLower(u.Host)
			host = strings.TrimPrefix(host, "www.")
			host = strings.TrimPrefix(host, "m.")
			counts[host]++
		}
	}
	out := make([]DomainCount, 0, len(counts))
	for d, c := range counts {
		out = append(out, DomainCount{Domain: d, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Domain < out[j].Domain
	})
	if len(out) > 10 {
		out = out[:10]
	}
	return out
}
