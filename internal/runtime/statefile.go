package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type InstanceState struct {
	PID       int       `json:"pid"`
	Address   string    `json:"address"`
	Token     string    `json:"token"`
	StartedAt time.Time `json:"started_at"`
}

func StatePath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "mobilelab", ".state.json")
}

func ReadState(configPath string) (InstanceState, error) {
	path := StatePath(configPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return InstanceState{}, fmt.Errorf("MobileLab is not running (state file %s was not found)", path)
		}
		return InstanceState{}, fmt.Errorf("read runtime state: %w", err)
	}
	var state InstanceState
	if err := json.Unmarshal(data, &state); err != nil {
		return InstanceState{}, fmt.Errorf("parse runtime state: %w", err)
	}
	if state.Address == "" || state.Token == "" || state.PID < 1 {
		return InstanceState{}, fmt.Errorf("runtime state is incomplete; remove %s if no MobileLab instance is running", path)
	}
	return state, nil
}

func writeState(configPath string, state InstanceState) error {
	path := StatePath(configPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create runtime state directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode runtime state: %w", err)
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return fmt.Errorf("write runtime state: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("activate runtime state: %w", err)
	}
	return nil
}

func removeOwnedState(configPath, token string) error {
	state, err := ReadState(configPath)
	if err != nil {
		return nil
	}
	if state.Token != token {
		return nil
	}
	if err := os.Remove(StatePath(configPath)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove runtime state: %w", err)
	}
	return nil
}
