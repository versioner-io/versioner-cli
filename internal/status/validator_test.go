package status

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input           string
		expectedOutput  string
		shouldNormalize bool
	}{
		// Canonical values (no normalization)
		{"pending", Pending, false},
		{"started", Started, false},
		{"completed", Completed, false},
		{"failed", Failed, false},
		{"aborted", Aborted, false},

		// Aliases for pending
		{"queued", Pending, true},
		{"scheduled", Pending, true},

		// Aliases for started
		{"in_progress", Started, true},
		{"init", Started, true},
		{"building", Started, true},
		{"deploying", Started, true},

		// Aliases for completed
		{"success", Completed, true},
		{"complete", Completed, true},
		{"finished", Completed, true},
		{"built", Completed, true},
		{"deployed", Completed, true},

		// Aliases for failed
		{"fail", Failed, true},
		{"failure", Failed, true},
		{"error", Failed, true},

		// Aliases for aborted
		{"abort", Aborted, true},
		{"cancelled", Aborted, true},
		{"cancel", Aborted, true},
		{"skipped", Aborted, true},

		// Case insensitive
		{"SUCCESS", Completed, true},
		{"PENDING", Pending, false},
		{"Failed", Failed, false},

		// With whitespace
		{" completed ", Completed, false},
		{" success ", Completed, true},

		// Unknown status
		{"unknown", "unknown", false},
		{"invalid", "invalid", false},
	}

	for _, test := range tests {
		canonical, wasNormalized := Normalize(test.input)

		if canonical != test.expectedOutput {
			t.Errorf("Normalize(%q) = %q, expected %q", test.input, canonical, test.expectedOutput)
		}

		if wasNormalized != test.shouldNormalize {
			t.Errorf("Normalize(%q) normalization flag = %v, expected %v", test.input, wasNormalized, test.shouldNormalize)
		}
	}
}

func TestIsValid(t *testing.T) {
	validStatuses := []string{
		"pending", "started", "completed", "failed", "aborted",
		"queued", "scheduled", "in_progress", "init", "building", "deploying",
		"success", "complete", "finished", "built", "deployed",
		"fail", "failure", "error",
		"abort", "cancelled", "cancel", "skipped",
		"SUCCESS", "PENDING", " completed ",
	}

	for _, status := range validStatuses {
		if !IsValid(status) {
			t.Errorf("IsValid(%q) = false, expected true", status)
		}
	}

	invalidStatuses := []string{
		"unknown", "invalid", "running", "done", "",
	}

	for _, status := range invalidStatuses {
		if IsValid(status) {
			t.Errorf("IsValid(%q) = true, expected false", status)
		}
	}
}

func TestGetCanonical(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"success", Completed},
		{"pending", Pending},
		{"in_progress", Started},
		{"failure", Failed},
		{"cancelled", Aborted},
		{"unknown", "unknown"},
	}

	for _, test := range tests {
		result := GetCanonical(test.input)
		if result != test.expected {
			t.Errorf("GetCanonical(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
