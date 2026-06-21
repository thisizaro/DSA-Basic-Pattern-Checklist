// Package services contains business logic. Handlers call into services;
// services call into repositories. Services never touch *http.Request or
// SQL directly — that separation is what makes each layer independently
// testable and replaceable.
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yourname/dsa-tracker/internal/models"
	"github.com/yourname/dsa-tracker/internal/oauth"
	"github.com/yourname/dsa-tracker/internal/repository"
	"github.com/yourname/dsa-tracker/internal/utils"
)

// ErrEmailNotVerified is returned if Google reports the user's email as unverified.
var ErrEmailNotVerified = errors.New("google email is not verified")

// AuthService handles Google OAuth sign-in, college-domain resolution, and
// session token issuance. There is no password path — Google is the only
// identity provider.
type AuthService struct {
	users     repository.UserRepository
	colleges  repository.CollegeRepository
	google    *oauth.GoogleClient
	jwtSecret string
	jwtExpiry time.Duration
}

// NewAuthService constructs an AuthService.
func NewAuthService(
	users repository.UserRepository,
	colleges repository.CollegeRepository,
	google *oauth.GoogleClient,
	jwtSecret string,
	jwtExpiry time.Duration,
) *AuthService {
	return &AuthService{
		users:     users,
		colleges:  colleges,
		google:    google,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// GoogleAuthCodeURL returns the URL to send the user's browser to in order
// to start the Google consent flow, embedding the given CSRF state.
func (s *AuthService) GoogleAuthCodeURL(state string) string {
	return s.google.AuthCodeURL(state)
}

// HandleGoogleCallback exchanges the OAuth code for Google profile info,
// then finds-or-creates the local user account, resolving their college
// tag and account status along the way. Returns the user and a signed
// session JWT. The JWT is issued even for non-active accounts — the
// frontend uses it to fetch GET /api/auth/me, which reports the status so
// the right blocked/pending screen can be shown; a blocked/pending account
// just can't reach any other endpoint (enforced by AuthMiddleware).
func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, error) {
	token, err := s.google.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("exchanging google code: %w", err)
	}

	info, err := s.google.FetchUserInfo(ctx, token)
	if err != nil {
		return nil, "", fmt.Errorf("fetching google profile: %w", err)
	}
	if !info.VerifiedEmail {
		return nil, "", ErrEmailNotVerified
	}

	email := strings.ToLower(strings.TrimSpace(info.Email))

	// Returning user: their college/status were resolved at signup time
	// and normally don't need re-resolving. The one exception is a
	// blocked account — that's the case where you, the admin, flip
	// is_dev=true directly in the DB after the fact, and we want that
	// change to unblock them on their very next login rather than
	// requiring some separate "recheck" action.
	existing, err := s.users.GetByGoogleID(ctx, info.ID)
	if err == nil {
		if existing.AccountStatus == models.AccountBlockedUnrecognizedDomain && existing.IsDev {
			if updateErr := s.users.UpdateAccountStatus(ctx, existing.ID, models.AccountActive); updateErr != nil {
				return nil, "", fmt.Errorf("unblocking dev account: %w", updateErr)
			}
			existing.AccountStatus = models.AccountActive
		}

		sessionToken, err := utils.GenerateJWT(existing.ID, s.jwtSecret, s.jwtExpiry)
		if err != nil {
			return nil, "", fmt.Errorf("generating token: %w", err)
		}
		return existing, sessionToken, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, "", fmt.Errorf("looking up user by google id: %w", err)
	}

	// New user: resolve college + account status from the email domain,
	// then create the account.
	collegeID, accountStatus, err := s.resolveCollegeAndStatus(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("resolving college: %w", err)
	}

	username, err := s.generateUniqueUsername(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("generating username: %w", err)
	}

	newUser := &models.User{
		GoogleID:      info.ID,
		Email:         email,
		Username:      username,
		DisplayName:   displayNameOrFallback(info.Name, email),
		AvatarURL:     info.Picture,
		CollegeID:     collegeID,
		AccountStatus: accountStatus,
	}
	if err := s.users.Create(ctx, newUser); err != nil {
		return nil, "", fmt.Errorf("creating user: %w", err)
	}

	sessionToken, err := utils.GenerateJWT(newUser.ID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("generating token: %w", err)
	}

	return newUser, sessionToken, nil
}

