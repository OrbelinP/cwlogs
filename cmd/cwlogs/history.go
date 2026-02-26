package cwlogs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TODO this should be configurable
const maxHistory = 20

type History struct {
	LogGroups []LogGroupDetails `json:"fetchedGroups"`
}

func historyPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting user config dir: %w", err)
	}

	appDir := filepath.Join(dir, "cwlogs")

	err = os.MkdirAll(appDir, 0700)
	if err != nil {
		return "", fmt.Errorf("creating app dir: %w", err)
	}

	return filepath.Join(appDir, "history.json"), nil
}

func LoadHistory() (History, error) {
	path, err := historyPath()
	if err != nil {
		return History{}, fmt.Errorf("getting history path: %w", err)
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return History{}, nil
		}

		return History{}, fmt.Errorf("reading history file: %w", err)
	}

	var res History
	err = json.Unmarshal(data, &res)
	if err != nil {
		return History{}, fmt.Errorf("parsing history file: %w", err)
	}

	return res, nil
}

func AddToHistory(detail LogGroupDetails) error {
	h, err := LoadHistory()
	if err != nil {
		return fmt.Errorf("loading history: %w", err)
	}

	for i, g := range h.LogGroups {
		if g.FullName == detail.FullName {
			h.LogGroups = append(h.LogGroups[:i], h.LogGroups[i+1:]...)
			break
		}
	}

	h.LogGroups = append([]LogGroupDetails{detail}, h.LogGroups...)

	if len(h.LogGroups) > maxHistory {
		h.LogGroups = h.LogGroups[:maxHistory]
	}

	path, err := historyPath()
	if err != nil {
		return fmt.Errorf("getting history path: %w", err)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("pretty-marshaling history file data: %w", err)
	}

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return fmt.Errorf("writing history file: %w", err)
	}

	return nil
}
