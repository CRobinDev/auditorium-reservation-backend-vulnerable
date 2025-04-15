package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
)

type userRepository struct {
	conn *sqlx.DB
}

func NewUserRepository(conn *sqlx.DB) contract.IUserRepository {
	return &userRepository{
		conn: conn,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user *entity.User) error {
	return r.createUser(ctx, r.conn, user)
}

func (r *userRepository) createUser(ctx context.Context, tx sqlx.ExtContext, user *entity.User) error {
	_, err := tx.ExecContext(
		ctx,
		fmt.Sprintf(
			`INSERT INTO users (
				id, name, password_hash, role, email
			) VALUES ('%s', '%s', '%s', '%s', '%s')`,
			user.ID, user.Name, user.PasswordHash, user.Role, user.Email,
		),
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetUserByField(ctx context.Context, field, value string) (*entity.User, error) {
	var user entity.User

	statement := `SELECT
			id,
			name,
			email,
			password_hash,
			role,
			bio,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE ` + field + ` = $1
		AND deleted_at IS NULL
		`

	err := r.conn.GetContext(ctx, &user, statement, value)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) updateUser(ctx context.Context, tx sqlx.ExtContext, user *entity.User) error {
	_, err := sqlx.NamedExecContext(
		ctx,
		tx,
		`UPDATE users
		SET name = :name,
			email = :email,
			password_hash = :password_hash,
			role = :role,
			bio = :bio,
			updated_at = now()
		WHERE id = :id`,
		user,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	return r.updateUser(ctx, r.conn, user)
}

func (r *userRepository) deleteUser(ctx context.Context, tx sqlx.ExtContext, id uuid.UUID) error {
	res, err := tx.ExecContext(ctx,
		`UPDATE users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.deleteUser(ctx, r.conn, id)
}
func (r *userRepository) UploadProfile(ctx context.Context, id uuid.UUID, url string) error {
	return r.uploadProfile(ctx, r.conn, id, url)
}

func (r *userRepository) uploadProfile(ctx context.Context, tx sqlx.ExtContext, id uuid.UUID, url string) error {
	res, err := tx.ExecContext(ctx, 
	fmt.Sprintf(`UPDATE users SET photo_url = $1 WHERE id = $2`), url, id)
    if err != nil {
        return fmt.Errorf("failed to update photo URL: %w", err)
    }

    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to check affected rows: %w", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("no user found with id %v", id)
    }

    return nil
}