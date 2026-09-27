package server

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSaveAndLoadState(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	want := State{PID: 42, Port: 9090}
	if err := saveState(want); err != nil {
		t.Fatalf("saveState() error = %v", err)
	}

	got, err := loadState()
	if err != nil {
		t.Fatalf("loadState() error = %v", err)
	}
	if got.PID != want.PID || got.Port != want.Port {
		t.Errorf("loadState() = %+v, want %+v", got, want)
	}
}

func TestClearState(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := saveState(State{PID: 1, Port: 8080}); err != nil {
		t.Fatal(err)
	}

	clearState()

	if _, err := loadState(); err == nil {
		t.Error("loadState() deve retornar erro após clearState()")
	}
}

func TestStateJSON(t *testing.T) {
	s := State{PID: 999, Port: 1234}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	var got State
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if got != s {
		t.Errorf("got %+v, want %+v", got, s)
	}
}

func TestBaseURL(t *testing.T) {
	got := BaseURL(8080)
	want := "http://localhost:8080"
	if got != want {
		t.Errorf("BaseURL(8080) = %q, want %q", got, want)
	}
}

func TestIsAlive_Offline(t *testing.T) {
	if IsAlive(19999) {
		t.Error("IsAlive(19999) deveria retornar false para porta sem servidor")
	}
}
