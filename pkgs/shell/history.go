package shell

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/reeflective/readline"
)

var (
	ErrOpenHistoryFile = errors.New("failed to open history file")
	ErrNegativeIndex   = errors.New("cannot use a negative index when requesting historic commands")
	errOutOfRangeIndex = errors.New("index requested greater than number of items in history")
)

type fileHistory struct {
	file        string
	lines       []Item
	enableLocal bool
	maxLines    int
}

type Item struct {
	Index    int
	DateTime time.Time
	Block    string
}

// NewSourceFromFile returns a new history source writing to and reading from a file.
func EmbeddedHistory(file string, maxLines int, enableLocal ...bool) (readline.History, error) {
	var err error

	hist := new(fileHistory)
	hist.SetMaxLines(maxLines)
	if len(enableLocal) > 0 && enableLocal[0] {
		hist.EnableLocalFile()
	}
	hist.file = file
	hist.lines, err = openHist(file)

	return hist, err
}

func openHist(filename string) (list []Item, err error) {
	file, err := os.Open(filename)
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

func (h *fileHistory) EnableLocalFile() {
	h.enableLocal = true
}

func (h *fileHistory) SetMaxLines(maxLines int) {
	if h.maxLines > 0 {
		h.maxLines = maxLines
	} else {
		h.maxLines = 300
	}
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

	// store history items to local file.
	if h.enableLocal {
		var (
			content []byte
			err     error
		)
		// read history.
		if ok, _ := gutils.PathIsExist(h.file); ok {
			content, err = os.ReadFile(h.file)
			if err != nil {
				return h.Len(), err
			}
		}

		// save to file.
		if itemByte, err := json.Marshal(item); err != nil {
			return h.Len(), err
		} else {
			// strContent := string(content) + "\n" + string(itemByte)
			// os.WriteFile(h.file, []byte(strContent), 0666)
			h.writeLine(content, itemByte)
		}
	}
	return h.Len(), nil
}

func (h *fileHistory) writeLine(content []byte, itemByte []byte) {
	var result string
	if h.maxLines > 0 && strings.Count(string(content), "\n") >= (h.maxLines-1) {
		sList := strings.Split(string(content), "\n")
		sList = sList[1:]
		result = strings.Join(sList, "\n") + "\n" + string(itemByte)
	} else {
		result = string(content) + "\n" + string(itemByte)
	}
	if result != "" {
		os.WriteFile(h.file, []byte(result), 0666)
	}
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
