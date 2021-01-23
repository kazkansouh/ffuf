package input

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/ffuf/ffuf/pkg/ffuf"
)

type WordlistInput struct {
	config     *ffuf.Config
	data       [][]byte
	inject_len int
	done       int
	position   int
	keyword    string
	current    []byte
	next       <-chan []byte
	inject     chan<- [][]byte
	reset      chan<- struct{}
}

func NewWordlistInput(keyword string, value string, conf *ffuf.Config) (*WordlistInput, error) {
	var wl WordlistInput
	wl.keyword = keyword
	wl.config = conf
	wl.done = 0
	wl.position = 0
	wl.inject_len = 0
	var valid bool
	var err error
	// stdin?
	if value == "-" {
		// yes
		valid = true
	} else {
		// no
		valid, err = wl.validFile(value)
	}
	if err != nil {
		return &wl, err
	}
	if valid {
		err = wl.readFile(value)
	}
	if err == nil {
		next := make(chan []byte)
		inject := make(chan [][]byte)
		reset := make(chan struct{})
		wl.next, wl.inject, wl.reset = next, inject, reset
		go wl.worker(next, inject, reset)
	}
	return &wl, err
}

//Position will return the current position in the input list
func (w *WordlistInput) Position() int {
	return w.position
}

//ResetPosition resets the position back to beginning of the wordlist.
func (w *WordlistInput) ResetPosition() {
	w.position = 0
	w.done = 0
	w.current = nil
	w.inject_len = 0
	w.reset <- struct{}{}
}

//Keyword returns the keyword assigned to this InternalInputProvider
func (w *WordlistInput) Keyword() string {
	return w.keyword
}

//Next will increment the cursor position, and return a boolean telling if there's words left in the list
func (w *WordlistInput) Next() bool {
	if w.Position() >= w.Total() {
		return false
	}
	return true
}

// marshalls access to the injected wordlist and is responsible for picking next word
func (w *WordlistInput) worker(next chan<- []byte, inject <-chan [][]byte, reset <-chan struct{}) {
	var idx_data, idx_inject int = 0, 0
	var idx *int = nil
	var pending []byte
	var inject_data = [][]byte{}

	for {
		switch {
		case idx_inject < len(inject_data):
			pending = inject_data[idx_inject]
			idx = &idx_inject
		case idx_data < len(w.data):
			pending = w.data[idx_data]
			idx = &idx_data
		default:
			idx = nil
			pending = nil
		}
		select {
		case next <- pending:
			// case where idx = nil should be unreachable
			if idx != nil {
				*idx += 1
			}
		case values := <-inject:
			// nil can be injected by the Total function to help with synchronisation
			// might be better for it to have its own queue
			if values == nil {
				continue
			}
			inject_data = append(inject_data, values...)
			w.inject_len = len(inject_data)
		case <-reset:
			idx_data, idx_inject = 0, 0
			inject_data = [][]byte{}
		}
	}
}

//IncrementPosition will increment the current position in the inputprovider data slice
func (w *WordlistInput) IncrementPosition() {
	w.position += 1
}

//Value returns the value from wordlist at current cursor position
func (w *WordlistInput) Value() []byte {
	for w.done <= w.position {
		w.current = <-w.next
		w.done += 1
	}
	return w.current
}

//Total returns the size of wordlist
func (w *WordlistInput) Total() int {
	// ensure all pending injections have been processed
	w.inject <- nil
	return len(w.data) + w.inject_len
}

//validFile checks that the wordlist file exists and can be read
func (w *WordlistInput) validFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	f.Close()
	return true, nil
}

//readFile reads the file line by line to a byte slice
func (w *WordlistInput) readFile(path string) error {
	var file *os.File
	var err error
	if path == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(path)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	var data [][]byte
	var ok bool
	reader := bufio.NewScanner(file)
	re := regexp.MustCompile(`(?i)%ext%`)
	for reader.Scan() {
		if w.config.DirSearchCompat && len(w.config.Extensions) > 0 {
			text := []byte(reader.Text())
			if re.Match(text) {
				for _, ext := range w.config.Extensions {
					contnt := re.ReplaceAll(text, []byte(ext))
					data = append(data, []byte(contnt))
				}
			} else {
				text := reader.Text()

				if w.config.IgnoreWordlistComments {
					text, ok = stripComments(text)
					if !ok {
						continue
					}
				}
				data = append(data, []byte(text))
			}
		} else {
			text := reader.Text()

			if w.config.IgnoreWordlistComments {
				text, ok = stripComments(text)
				if !ok {
					continue
				}
			}
			data = append(data, []byte(text))
			if w.keyword == "FUZZ" && len(w.config.Extensions) > 0 {
				for _, ext := range w.config.Extensions {
					data = append(data, []byte(text+ext))
				}
			}
		}
	}
	w.data = data
	return reader.Err()
}

// stripComments removes all kind of comments from the word
func stripComments(text string) (string, bool) {
	// If the line starts with a # ignoring any space on the left,
	// return blank.
	if strings.HasPrefix(strings.TrimLeft(text, " "), "#") {
		return "", false
	}

	// If the line has # later after a space, that's a comment.
	// Only send the word upto space to the routine.
	index := strings.Index(text, " #")
	if index == -1 {
		return text, true
	}
	return text[:index], true
}

// Insert into the wordlist additional values (does not check for duplications)
func (w *WordlistInput) Inject(values [][]byte) {
	w.inject <- values
}
