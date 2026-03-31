package output

import (
	"testing"
)

// --- walkPath ---

func TestWalkPath_SimpleKey(t *testing.T) {
	obj := map[string]interface{}{"name": "Alice"}
	v, err := walkPath(obj, "name")
	if err != nil {
		t.Fatal(err)
	}
	if v != "Alice" {
		t.Errorf("expected 'Alice', got %v", v)
	}
}

func TestWalkPath_NestedKey(t *testing.T) {
	obj := map[string]interface{}{
		"user": map[string]interface{}{
			"id": float64(42),
		},
	}
	v, err := walkPath(obj, "user.id")
	if err != nil {
		t.Fatal(err)
	}
	if v != float64(42) {
		t.Errorf("expected 42, got %v", v)
	}
}

func TestWalkPath_ArrayIndex(t *testing.T) {
	obj := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}
	v, err := walkPath(obj, "items[1]")
	if err != nil {
		t.Fatal(err)
	}
	if v != "b" {
		t.Errorf("expected 'b', got %v", v)
	}
}

func TestWalkPath_NestedWithArray(t *testing.T) {
	obj := map[string]interface{}{
		"data": map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"name": "Bob"},
			},
		},
	}
	v, err := walkPath(obj, "data.users[0].name")
	if err != nil {
		t.Fatal(err)
	}
	if v != "Bob" {
		t.Errorf("expected 'Bob', got %v", v)
	}
}

func TestWalkPath_MissingKey(t *testing.T) {
	obj := map[string]interface{}{"name": "Alice"}
	_, err := walkPath(obj, "nothere")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestWalkPath_IndexOutOfRange(t *testing.T) {
	obj := map[string]interface{}{
		"items": []interface{}{"a"},
	}
	_, err := walkPath(obj, "items[5]")
	if err == nil {
		t.Error("expected error for out-of-range index")
	}
}

func TestWalkPath_EmptyPath(t *testing.T) {
	obj := map[string]interface{}{"x": 1}
	v, err := walkPath(obj, "")
	if err != nil {
		t.Fatal(err)
	}
	if v == nil {
		t.Error("expected non-nil for empty path (returns root)")
	}
}

// --- splitPath ---

func TestSplitPath_Simple(t *testing.T) {
	segs := splitPath("a.b.c")
	if len(segs) != 3 || segs[0] != "a" || segs[1] != "b" || segs[2] != "c" {
		t.Errorf("unexpected segments: %v", segs)
	}
}

func TestSplitPath_WithArrayIndex(t *testing.T) {
	segs := splitPath("data.items[0].id")
	if len(segs) != 3 || segs[1] != "items[0]" {
		t.Errorf("unexpected segments: %v", segs)
	}
}

func TestSplitPath_SingleSegment(t *testing.T) {
	segs := splitPath("name")
	if len(segs) != 1 || segs[0] != "name" {
		t.Errorf("unexpected segments: %v", segs)
	}
}

// --- parseSegment ---

func TestParseSegment_Plain(t *testing.T) {
	name, idx, has := parseSegment("data")
	if name != "data" || idx != 0 || has {
		t.Errorf("unexpected: name=%q idx=%d has=%v", name, idx, has)
	}
}

func TestParseSegment_WithIndex(t *testing.T) {
	name, idx, has := parseSegment("items[2]")
	if name != "items" || idx != 2 || !has {
		t.Errorf("unexpected: name=%q idx=%d has=%v", name, idx, has)
	}
}

func TestParseSegment_MissingClosingBracket(t *testing.T) {
	name, _, has := parseSegment("items[2")
	if has {
		t.Error("expected hasIdx=false for missing closing bracket")
	}
	if name != "items[2" {
		t.Errorf("expected raw segment returned, got %q", name)
	}
}

// --- FormatJSON ---

func TestFormatJSON_ValidJSON(t *testing.T) {
	raw := `{"a":1,"b":"two"}`
	out := FormatJSON(raw)
	if out == raw {
		t.Error("expected pretty-printed output, not same string")
	}
	if len(out) < len(raw) {
		t.Error("pretty-printed output should be longer than compact")
	}
}

func TestFormatJSON_InvalidJSON(t *testing.T) {
	raw := "not json at all"
	out := FormatJSON(raw)
	if out != raw {
		t.Errorf("expected raw string returned, got %q", out)
	}
}

func TestFormatJSON_EmptyString(t *testing.T) {
	out := FormatJSON("")
	if out != "" {
		t.Errorf("expected empty string, got %q", out)
	}
}

// --- formatSize ---

func TestFormatSize_Bytes(t *testing.T) {
	if s := formatSize(512); s != "512 B" {
		t.Errorf("expected '512 B', got %q", s)
	}
}

func TestFormatSize_Kilobytes(t *testing.T) {
	if s := formatSize(2048); s != "2.0 KB" {
		t.Errorf("expected '2.0 KB', got %q", s)
	}
}

func TestFormatSize_Megabytes(t *testing.T) {
	if s := formatSize(2 * 1024 * 1024); s != "2.0 MB" {
		t.Errorf("expected '2.0 MB', got %q", s)
	}
}
