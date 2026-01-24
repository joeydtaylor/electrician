package builder

import "testing"

func TestHasSuffixFold(t *testing.T) {
	if !hasSuffixFold("demo/file.PARQUET", []string{".parquet"}) {
		t.Fatalf("expected suffix match")
	}
	if hasSuffixFold("demo/file.json", []string{".parquet"}) {
		t.Fatalf("unexpected suffix match")
	}
}
