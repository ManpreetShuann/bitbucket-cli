package output

import (
	"strings"
	"testing"
)

type testItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestJSONFormatter(t *testing.T) {
	f := &JSONFormatter{}
	items := []testItem{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}}
	out, err := f.Format(items)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("expected JSON with id:1, got: %s", out)
	}
}

func TestTemplateFormatter(t *testing.T) {
	f := &TemplateFormatter{Template: "{{range .}}{{.ID}}\t{{.Name}}\n{{end}}"}
	items := []testItem{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}}
	out, err := f.Format(items)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(out, "1\tfoo") {
		t.Errorf("expected template output, got: %s", out)
	}
}

func TestTableFormatter(t *testing.T) {
	cols := []Column{
		{Header: "ID", Width: 5},
		{Header: "NAME", Width: 10},
	}
	f := NewTableFormatter(cols, false)
	rows := [][]string{{"1", "foo"}, {"2", "bar"}}
	out := f.FormatRows(rows)
	if !strings.Contains(out, "ID") {
		t.Errorf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("expected data, got: %s", out)
	}
}

func TestNewFormatter(t *testing.T) {
	f := NewFormatter(true, "", false)
	if _, ok := f.(*JSONFormatter); !ok {
		t.Error("expected JSONFormatter when json=true")
	}

	f = NewFormatter(false, "{{.ID}}", false)
	if _, ok := f.(*TemplateFormatter); !ok {
		t.Error("expected TemplateFormatter when format is set")
	}
}
