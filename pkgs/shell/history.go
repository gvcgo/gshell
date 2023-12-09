package shell

import (
	"bufio"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/reeflective/readline"
)

var historyFile embed.FS

var (
	ErrOpenHistoryFile = errors.New("failed to open history file")
	ErrNegativeIndex   = errors.New("cannot use a negative index when requesting historic commands")
	errOutOfRangeIndex = errors.New("index requested greater than number of items in history")
)

type fileHistory struct {
	file  string
	lines []Item
}

type Item struct {
	Index    int
	DateTime time.Time
	Block    string
}

// NewSourceFromFile returns a new history source writing to and reading from a file.
func EmbeddedHistory(file string) (readline.History, error) {
	var err error

	hist := new(fileHistory)
	hist.file = file
	hist.lines, err = openHist(file)

	return hist, err
}

func openHist(filename string) (list []Item, err error) {
	file, err := historyFile.Open(filename)
	if err != nil {
		return list, fmt.Errorf("error opening history file: %s", err.Error())
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var item Item

		err := json.Unmarshal(scanner.Bytes(), &item)
		if err != nil || len(item.Block) == 0 {
			continue
		}

		item.Index = len(list)
		list = append(list, item)
	}

	file.Close()

	return list, nil
}

// Write item to history file.
func (h *fileHistory) Write(s string) (int, error) {
	block := strings.TrimSpace(s)
	if block == "" {
		return 0, nil
	}

	item := Item{
		DateTime: time.Now(),
		Block:    block,
		Index:    len(h.lines),
	}

	if len(h.lines) == 0 || h.lines[len(h.lines)-1].Block != block {
		h.lines = append(h.lines, item)
	}
	return h.Len(), nil
}

// GetLine returns a specific line from the history file.
func (h *fileHistory) GetLine(pos int) (string, error) {
	if pos < 0 {
		return "", ErrNegativeIndex
	}

	if pos < len(h.lines) {
		return h.lines[pos].Block, nil
	}

	return "", errOutOfRangeIndex
}

// Len returns the number of items in the history file.
func (h *fileHistory) Len() int {
	return len(h.lines)
}

// Dump returns the entire history file.
func (h *fileHistory) Dump() interface{} {
	return h.lines
}