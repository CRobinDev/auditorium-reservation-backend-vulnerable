package repository

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/redis/go-redis/v9"
	"time"
)

type authRepository struct {
	db  *sqlx.DB
	rds *redis.Client
}

func NewAuthRepository(db *sqlx.DB, rds *redis.Client) contract.IAuthRepository {
	return &authRepository{
		db:  db,
		rds: rds,
	}
}

func (r *authRepository) SetOTPRegisterUser(ctx context.Context, email string, otp string) error {
	return r.rds.Set(ctx, "auth:"+email+":register_otp", otp, 10*time.Minute).Err()
}

func (r *authRepository) GetOTPRegisterUser(ctx context.Context, email string) (string, error) {
	return r.rds.Get(ctx, "auth:"+email+":register_otp").Result()
}

func (r *authRepository) DeleteOTPRegisterUser(ctx context.Context, email string) error {
	return r.rds.Del(ctx, "auth:"+email+":register_otp").Err()
}

func (r *authRepository) CreateAuthSession(ctx context.Context, session *entity.AuthSession) error {
	return r.createAuthSession(ctx, r.db, session)
}

func (r *authRepository) createAuthSession(ctx context.Context, tx sqlx.ExtContext, authSession *entity.AuthSession) error {
	query := `INSERT INTO auth_sessions (token, user_id, expires_at)
				VALUES (:token, :user_id, :expires_at)
				ON CONFLICT (user_id) DO UPDATE SET token = :token, expires_at = :expires_at`

	_, err := sqlx.NamedExecContext(ctx, tx, query, authSession)
	if err != nil {
		return err
	}

	return nil
}

func (r *authRepository) GetAuthSessionByToken(ctx context.Context, token string) (*entity.AuthSession, error) {
	var authSession entity.AuthSession

	statement := `SELECT
    		token,
			user_id,
			expires_at
		FROM auth_sessions
		WHERE token = $1
		`

	err := r.db.GetContext(ctx, &authSession, statement, token)
	if err != nil {
		return nil, err
	}

	return &authSession, nil
}

func (r *authRepository) deleteAuthSession(ctx context.Context, tx sqlx.ExtContext, userID uuid.UUID) error {
	query := `DELETE FROM auth_sessions WHERE user_id = $1`

	res, err := tx.ExecContext(ctx, query, userID)
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

func (r *authRepository) DeleteAuthSession(ctx context.Context, userID uuid.UUID) error {
	return r.deleteAuthSession(ctx, r.db, userID)
}

func (r *authRepository) SetOTPResetPassword(ctx context.Context, email, otp string) error {
	return r.rds.Set(ctx, "auth:"+email+":reset_password_otp", otp, 10*time.Minute).Err()
}

func (r *authRepository) GetOTPResetPassword(ctx context.Context, email string) (string, error) {
	return r.rds.Get(ctx, "auth:"+email+":reset_password_otp").Result()
}

func (r *authRepository) DeleteOTPResetPassword(ctx context.Context, email string) error {
	return r.rds.Del(ctx, "auth:"+email+":reset_password_otp").Err()
}
