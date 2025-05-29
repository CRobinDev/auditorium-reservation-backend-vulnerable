package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
)

type conferenceRepository struct {
	db *sqlx.DB
}

func NewConferenceRepository(db *sqlx.DB) contract.IConferenceRepository {
	return &conferenceRepository{
		db: db,
	}
}

func (r *conferenceRepository) createConference(ctx context.Context, tx sqlx.ExtContext,
	conference *entity.Conference) error {

	_, err := sqlx.NamedExecContext(
		ctx,
		tx,
		`INSERT INTO conferences (
                         id, title, description, speaker_name, speaker_title,
                         target_audience, prerequisites, seats, starts_at, ends_at,
                         host_id, status
					) VALUES (
					          :id, :title, :description, :speaker_name, :speaker_title,
					          :target_audience, :prerequisites, :seats, :starts_at, :ends_at,
					          :host_id, :status)`,
		conference,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *conferenceRepository) CreateConference(ctx context.Context, conference *entity.Conference) error {
	return r.createConference(ctx, r.db, conference)
}

func (r *conferenceRepository) GetConferenceByID(ctx context.Context, id uuid.UUID) (*entity.Conference, error) {
	var row dto.ConferenceJoinUserRow

	statement := `SELECT
						c.id, c.title, c.description, c.speaker_name, c.speaker_title,
						c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
						c.host_id, c.status, c.created_at, c.updated_at, u.name AS host_name,
						COUNT(r.user_id) AS registration_count
					FROM conferences c
					JOIN users u ON c.host_id = u.id
					LEFT JOIN registrations r ON c.id = r.conference_id
					WHERE c.id = $1
					AND c.deleted_at IS NULL
					GROUP BY
						c.id, c.title, c.description, c.speaker_name, c.speaker_title,
						c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
						c.host_id, c.status, c.created_at, c.updated_at, u.name
		`

	err := r.db.GetContext(ctx, &row, statement, id)
	if err != nil {
		return nil, err
	}

	conference := row.ToEntity()
	return &conference, nil
}

func (r *conferenceRepository) GetConferences(ctx context.Context,
	query *dto.GetConferenceQuery) ([]entity.Conference, dto.LazyLoadResponse, error) {

	// Build base query
	baseQuery := `
        SELECT
            c.id, c.title, c.description, c.speaker_name, c.speaker_title,
            c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
            c.host_id, c.status, c.created_at, c.updated_at, u.name AS host_name,
            COUNT(r.user_id) AS registration_count
        FROM conferences c
        JOIN users u ON c.host_id = u.id
        LEFT JOIN registrations r ON c.id = r.conference_id
        WHERE c.deleted_at IS NULL`

	// Initialize query arguments
	var args []interface{}

	// Build WHERE clause
	var conditions []string

	if !query.IncludePast {
		args = append(args, time.Now())
		conditions = append(conditions, fmt.Sprintf("c.ends_at > $%d", len(args)))
	}

	if query.Title != nil {
		args = append(args, *query.Title)
		conditions = append(conditions, fmt.Sprintf("c.title ILIKE '%%' || $%d || '%%'", len(args)))
	}

	if query.HostID != nil {
		args = append(args, query.HostID)
		conditions = append(conditions, fmt.Sprintf("c.host_id = $%d", len(args)))
	}

	args = append(args, query.Status)
	conditions = append(conditions, fmt.Sprintf("c.status = $%d", len(args)))

	if query.StartsBefore != nil {
		args = append(args, query.StartsBefore)
		conditions = append(conditions, fmt.Sprintf("c.starts_at < $%d", len(args)))
	}

	if query.StartsAfter != nil {
		args = append(args, query.StartsAfter)
		conditions = append(conditions, fmt.Sprintf("c.starts_at > $%d", len(args)))
	}

	// Handle cursor-based pagination
	if query.AfterID != nil {
		args = append(args, query.AfterID)

		if query.OrderBy == "c.created_at" {
			// For created_at sorting, use only ID since UUIDv7 has timestamp
			orderOp := ">"
			if query.Order == "desc" {
				orderOp = "<"
			}
			conditions = append(conditions, fmt.Sprintf("id %s $%d", orderOp, len(args)))
		} else {
			// For starts_at sorting, use composite ordering
			orderOp := ">"
			if query.Order == "desc" {
				orderOp = "<"
			}
			conditions = append(conditions, fmt.Sprintf(`
                (
                    c.starts_at, c.id
                ) %s (
                    SELECT c.starts_at, c.id
                    FROM conferences c
                    WHERE c.id = $%d
                )`, orderOp, len(args)))
		}
	}

	if query.BeforeID != nil {
		args = append(args, query.BeforeID)

		if query.OrderBy == "c.created_at" {
			// For created_at sorting, use only ID since UUIDv7 has timestamp
			orderOp := "<"
			if query.Order == "desc" {
				orderOp = ">"
			}
			conditions = append(conditions, fmt.Sprintf("id %s $%d", orderOp, len(args)))
		} else {
			// For starts_at sorting, use composite ordering
			orderOp := "<"
			if query.Order == "desc" {
				orderOp = ">"
			}
			conditions = append(conditions, fmt.Sprintf(`
                (
                    c.starts_at, c.id
                ) %s (
                    SELECT c.starts_at, c.id
                    FROM conferences c
                    WHERE c.id = $%d
                )`, orderOp, len(args)))
		}
	}

	// Add conditions to base query
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add GROUP BY clause before ORDER BY
	baseQuery += `
        GROUP BY
            c.id, c.title, c.description, c.speaker_name, c.speaker_title,
            c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
            c.host_id, c.status, c.created_at, c.updated_at, u.name`

	// Add ORDER BY clause
	if query.OrderBy == "c.created_at" {
		// For created_at, only order by id since UUIDv7 has timestamp
		orderDirection := "ASC"
		if query.Order == "desc" {
			orderDirection = "DESC"
		}
		baseQuery += fmt.Sprintf(" ORDER BY c.id %s", orderDirection)
	} else {
		// For starts_at, use composite ordering
		orderDirection := "ASC"
		if query.Order == "desc" {
			orderDirection = "DESC"
		}
		baseQuery += fmt.Sprintf(" ORDER BY c.starts_at %s, c.id %s", orderDirection, orderDirection)
	}

	// Add LIMIT
	args = append(args, query.Limit+1) // Fetch one extra record to determine if there are more pages
	baseQuery += fmt.Sprintf(" LIMIT $%d", len(args))

	// Execute query
	rows, err := r.db.QueryxContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to query conferences: %w", err)
	}
	defer rows.Close()

	// Scan results
	var conferences []entity.Conference
	for rows.Next() {
		var row dto.ConferenceJoinUserRow
		if err2 := rows.StructScan(&row); err2 != nil {
			return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to scan conference: %w", err)
		}
		conferences = append(conferences, row.ToEntity())
	}

	if err = rows.Err(); err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("error iterating conference rows: %w", err)
	}

	// Prepare pagination response
	hasMore := len(conferences) > query.Limit
	if hasMore {
		conferences = conferences[:len(conferences)-1] // Remove the extra record
	}

	var firstID, lastID *uuid.UUID
	if len(conferences) > 0 {
		firstID = &conferences[0].ID
		lastID = &conferences[len(conferences)-1].ID
	}

	lazyLoadResponse := dto.LazyLoadResponse{
		HasMore: hasMore,
		FirstID: firstID,
		LastID:  lastID,
	}

	return conferences, lazyLoadResponse, nil
}

