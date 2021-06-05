package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Handler struct {
	visitedStories map[string]string
}

func (h *Handler) Initialize(enableHistory bool) {
	if !enableHistory {
		h.visitedStories = nil
	}

	historyPath := getPathToHistory()

	if !exists(historyPath) {
		return
	}

	historyFileContent, err := os.ReadFile(historyPath)
	if err != nil {
		print("Error: could not read from file")

		os.Exit(1)
	}

	h.visitedStories = unmarshal(historyFileContent)
}

func getPathToHistory() string {
	homeDir, _ := os.UserHomeDir()
	configDir := ".cache"
	circumflexDir := "circumflex"
	historyFile := "history.json"

	return path.Join(homeDir, configDir, circumflexDir, historyFile)
}

func exists(pathToFile string) bool {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func unmarshal(input []byte) map[string]string {
	overrides := make(map[string]string)

	err := json.Unmarshal(input, &overrides)
	if err != nil {
		fmt.Printf("Error: %s\n", err)

		os.Exit(1)
	}

	return overrides
}
