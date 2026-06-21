package models

import "time"

// Progress tracks one user's checklist state for one pattern. The four
// booleans mirror the sheet's own checklist:
//   1. Understood       - "Topic understood conceptually"
//   2. Explained         - "Pattern explained out loud / on paper"
//   3. SolvedBlind        - "Solved in Notepad without help"
//   4. SolvedAfterGap     - "Solved again 3 days later from memory"
type Progress struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	PatternID       string    `json:"pattern_id"`
	Understood      bool      `json:"understood"`
	Explained       bool      `json:"explained"`
	SolvedBlind     bool      `json:"solved_blind"`
	SolvedAfterGap  bool      `json:"solved_after_gap"`
	Notes           string    `json:"notes"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// IsComplete reports whether every checklist step is done.
func (p Progress) IsComplete() bool {
	return p.Understood && p.Explained && p.SolvedBlind && p.SolvedAfterGap
}

// StepsDone returns how many of the 4 checklist steps are checked (0-4).
func (p Progress) StepsDone() int {
	count := 0
	if p.Understood {
		count++
	}
	if p.Explained {
		count++
	}
	if p.SolvedBlind {
		count++
	}
	if p.SolvedAfterGap {
		count++
	}
	return count
}

// PatternWithProgress is the shape returned by the API: a pattern merged
// with the requesting user's progress on it (or a zero-value Progress if
// they haven't started it yet).
type PatternWithProgress struct {
	Pattern
	Progress Progress `json:"progress"`
}

// TopicWithPatterns groups a topic with all its patterns + progress, which
// is exactly the shape the checklist UI renders per section.
type TopicWithPatterns struct {
	Topic
	Patterns []PatternWithProgress `json:"patterns"`
}

// Summary holds aggregate checklist stats for one user — total patterns
// vs. fully completed, total checklist steps vs. checked. Powers the
// progress bar on the checklist page and the stats shown on profiles.
type Summary struct {
	TotalPatterns     int `json:"total_patterns"`
	CompletedPatterns int `json:"completed_patterns"`
	TotalSteps        int `json:"total_steps"`
	CompletedSteps    int `json:"completed_steps"`
}