// resolveCollegeAndStatus implements the domain → college → account-status
// rules:
//   - known, approved domain   -> that college, account active
//   - known, pending domain    -> that college, account pending_review
//     (a second/third person from the same still-unreviewed domain also
//     lands here, not blocked — only the *domain* needs review once)
//   - unknown domain           -> auto-create it as pending, account
//     pending_review (this person is "the first to enter the name")
//   - common consumer email domain (gmail.com etc) -> blocked_unrecognized_domain;
//     you flip is_dev=true manually in the DB afterwards, and their *next*
//     login succeeds, since existing users skip this whole function.
func (s *AuthService) resolveCollegeAndStatus(ctx context.Context, email string) (*string, models.AccountStatus, error) {
	domain := domainFromEmail(email)
	if domain == "" {
		return nil, models.AccountBlockedUnrecognizedDomain, nil
	}

	college, err := s.colleges.GetByDomain(ctx, domain)
	if errors.Is(err, repository.ErrNotFound) {
		// First time anyone has signed in with this domain. Per product
		// decision, treat it as a "new college" candidate (pending_review)
		// rather than an outright block — most unrecognized domains here
		// will be legitimate colleges that just aren't in the table yet,
		// not consumer email providers. Common consumer domains are
		// excluded below so gmail.com etc. don't spam the pending queue.
		if isCommonConsumerDomain(domain) {
			return nil, models.AccountBlockedUnrecognizedDomain, nil
		}

		created, err := s.colleges.CreatePending(ctx, domain)
		if err != nil {
			return nil, "", fmt.Errorf("creating pending college: %w", err)
		}
		return &created.ID, models.AccountPendingReview, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("looking up college: %w", err)
	}

	if college.Status == models.CollegeApproved {
		return &college.ID, models.AccountActive, nil
	}
	return &college.ID, models.AccountPendingReview, nil
}

// commonConsumerDomains are mainstream personal-email providers that
// should never silently become "pending college" entries — anyone signing
// in with one of these and no is_dev flag set gets blocked outright,
// matching the "non-college domains are blocked, contact admin" flow.
var commonConsumerDomains = map[string]bool{
	"gmail.com":      true,
	"googlemail.com": true,
	"yahoo.com":      true,
	"outlook.com":    true,
	"hotmail.com":    true,
	"icloud.com":     true,
	"protonmail.com": true,
	"proton.me":      true,
}

func isCommonConsumerDomain(domain string) bool {
	return commonConsumerDomains[domain]
}

// domainFromEmail extracts the part after '@', lowercased. Returns "" if
// the email doesn't look like a valid address (defensive — Google's
// emails are always well-formed, but don't trust blindly).
func domainFromEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return ""
	}
	return strings.ToLower(email[at+1:])
}

// generateUniqueUsername slugifies the email's local part and appends a
// numeric suffix if needed to guarantee uniqueness.
func (s *AuthService) generateUniqueUsername(ctx context.Context, email string) (string, error) {
	local := email
	if at := strings.Index(email, "@"); at >= 0 {
		local = email[:at]
	}
	base := utils.Slugify(local)

	candidate := base
	for i := 0; i < 50; i++ {
		taken, err := s.users.UsernameTaken(ctx, candidate, "")
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i+2)
	}
	// Extremely unlikely fallback if 50 collisions happened, ensures we
	// never loop forever or fail signup outright.
	suffix, err := utils.RandomToken(4)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", base, suffix), nil
}

func displayNameOrFallback(name, email string) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	if at := strings.Index(email, "@"); at > 0 {
		return email[:at]
	}
	return "User"
}

// ValidateToken parses a JWT and loads the associated user. Used by auth
// middleware on every protected request.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	userID, err := utils.ParseJWT(token, s.jwtSecret)
	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, utils.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("looking up user from token: %w", err)
	}

	return user, nil
}

// Sentinel errors for NameMyPendingCollege.
var (
	ErrNoCollegeToName    = errors.New("your account isn't linked to a pending college")
	ErrCollegeNameInvalid = errors.New("college name must be 2-100 characters")
)

// NameMyPendingCollege lets a user whose account is pending_review supply
// a proper display name for their (currently domain-named) college — the
// "be the first to enter the name" step in the product's onboarding flow.
// Only works while the college is still pending review; once approved,
// changing the name requires direct DB access.
func (s *AuthService) NameMyPendingCollege(ctx context.Context, userID string, name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return ErrCollegeNameInvalid
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("loading user: %w", err)
	}
	if user.AccountStatus != models.AccountPendingReview || user.CollegeID == nil {
		return ErrNoCollegeToName
	}

	if err := s.colleges.SetName(ctx, *user.CollegeID, name); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// College was already approved/renamed between the page load
			// and this request — not an error worth surfacing loudly.
			return ErrNoCollegeToName
		}
		return fmt.Errorf("setting college name: %w", err)
	}

	return nil
}
