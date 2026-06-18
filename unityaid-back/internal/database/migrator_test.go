package database

import (
	"strings"
	"testing"
)

func TestSplitSQLStatements(t *testing.T) {
	t.Parallel()

	script := `
INSERT INTO demo (name) VALUES ('first; value');
-- comment with ;
INSERT INTO demo (name) VALUES ('second');
DO $$
BEGIN
	RAISE NOTICE 'third; value';
END
$$;
`

	statements, err := splitSQLStatements(script)
	if err != nil {
		t.Fatalf("splitSQLStatements returned error: %v", err)
	}

	if len(statements) != 3 {
		t.Fatalf("expected 3 statements, got %d: %#v", len(statements), statements)
	}
	if !strings.Contains(statements[0], "'first; value'") {
		t.Fatalf("first statement was split inside string: %q", statements[0])
	}
	if !strings.Contains(statements[2], "RAISE NOTICE 'third; value'") {
		t.Fatalf("third statement was split inside dollar quote: %q", statements[2])
	}
}

func TestSplitSQLStatementsRejectsUnterminatedLiteral(t *testing.T) {
	t.Parallel()

	if _, err := splitSQLStatements("INSERT INTO demo VALUES ('broken);"); err == nil {
		t.Fatal("expected unterminated literal error")
	}
}
