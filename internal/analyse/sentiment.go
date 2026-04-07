package analyse

import "github.com/seanmcn/chinwag/internal/lexicon"

// scoreSentiment is a thin compatibility shim around ScoreVader. It
// returns +1 / 0 / -1 using VADER's standard compound cutoffs.
//
// New code should call ScoreMessage / ScoreVader directly to get the
// full compound score and per-class proportions.
func scoreSentiment(body string, lex *lexicon.Lexicons) int {
	return ScoreVader(body, lex).Label()
}
