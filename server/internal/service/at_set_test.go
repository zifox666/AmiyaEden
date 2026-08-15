package service

import (
	"reflect"
	"testing"
)

func TestNormalizeTagListAndFittingDisplayName(t *testing.T) {
	tags, err := normalizeTagList([]string{" web ", "scram", "web"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tags, []string{"scram", "web"}) {
		t.Fatalf("unexpected tags: %#v", tags)
	}
	if got := fittingDisplayName("Sleipnir", tags, []string{"speed"}); got != "Sleipnir · scram+web+speed" {
		t.Fatalf("unexpected fitting name: %s", got)
	}
}

func TestNormalizeTagListRejectsEmptyTag(t *testing.T) {
	if _, err := normalizeTagList([]string{"web", " "}); err == nil {
		t.Fatal("expected empty tag to be rejected")
	}
}
