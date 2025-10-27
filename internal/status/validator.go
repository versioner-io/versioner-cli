package status

import "strings"

// Canonical status values
const (
	Pending   = "pending"
	Started   = "started"
	Completed = "completed"
	Failed    = "failed"
	Aborted   = "aborted"
)

// statusAliases maps user input to canonical status values
var statusAliases = map[string]string{
	// Canonical values (pass through)
	"pending":   Pending,
	"started":   Started,
	"completed": Completed,
	"failed":    Failed,
	"aborted":   Aborted,

	// Aliases for pending
	"queued":    Pending,
	"scheduled": Pending,

	// Aliases for started
	"in_progress": Started,
	"init":        Started,
	"building":    Started,
	"deploying":   Started,

	// Aliases for completed
	"success":  Completed,
	"complete": Completed,
	"finished": Completed,
	"built":    Completed,
	"deployed": Completed,

	// Aliases for failed
	"fail":    Failed,
	"failure": Failed,
	"error":   Failed,

	// Aliases for aborted
	"abort":     Aborted,
	"cancelled": Aborted,
	"cancel":    Aborted,
	"skipped":   Aborted,
}

// Normalize converts a status value to its canonical form
// Returns the canonical status and a boolean indicating if normalization occurred
func Normalize(status string) (canonical string, wasNormalized bool) {
	normalized := strings.ToLower(strings.TrimSpace(status))

	if canonical, ok := statusAliases[normalized]; ok {
		return canonical, normalized != canonical
	}

	// Unknown status - return as-is (API will validate)
	return status, false
}

// IsValid checks if a status value is valid (canonical or alias)
func IsValid(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	_, ok := statusAliases[normalized]
	return ok
}

// GetCanonical returns the canonical form of a status value
func GetCanonical(status string) string {
	canonical, _ := Normalize(status)
	return canonical
}
