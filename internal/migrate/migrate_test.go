package migrate

import (
	"testing"
)

func TestSplitStatements(t *testing.T) {
	script := `-- a line comment
SET NAMES utf8mb4; -- trailing comment
SET FOREIGN_KEY_CHECKS = 0;

/* block
   comment */
CREATE TABLE t (
  id INT NOT NULL,
  name VARCHAR(20) DEFAULT 'a;b;c', -- semicolon inside string
  remark VARCHAR(10) DEFAULT "x;y"
);
INSERT INTO t (name) VALUES ('it''s;ok');
SET FOREIGN_KEY_CHECKS = 1;
`
	got := splitStatements(script)
	// 5 statements; semicolons inside string literals and comments must not split.
	if len(got) != 5 {
		t.Fatalf("expected 5 statements, got %d: %#v", len(got), got)
	}
	for _, want := range []string{"SET NAMES utf8mb4", "SET FOREIGN_KEY_CHECKS = 0", "SET FOREIGN_KEY_CHECKS = 1"} {
		if !contains(got, want) {
			t.Fatalf("missing %q in %#v", want, got)
		}
	}
	// The CREATE TABLE statement must keep its embedded semicolons intact.
	for _, s := range got {
		if len(s) > 12 && s[:12] == "CREATE TABLE" {
			if !stringsContains(s, "'a;b;c'") || !stringsContains(s, "\"x;y\"") {
				t.Fatalf("embedded semicolons lost: %q", s)
			}
		}
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestSplitStatementsDropsCommentOnly(t *testing.T) {
	got := splitStatements("-- only a comment\n   \n/* another */")
	if len(got) != 0 {
		t.Fatalf("expected no statements, got %#v", got)
	}
}

func TestSplitStatementsNoTrailingSemicolon(t *testing.T) {
	// A final statement without a trailing ';' is still captured.
	got := splitStatements("SELECT 1; SELECT 2")
	if len(got) != 2 {
		t.Fatalf("expected 2 statements, got %d: %#v", len(got), got)
	}
}
