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

func (r *conferenceRepository) createConference(ctx context.Context, tx sqlx.ExtContext, conference *entity.Conference) error {
	// Rentan terhadap SQL Injection: Menggunakan fmt.Sprintf untuk menggabungkan input langsung
	query := fmt.Sprintf(`INSERT INTO conferences (
                         id, title, description, speaker_name, speaker_title,
                         target_audience, prerequisites, seats, starts_at, ends_at,
                         host_id, status
					) VALUES (
					          '%s', '%s', '%s', '%s', '%s',
					          '%s', '%s', '%d', '%s', '%s',
					          '%s', '%s')`,
		conference.ID, conference.Title, conference.Description, conference.SpeakerName, conference.SpeakerTitle,
		conference.TargetAudience, *conference.Prerequisites, conference.Seats, conference.StartsAt, conference.EndsAt,
		conference.HostID, conference.Status)

	_, err := tx.ExecContext(ctx, query)
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

	// Rentan terhadap SQL Injection: Menggunakan fmt.Sprintf untuk menggabungkan input langsung
	query := fmt.Sprintf(`
		SELECT
			c.id, c.title, c.description, c.speaker_name, c.speaker_title,
			c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
			c.host_id, c.status, c.created_at, c.updated_at, u.name AS host_name,
			COUNT(r.user_id) AS registration_count
		FROM conferences c
		JOIN users u ON c.host_id = u.id
		LEFT JOIN registrations r ON c.id = r.conference_id
		WHERE c.id = '%s'
		AND c.deleted_at IS NULL
		GROUP BY
			c.id, c.title, c.description, c.speaker_name, c.speaker_title,
			c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
			c.host_id, c.status, c.created_at, c.updated_at, u.name
	`, id)

	err := r.db.GetContext(ctx, &row, query)
	if err != nil {
		return nil, err
	}

	conference := row.ToEntity()
	return &conference, nil
}
	

func (r *conferenceRepository) GetConferences(ctx context.Context,
	query *dto.GetConferenceQuery) ([]entity.Conference, dto.LazyLoadResponse, error) {

	// Build base query
	baseQuery := fmt.Sprintf(`
        SELECT
            c.id, c.title, c.description, c.speaker_name, c.speaker_title,
            c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
            c.host_id, c.status, c.created_at, c.updated_at, u.name AS host_name,
            COUNT(r.user_id) AS registration_count
        FROM conferences c
        JOIN users u ON c.host_id = u.id
        LEFT JOIN registrations r ON c.id = r.conference_id
        WHERE c.deleted_at IS NULL`)

	

	// Build WHERE clause
	var conditions []string

	if !query.IncludePast {
		conditions = append(conditions, fmt.Sprintf("c.ends_at > NOW()"))
	}

	if query.Title != nil {
		conditions = append(conditions, fmt.Sprintf("c.title ILIKE '%%' || '%s' || '%%'", *query.Title))
	}

	if query.HostID != nil {
		conditions = append(conditions, fmt.Sprintf("c.host_id = '%s'", *query.HostID))
	}

	conditions = append(conditions, fmt.Sprintf("c.status = '%s'", query.Status))

	if query.StartsBefore != nil {
		conditions = append(conditions, fmt.Sprintf("c.starts_at < '%s'", query.StartsBefore))
	}

	if query.StartsAfter != nil {
		conditions = append(conditions, fmt.Sprintf("c.starts_at > '%s'", query.StartsAfter))
	}

	// Handle cursor-based pagination
	if query.AfterID != nil {
		if query.OrderBy == "c.created_at" {
			orderOp := ">"
			if query.Order == "desc" {
				orderOp = "<"
			}
			conditions = append(conditions, fmt.Sprintf("id %s '%s'", orderOp, *query.AfterID))
		} else {
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
                    WHERE c.id = '%s'
                )`, orderOp, *query.AfterID))
		}
	}

	if query.BeforeID != nil {
		if query.OrderBy == "c.created_at" {
			orderOp := "<"
			if query.Order == "desc" {
				orderOp = ">"
			}
			conditions = append(conditions, fmt.Sprintf("id %s '%s'", orderOp, *query.BeforeID))
		} else {
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
                    WHERE c.id = '%s'
                )`, orderOp, *query.BeforeID))
		}
	}

	// Add conditions to base query
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add GROUP BY clause before ORDER BY
	baseQuery += fmt.Sprintf(`
        GROUP BY
            c.id, c.title, c.description, c.speaker_name, c.speaker_title,
            c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
            c.host_id, c.status, c.created_at, c.updated_at, u.name`)

	// Add ORDER BY clause
	if query.OrderBy == "c.created_at" {
		orderDirection := "ASC"
		if query.Order == "desc" {
			orderDirection = "DESC"
		}
		baseQuery += fmt.Sprintf(" ORDER BY c.id %s", orderDirection)
	} else {
		orderDirection := "ASC"
		if query.Order == "desc" {
			orderDirection = "DESC"
		}
		baseQuery += fmt.Sprintf(" ORDER BY c.starts_at %s, c.id %s", orderDirection, orderDirection)
	}

	// Add LIMIT
	baseQuery += fmt.Sprintf(" LIMIT %d", query.Limit+1) // Request one extra record to determine if there are more pages

	// Execute query
	rows, err := r.db.QueryxContext(ctx, baseQuery)
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

func (r *conferenceRepository) updateConference(ctx context.Context, tx sqlx.ExtContext, conference *entity.Conference) error {
	// Rentan terhadap SQL Injection: Menggunakan fmt.Sprintf untuk menggabungkan input langsung
	query := fmt.Sprintf(`UPDATE conferences
		SET title = '%s',
			description = '%s',
			speaker_name = '%s',
			speaker_title = '%s',
			target_audience = '%s',
			prerequisites = '%s',
			seats = '%d',
			starts_at = '%s',
			ends_at = '%s',
			host_id = '%s',
			status = '%s',
			updated_at = now()
		WHERE id = '%s'`,
		conference.Title, conference.Description, conference.SpeakerName, conference.SpeakerTitle,
		conference.TargetAudience, *conference.Prerequisites, conference.Seats,
		conference.StartsAt, conference.EndsAt, conference.HostID, conference.Status, conference.ID)

	res, err := tx.ExecContext(ctx, query)
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
	// Rentan terhadap SQL Injection: Menggunakan fmt.Sprintf untuk menggabungkan input langsung
	query := fmt.Sprintf(`UPDATE conferences SET deleted_at = now() WHERE id = '%s'`, id)

	res, err := tx.ExecContext(ctx, query)
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

	// Rentan terhadap SQL Injection: Menggunakan fmt.Sprintf untuk menggabungkan input langsung
	query := fmt.Sprintf(`
		SELECT
			c.id, c.title, c.description, c.speaker_name, c.speaker_title,
			c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
			c.host_id, c.status, c.created_at, c.updated_at
		FROM conferences c
		WHERE c.deleted_at IS NULL
		AND c.id != '%s'
		AND c.status = 'approved'
		AND c.starts_at < '%s'
		AND c.ends_at > '%s'
		ORDER BY c.starts_at
		LIMIT 10`, excludeID, endsAt, startsAt)

	err := r.db.SelectContext(ctx, &conferences, query)
	if err != nil {
		return nil, err
	}

	return conferences, nil
}
