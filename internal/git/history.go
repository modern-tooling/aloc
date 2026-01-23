package git

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/modern-tooling/aloc/internal/model"
)

// ParseOptions controls git log parsing
type ParseOptions struct {
	SinceMonths int    // how far back to look
	Root        string // repository root
}

// ParseHistory runs git log and returns change events
func ParseHistory(opts ParseOptions) ([]ChangeEvent, error) {
	since := time.Now().AddDate(0, -opts.SinceMonths, 0).Format("2006-01-02")

	// single efficient git command
	cmd := exec.Command("git", "-C", opts.Root,
		"log",
		"--numstat",
		"--format=%H|%ae|%aI",
		"--since="+since,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseGitLog(string(out)), nil
}

// parseGitLog parses the git log output into change events
func parseGitLog(output string) []ChangeEvent {
	var events []ChangeEvent
	var currentAuthor string
	var currentTime time.Time

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// commit header line: hash|email|timestamp
		if strings.Contains(line, "|") && strings.Count(line, "|") == 2 {
			parts := strings.Split(line, "|")
			if len(parts) == 3 {
				currentAuthor = hashAuthor(parts[1])
				t, err := time.Parse(time.RFC3339, parts[2])
				if err == nil {
					currentTime = t
				}
			}
			continue
		}

		// numstat line: added\tdeleted\tpath
		fields := strings.Split(line, "\t")
		if len(fields) == 3 {
			// handle binary files (- - path)
			if fields[0] == "-" || fields[1] == "-" {
				continue
			}

			added, err1 := strconv.Atoi(fields[0])
			deleted, err2 := strconv.Atoi(fields[1])
			if err1 != nil || err2 != nil {
				continue
			}

			events = append(events, ChangeEvent{
				When:    currentTime,
				Path:    fields[2],
				Added:   added,
				Deleted: deleted,
				Author:  currentAuthor,
			})
		}
	}

	return events
}

// hashAuthor creates a privacy-preserving hash of an email
func hashAuthor(email string) string {
	h := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(h[:8]) // 16 char hex, anonymous
}

// MapRoles assigns roles to events based on file records
func MapRoles(events []ChangeEvent, records []*model.FileRecord) {
	roleMap := make(map[string]model.Role)
	for _, r := range records {
		roleMap[r.Path] = r.Role
	}

	for i := range events {
		if role, ok := roleMap[events[i].Path]; ok {
			events[i].Role = role
		}
	}
}

// BuildFileLOCMap creates a map of file paths to their current LOC
func BuildFileLOCMap(records []*model.FileRecord) map[string]int {
	m := make(map[string]int)
	for _, r := range records {
		m[r.Path] = r.LOC
	}
	return m
}
