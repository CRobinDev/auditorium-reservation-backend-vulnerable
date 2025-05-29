package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"time"
)

type feedbackRepository struct {
	db *sqlx.DB
}

func NewFeedbackRepository(db *sqlx.DB) contract.IFeedbackRepository {
	return &feedbackRepository{
		db: db,
	}
}

func (r *feedbackRepository) createFeedback(ctx context.Context, tx sqlx.ExtContext, feedback *entity.Feedback) error {
	query := `INSERT INTO feedbacks (id, user_id, conference_id, comment, created_at)
		VALUES (:id, :user_id, :conference_id, :comment, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, tx, query, feedback)
	return err
}

func (r *feedbackRepository) CreateFeedback(ctx context.Context, feedback *entity.Feedback) error {
	return r.createFeedback(ctx, r.db, feedback)
}

func (r *feedbackRepository) GetFeedbacksByConferenceID(ctx context.Context,
	conferenceID uuid.UUID, lazy dto.LazyLoadQuery) ([]entity.Feedback, dto.LazyLoadResponse, error) {

	var feedbacks []entity.Feedback
	var args []interface{}
	args = append(args, conferenceID)
	argCount := 1

	query := `SELECT f.id, f.user_id, f.conference_id, f.comment, f.created_at, u.name as user_name
        FROM feedbacks f
        JOIN users u ON f.user_id = u.id
        WHERE f.conference_id = $1 AND f.deleted_at IS NULL`

	// Add pagination filters
	if lazy.AfterID != uuid.Nil {
		query += fmt.Sprintf(" AND f.id > $%d", argCount+1)
		args = append(args, lazy.AfterID)
		argCount++
	}
	if lazy.BeforeID != uuid.Nil {
		query += fmt.Sprintf(" AND f.id < $%d", argCount+1)
		args = append(args, lazy.BeforeID)
		argCount++
	}

	// Add ordering and limit
	if lazy.BeforeID != uuid.Nil {
		query += " ORDER BY f.id DESC"
	} else {
		query += " ORDER BY f.id ASC"
	}
	query += fmt.Sprintf(" LIMIT $%d", argCount+1)
	args = append(args, lazy.Limit+1) // Request one extra record to determine if there are more results

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to query feedbacks: %w", err)
	}
	defer rows.Close()

	// Scan results
	for rows.Next() {
		var row struct {
			ID           uuid.UUID `db:"id"`
			UserID       uuid.UUID `db:"user_id"`
			ConferenceID uuid.UUID `db:"conference_id"`
			Comment      string    `db:"comment"`
			CreatedAt    time.Time `db:"created_at"`
			UserName     string    `db:"user_name"`
		}

		if err2 := rows.Scan(&row.ID, &row.UserID, &row.ConferenceID, &row.Comment, &row.CreatedAt,
			&row.UserName); err2 != nil {
			return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to scan feedback: %w", err2)
		}

		feedback := entity.Feedback{
			ID:           row.ID,
			UserID:       row.UserID,
			ConferenceID: row.ConferenceID,
			Comment:      row.Comment,
			CreatedAt:    row.CreatedAt,
			User: &entity.User{
				ID:   row.UserID,
				Name: row.UserName,
			},
		}
		feedbacks = append(feedbacks, feedback)
	}

	if err := rows.Err(); err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("error iterating feedbacks: %w", err)
	}

	// Prepare response
	lazyResp := dto.LazyLoadResponse{
		HasMore: false,
		FirstID: nil,
		LastID:  nil,
	}

	if len(feedbacks) > 0 {
		// Check if we got an extra record
		if len(feedbacks) > lazy.Limit {
			lazyResp.HasMore = true
			if lazy.BeforeID != uuid.Nil {
				feedbacks = feedbacks[1:] // Remove first record when paginating backwards
			} else {
				feedbacks = feedbacks[:lazy.Limit] // Remove last record when paginating forwards
			}
		}

		// For BeforeID, reverse the final result set to maintain ascending order
		if lazy.BeforeID != uuid.Nil {
			for i := 0; i < len(feedbacks)/2; i++ {
				j := len(feedbacks) - 1 - i
				feedbacks[i], feedbacks[j] = feedbacks[j], feedbacks[i]
			}
		}

		lazyResp.FirstID = feedbacks[0].ID
		lazyResp.LastID = feedbacks[len(feedbacks)-1].ID
	}

	return feedbacks, lazyResp, nil
}

func (r *feedbackRepository) deleteFeedback(ctx context.Context, tx sqlx.ExtContext, id uuid.UUID) error {
	query := `UPDATE feedbacks SET deleted_at = now() WHERE id = $1`
	res, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete feedback: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return err
}

func (r *feedbackRepository) DeleteFeedback(ctx context.Context, id uuid.UUID) error {
	return r.deleteFeedback(ctx, r.db, id)
}

func (r *feedbackRepository) IsFeedbackGiven(ctx context.Context, userID, conferenceID uuid.UUID) (bool, error) {
	var exists bool
	if err := r.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS (
				SELECT 1 FROM feedbacks
				WHERE conference_id = $1
				AND user_id = $2
			)`,
		conferenceID, userID,
	); err != nil {
		return false, err
	}

	return exists, nil
}
