package services

import (
	"context"
	"fmt"

	"github.com/yourname/dsa-tracker/internal/models"
	"github.com/yourname/dsa-tracker/internal/repository"
)

// ChecklistService merges the static pattern catalog with a user's personal
// progress to produce the data the checklist UI renders.
type ChecklistService struct {
	catalog  repository.CatalogRepository
	progress repository.ProgressRepository
}

// NewChecklistService constructs a ChecklistService.
func NewChecklistService(catalog repository.CatalogRepository, progress repository.ProgressRepository) *ChecklistService {
	return &ChecklistService{catalog: catalog, progress: progress}
}

// GetFullChecklist returns every topic with its patterns, each annotated
// with the given user's progress (zero-value progress if untouched).
func (s *ChecklistService) GetFullChecklist(ctx context.Context, userID string) ([]models.TopicWithPatterns, error) {
	topics, err := s.catalog.ListTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing topics: %w", err)
	}

	allPatterns, err := s.catalog.ListAllPatterns(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing patterns: %w", err)
	}

	userProgress, err := s.progress.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing user progress: %w", err)
	}

	// Index progress rows by pattern ID for O(1) lookup while merging.
	progressByPattern := make(map[string]models.Progress, len(userProgress))
	for _, p := range userProgress {
		progressByPattern[p.PatternID] = p
	}

	// Group patterns by topic ID, preserving the sort order already applied
	// by the repository query.
	patternsByTopic := make(map[string][]models.PatternWithProgress)
	for _, pattern := range allPatterns {
		prog, ok := progressByPattern[pattern.ID]
		if !ok {
			// No row yet means the user hasn't touched this pattern —
			// represent that as an empty, unsaved progress state.
			prog = models.Progress{UserID: userID, PatternID: pattern.ID}
		}
		patternsByTopic[pattern.TopicID] = append(patternsByTopic[pattern.TopicID], models.PatternWithProgress{
			Pattern:  pattern,
			Progress: prog,
		})
	}

	result := make([]models.TopicWithPatterns, 0, len(topics))
	for _, topic := range topics {
		result = append(result, models.TopicWithPatterns{
			Topic:    topic,
			Patterns: patternsByTopic[topic.ID],
		})
	}

	return result, nil
}

// UpdateProgress saves the checklist state for one pattern for one user.
// It validates the pattern exists before writing, so a bad pattern_id from
// the client fails cleanly instead of silently creating an orphaned row.
func (s *ChecklistService) UpdateProgress(ctx context.Context, userID, patternID string, understood, explained, solvedBlind, solvedAfterGap bool, notes string) (*models.Progress, error) {
	if _, err := s.catalog.GetPattern(ctx, patternID); err != nil {
		return nil, fmt.Errorf("pattern not found: %w", err)
	}

	p := &models.Progress{
		UserID:         userID,
		PatternID:      patternID,
		Understood:     understood,
		Explained:      explained,
		SolvedBlind:    solvedBlind,
		SolvedAfterGap: solvedAfterGap,
		Notes:          notes,
	}

	if err := s.progress.Upsert(ctx, p); err != nil {
		return nil, fmt.Errorf("saving progress: %w", err)
	}

	return p, nil
}

// GetSummary returns aggregate stats (total patterns, fully completed count)
// for a user — powers a simple progress bar on the UI.
func (s *ChecklistService) GetSummary(ctx context.Context, userID string) (*models.Summary, error) {
	allPatterns, err := s.catalog.ListAllPatterns(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing patterns: %w", err)
	}

	userProgress, err := s.progress.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing user progress: %w", err)
	}

	progressByPattern := make(map[string]models.Progress, len(userProgress))
	for _, p := range userProgress {
		progressByPattern[p.PatternID] = p
	}

	summary := &models.Summary{
		TotalPatterns: len(allPatterns),
		TotalSteps:    len(allPatterns) * 4,
	}

	for _, pattern := range allPatterns {
		if p, ok := progressByPattern[pattern.ID]; ok {
			summary.CompletedSteps += p.StepsDone()
			if p.IsComplete() {
				summary.CompletedPatterns++
			}
		}
	}

	return summary, nil
}
