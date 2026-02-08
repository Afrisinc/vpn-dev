package repository

import (
	"context"
	"fmt"

	"github.com/ElissaDesign/vpn-dev/internal/errors"
	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository handles database operations for users
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

// GetAll retrieves all users from the database
func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	query := `
		SELECT id, email, ip::text, private_key, public_key, status, server_id, created_at, updated_at, last_connected
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.IP,
			&user.PrivateKey,
			&user.PublicKey,
			&user.Status,
			&user.ServerID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastConnected,
		); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return users, nil
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	query := `
		INSERT INTO users (id, email, ip, private_key, public_key, status, server_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.ID,
		user.Email,
		user.IP,
		user.PrivateKey,
		user.PublicKey,
		user.Status,
		user.ServerID,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok {
			// Handle unique constraint violations
			if pgErr.Code == "23505" { // unique_violation
				constraint := pgErr.ConstraintName
				switch constraint {
				case "users_email_key":
					return errors.ErrDuplicateEmail
				case "users_ip_key":
					return errors.ErrDuplicateIP
				case "users_public_key_key":
					return errors.ErrDuplicatePublicKey
				}
			}
		}
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return nil
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, ip::text, private_key, public_key, status, server_id, created_at, updated_at, last_connected
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.IP,
		&user.PrivateKey,
		&user.PublicKey,
		&user.Status,
		&user.ServerID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastConnected,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return &user, nil
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, ip::text, private_key, public_key, status, server_id, created_at, updated_at, last_connected
		FROM users
		WHERE email = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.IP,
		&user.PrivateKey,
		&user.PublicKey,
		&user.Status,
		&user.ServerID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastConnected,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return &user, nil
}

// FindByIP retrieves a user by IP address
func (r *UserRepository) FindByIP(ctx context.Context, ip string) (*model.User, error) {
	query := `
		SELECT id, email, ip::text, private_key, public_key, status, server_id, created_at, updated_at, last_connected
		FROM users
		WHERE ip = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, ip).Scan(
		&user.ID,
		&user.Email,
		&user.IP,
		&user.PrivateKey,
		&user.PublicKey,
		&user.Status,
		&user.ServerID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastConnected,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return &user, nil
}

// FindByPublicKey retrieves a user by WireGuard public key
func (r *UserRepository) FindByPublicKey(ctx context.Context, publicKey string) (*model.User, error) {
	query := `
		SELECT id, email, ip::text, private_key, public_key, status, server_id, created_at, updated_at, last_connected
		FROM users
		WHERE public_key = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, publicKey).Scan(
		&user.ID,
		&user.Email,
		&user.IP,
		&user.PrivateKey,
		&user.PublicKey,
		&user.Status,
		&user.ServerID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastConnected,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return &user, nil
}

// UpdateStatus updates a user's connection status
func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}

// UpdateLastConnected updates the last connection timestamp
func (r *UserRepository) UpdateLastConnected(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET last_connected = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}

// GetByServerID retrieves all users on a specific server
func (r *UserRepository) GetByServerID(ctx context.Context, serverID string) ([]model.User, error) {
	query := `
		SELECT id, email, ip::text, private_key, public_key, status, server_id, created_at, updated_at, last_connected
		FROM users
		WHERE server_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.IP,
			&user.PrivateKey,
			&user.PublicKey,
			&user.Status,
			&user.ServerID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastConnected,
		); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return users, nil
}

// UpdateServerID updates a user's assigned server
func (r *UserRepository) UpdateServerID(ctx context.Context, id uuid.UUID, serverID *string) error {
	query := `
		UPDATE users
		SET server_id = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, serverID, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}

// Delete deletes a user from the database (used for transaction rollback)
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}
