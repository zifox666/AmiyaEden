package xlsx

import (
	"os"
	"testing"
)

func TestOpenATScoreWorkbook(t *testing.T) {
	path := os.Getenv("AT_SCORE_TEST_XLSX")
	if path == "" {
		t.Skip("AT_SCORE_TEST_XLSX not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	book, err := Open(data)
	if err != nil {
		t.Fatal(err)
	}
	if book.Sheets["New Static Values (Do Not Edit)"] != nil {
		if got := book.Cell("New Static Values (Do Not Edit)", "F4"); got != "Impairor" {
			t.Fatalf("New Static Values!F4 = %q", got)
		}
		if got := book.Cell("New Static Values (Do Not Edit)", "H4"); got != "1" {
			t.Fatalf("New Static Values!H4 = %q", got)
		}
		return
	}
	if got := book.Cell("Version History", "B2"); got != "0.7.0001" {
		t.Fatalf("version = %q", got)
	}
}
