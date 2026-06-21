package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/yourname/dsa-tracker/internal/models"
	"github.com/yourname/dsa-tracker/internal/repository"
	"github.com/yourname/dsa-tracker/internal/utils"
)

// Sentinel errors for profile updates, mapped to HTTP statuses by the handler.
var (
	ErrUsernameTaken   = errors.New("username is already taken")
	ErrUsernameInvalid = errors.New("username must be 3-30 characters: lowercase letters, numbers, and dashes only")
	ErrProfileNotFound = errors.New("profile not found")
)

// ProfileService handles reading and updating user profiles, including the
// public, shareable /u/<username> view.
type ProfileService struct {
	users     repository.UserRepository
	colleges  repository.CollegeRepository
	checklist *ChecklistService
}

// NewProfileService constructs a ProfileService.
func NewProfileService(users repository.UserRepository, colleges repository.CollegeRepository, checklist *ChecklistService) *ProfileService {
	return &ProfileService{users: users, colleges: colleges, checklist: checklist}
}

// UpdateProfileInput holds the editable profile fields. Username is a
// pointer so the handler can distinguish "not provided, keep current
// value" from "explicitly set to this value" on a partial update.
type UpdateProfileInput struct {
	Username    *string
	LeetCodeURL *string
	IsAnonymous *bool
}

// UpdateProfile applies a partial update to the given user's mutable
// profile fields and persists it. Validates the new username (format +
// uniqueness) if one is provided.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*models.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("loading user: %w", err)
	}

	if input.Username != nil {
		newUsername := *input.Username
		if !utils.IsValidUsername(newUsername) {
			return nil, ErrUsernameInvalid
		}
		if newUsername != user.Username {
			taken, err := s.users.UsernameTaken(ctx, newUsername, user.ID)
			if err != nil {
				return nil, fmt.Errorf("checking username availability: %w", err)
			}
			if taken {
				return nil, ErrUsernameTaken
			}
			user.Username = newUsername
		}
	}

	if input.LeetCodeURL != nil {
		user.LeetCodeURL = *input.LeetCodeURL
	}

	if input.IsAnonymous != nil {
		user.IsAnonymous = *input.IsAnonymous
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("saving profile: %w", err)
	}

	return user, nil
}

// GetMyProfile returns the caller's own profile, fully unmasked regardless
// of their is_anonymous setting — this powers the edit-profile screen,
// where a person must always be able to see their own real info even if
// they've chosen to hide it from everyone else. GetPublicProfile, by
// contrast, applies anonymity masking unconditionally and is for viewers
// other than the owner.
func (s *ProfileService) GetMyProfile(ctx context.Context, userID string) (*models.PublicProfile, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("loading user: %w", err)
	}

	summary, err := s.checklist.GetSummary(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("loading progress summary: %w", err)
	}

	profile := &models.PublicProfile{
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		LeetCodeURL: user.LeetCodeURL,
		IsAnonymous: user.IsAnonymous,
		JoinedAt:    user.CreatedAt,
		Summary:     *summary,
	}

	if user.CollegeID != nil {
		college, err := s.colleges.GetByID(ctx, *user.CollegeID)
		if err == nil {
			profile.CollegeName = college.Name
			profile.CollegeStatus = string(college.Status)
		}
	}

	return profile, nil
}

// GetPublicProfile builds the /u/<username> view. If the profile owner has
// is_anonymous set, identity fields (display name/avatar/college/leetcode)
// are blanked — stats remain visible either way, matching the product
// decision used for the leaderboard.
//
// Note: the *requested* username always resolves the right user even when
// anonymous — anonymity hides identity *details* on the page, it doesn't
// make the page itself unreachable. That mirrors how the leaderboard keeps
// the row but hides the name.
func (s *ProfileService) GetPublicProfile(ctx context.Context, username string) (*models.PublicProfile, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("loading user: %w", err)
	}

	summary, err := s.checklist.GetSummary(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("loading progress summary: %w", err)
	}

	profile := &models.PublicProfile{
		Username:    user.Username,
		IsAnonymous: user.IsAnonymous,
		JoinedAt:    user.CreatedAt,
		Summary:     *summary,
	}

	if user.IsAnonymous {
		// Identity stays hidden; stats and join date are still shown.
		profile.DisplayName = "Anonymous"
		return profile, nil
	}

	profile.DisplayName = user.DisplayName
	profile.AvatarURL = user.AvatarURL
	profile.LeetCodeURL = user.LeetCodeURL

	if user.CollegeID != nil {
		college, err := s.colleges.GetByID(ctx, *user.CollegeID)
		if err == nil {
			profile.CollegeName = college.Name
			profile.CollegeStatus = string(college.Status)
		}
		// A lookup failure here (e.g. college deleted) just means no
		// college badge is shown — not worth failing the whole profile.
	}

	return profile, nil
}
