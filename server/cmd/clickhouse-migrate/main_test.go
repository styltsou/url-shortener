package main

import "testing"

func TestSplitStatements(t *testing.T) {
	sql := `
-- create analytics storage
CREATE DATABASE IF NOT EXISTS link4it;

CREATE TABLE link4it.schema_migrations (
  version String
) ENGINE = MergeTree()
ORDER BY version;
`
	statements := splitStatements(sql)
	if len(statements) != 2 {
		t.Fatalf("splitStatements() returned %d statements, want 2", len(statements))
	}
	if statements[0] != "CREATE DATABASE IF NOT EXISTS link4it;\n" {
		t.Fatalf("first statement = %q", statements[0])
	}
}
