package parser

import (
	"testing"
)

func TestParseFile(t *testing.T) {
	msgs, err := ParseFile("../../testdata/sample_chat.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 12 {
		t.Fatalf("want 12 messages, got %d", len(msgs))
	}
	if msgs[0].Author != "Ben" || msgs[0].Body != "Hey! How are you?" {
		t.Errorf("first msg wrong: %+v", msgs[0])
	}
	var imgs, links int
	for _, m := range msgs {
		switch m.Kind {
		case KindImage:
			imgs++
		case KindLink:
			links++
		}
	}
	if imgs != 1 {
		t.Errorf("want 1 image, got %d", imgs)
	}
	if links != 1 {
		t.Errorf("want 1 link, got %d", links)
	}
}
