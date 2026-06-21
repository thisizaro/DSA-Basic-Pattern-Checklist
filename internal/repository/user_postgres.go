package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/dsa-tracker/internal/models"
)

// ErrNotFound is returned by repository methods when a row doesn't exist.
// Services check for this with errors.Is to distinguish "not found" from
// real failures.
var ErrNotFound = errors.New("not found")

// PostgresUserRepository implements UserRepository against a Postgres pool.
type PostgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository constructs a PostgresUserRepository.
func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// userColumns is the fixed column list/order used by every SELECT below,
// kept as one constant so the Scan order can never drift from the query.
const userColumns = `
	id, google_id, email, username, display_name, avatar_url,
	college_id, leetcode_url, is_anonymous, is_dev, account_status,
	created_at, updated_at`

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.GoogleID, &u.Email, &u.Username, &u.DisplayName, &u.AvatarURL,
		&u.CollegeID, &u.LeetCodeURL, &u.IsAnonymous, &u.IsDev, &u.AccountStatus,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning user: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := fmt.Sprintf(`
		INSERT INTO users (google_id, email, username, display_name, avatar_url, college_id, account_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING %s`, userColumns)

	row := r.db.QueryRow(ctx, query,
		user.GoogleID, user.Email, user.Username, user.DisplayName, user.AvatarURL,
		user.CollegeID, string(user.AccountStatus),
	)

	created, err := scanUser(row)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	*user = *created
	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, userColumns)
	return scanUser(r.db.QueryRow(ctx, query, id))
}

func (r *PostgresUserRepository) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE google_id = $1`, userColumns)
	return scanUser(r.db.QueryRow(ctx, query, googleID))
}

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE username = $1`, userColumns)
	return scanUser(r.db.QueryRow(ctx, query, username))
}

func (r *PostgresUserRepository) UsernameTaken(ctx context.Context, username string, excludeUserID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2)`

	var taken bool
	// excludeUserID is "" at signup time (no existing row to exclude yet).
	// A zero-UUID-like empty string never matches a real id, so the
	// `id != $2` comparison is harmless in that case.
	err := r.db.QueryRow(ctx, query, username, nullableExclude(excludeUserID)).Scan(&taken)
	if err != nil {
		return false, fmt.Errorf("checking username availability: %w", err)
	}
	return taken, nil
}

// nullableExclude turns an empty exclude-ID into a value that can never
// equal a real UUID, so "id != $2" behaves correctly even when there's no
// user to exclude yet.
func nullableExclude(id string) string {
	if id == "" {
		return "00000000-0000-0000-0000-000000000000"
	}
	return id
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	query := fmt.Sprintf(`
		UPDATE users
		SET username = $1, leetcode_url = $2, is_anonymous = $3, updated_at = now()
		WHERE id = $4
		RETURNING %s`, userColumns)

	row := r.db.QueryRow(ctx, query, user.Username, user.LeetCodeURL, user.IsAnonymous, user.ID)

	updated, err := scanUser(row)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	*user = *updated
	return nil
}

func (r *PostgresUserRepository) UpdateAccountStatus(ctx context.Context, userID string, status models.AccountStatus) error {
	const query = `UPDATE users SET account_status = $1, updated_at = now() WHERE id = $2`

	tag, err := r.db.Exec(ctx, query, string(status), userID)
	if err != nil {
		return fmt.Errorf("updating account status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
