package session

import "testing"

func TestManagerPutGetIsolation(t *testing.T) {
	m := NewManager()

	input := map[string]any{
		"name": "alpha",
	}
	m.Put("cli_1", ClientMeta{Client: input})
	input["name"] = "mutated"

	got, ok := m.Get("cli_1")
	if !ok {
		t.Fatal("expected session meta to exist")
	}
	if got.Client["name"] != "alpha" {
		t.Fatalf("expected stored value alpha, got %#v", got.Client["name"])
	}

	got.Client["name"] = "changed-by-caller"
	again, ok := m.Get("cli_1")
	if !ok {
		t.Fatal("expected session meta to exist on second read")
	}
	if again.Client["name"] != "alpha" {
		t.Fatalf("expected immutable stored value alpha, got %#v", again.Client["name"])
	}
}

func TestManagerNewClientIDUnique(t *testing.T) {
	m := NewManager()
	id1 := m.NewClientID()
	id2 := m.NewClientID()
	if id1 == id2 {
		t.Fatalf("expected unique ids, got same value %q", id1)
	}
}
