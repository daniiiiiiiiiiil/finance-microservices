package main

import "testing"

func TestExportContentType(t *testing.T) {
	cases := map[string]string{
		"report.csv":  "text/csv; charset=utf-8",
		"report.pdf":  "application/pdf",
		"report.json": "application/json; charset=utf-8",
		"report.txt":  "text/plain; charset=utf-8",
		"report.bin":  "application/octet-stream",
	}
	for filename, expected := range cases {
		if actual := exportContentType(filename); actual != expected {
			t.Fatalf("exportContentType(%q): expected %q, got %q", filename, expected, actual)
		}
	}
}