func (r *conferenceRepository) updateConference(ctx context.Context, tx sqlx.ExtContext,
	conference *entity.Conference) error {

	res, err := sqlx.NamedExecContext(
		ctx,
		tx,
		`UPDATE conferences
		SET title = :title,
			description = :description,
			speaker_name = :speaker_name,
			speaker_title = :speaker_title,
			target_audience = :target_audience,
			prerequisites = :prerequisites,
			seats = :seats,
			starts_at = :starts_at,
			ends_at = :ends_at,
			host_id = :host_id,
			status = :status,
			updated_at = now()
		WHERE id = :id`,
		conference,
	)
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

func (r *conferenceRepository) UpdateConference(ctx context.Context, conference *entity.Conference) error {
	return r.updateConference(ctx, r.db, conference)
}

func (r *conferenceRepository) deleteConference(ctx context.Context, tx sqlx.ExtContext, id uuid.UUID) error {
	res, err := tx.ExecContext(ctx, `UPDATE conferences SET deleted_at = now() WHERE id = $1`, id)
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

func (r *conferenceRepository) DeleteConference(ctx context.Context, id uuid.UUID) error {
	return r.deleteConference(ctx, r.db, id)
}

func (r *conferenceRepository) GetConferencesConflictingWithTime(ctx context.Context, startsAt,
	endsAt time.Time, excludeID uuid.UUID) ([]entity.Conference, error) {

	var conferences []entity.Conference

	err := r.db.SelectContext(ctx, &conferences, `
		SELECT
			c.id, c.title, c.description, c.speaker_name, c.speaker_title,
			c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
			c.host_id, c.status, c.created_at, c.updated_at
		FROM conferences c
		WHERE c.deleted_at IS NULL
		AND c.id != $1
		AND c.status = 'approved'
		AND c.starts_at < $2
		AND c.ends_at > $3
		ORDER BY c.starts_at
		LIMIT 10
		`, excludeID, endsAt, startsAt)
	if err != nil {
		return nil, err
	}

	return conferences, nil
}
