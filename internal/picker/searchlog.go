package picker

import (
	"fmt"
	"os"
	"time"
)

// LogSearch appends a search event to the search log file.
// If selected is empty, the search was abandoned (user cancelled).
// Uses O_APPEND (not renameio) because this is an append-only log.
func LogSearch(path string, query string, selected string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening search log: %w", err)
	}
	defer f.Close()

	status := selected
	if status == "" {
		status = "(cancelled)"
	}
	_, err = fmt.Fprintf(f, "%s\t%s\t%s\n",
		time.Now().Format("2006-01-02T15:04:05.000000"),
		query,
		status)
	return err
}
