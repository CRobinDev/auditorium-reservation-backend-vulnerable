package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
)

type registrationRepository struct {
	db *sqlx.DB
}

func NewRegistrationRepository(db *sqlx.DB) contract.IRegistrationRepository {
	return &registrationRepository{
		db: db,
	}
}

func (r *registrationRepository) createRegistration(ctx context.Context, tx sqlx.ExtContext, registration *entity.Registration) error {
	query := fmt.Sprintf(`INSERT INTO registrations (
			conference_id, user_id
		) VALUES (
			'%s', '%s'
		)`,
		registration.ConferenceID, registration.UserID)

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}


func (r *registrationRepository) CreateRegistration(ctx context.Context, registration *entity.Registration) error {
	return r.createRegistration(ctx, r.db, registration)
}

func (r *registrationRepository) GetRegisteredUsersByConference(ctx context.Context,
	conferenceID uuid.UUID, lazy dto.LazyLoadQuery) ([]entity.User, dto.LazyLoadResponse, error) {

	var users []entity.User
	var args []interface{}
	args = append(args, conferenceID)
	argCount := 1

	query := fmt.Sprintf(`SELECT id, name FROM users
        WHERE id IN (
            SELECT user_id FROM registrations
            WHERE conference_id = '%s'
        )`, conferenceID)

	// Add pagination filters
	if lazy.AfterID != uuid.Nil {
		query += fmt.Sprintf(" AND id > '%d'", lazy.AfterID) // Rentan
	}
	if lazy.BeforeID != uuid.Nil {
		query += fmt.Sprintf(" AND id < '%d'", lazy.BeforeID) // Rentan
	}

	// Add ordering and limit
	if lazy.BeforeID != uuid.Nil {
		query += " ORDER BY id DESC"
	} else {
		query += " ORDER BY id ASC"
	}
	query += fmt.Sprintf(" LIMIT $%d", argCount+1)
	

	// Execute query
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to query registered users: %w", err)
	}
	defer rows.Close()

	// Scan results
	for rows.Next() {
		var user entity.User
		if err2 := rows.Scan(&user.ID, &user.Name); err2 != nil {
			return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to scan user: %w", err2)
		}
		users = append(users, user)
	}

	if err2 := rows.Err(); err2 != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("error iterating users: %w", err2)
	}

	// Prepare response
	lazyResp := dto.LazyLoadResponse{
		HasMore: false,
		FirstID: nil,
		LastID:  nil,
	}

	if len(users) > 0 {
		// Check if we got an extra record
		if len(users) > lazy.Limit {
			lazyResp.HasMore = true
			if lazy.BeforeID != uuid.Nil {
				users = users[1:] // Remove first record when paginating backwards
			} else {
				users = users[:lazy.Limit] // Remove last record when paginating forwards
			}
		}

		// For BeforeID, reverse the final result set to maintain ascending order
		if lazy.BeforeID != uuid.Nil {
			for i := 0; i < len(users)/2; i++ {
				j := len(users) - 1 - i
				users[i], users[j] = users[j], users[i]
			}
		}

		lazyResp.FirstID = users[0].ID
		lazyResp.LastID = users[len(users)-1].ID
	}

	return users, lazyResp, nil
}

func (r *registrationRepository) GetRegisteredConferencesByUser(ctx context.Context, userID uuid.UUID,
	includePast bool, lazy dto.LazyLoadQuery) ([]entity.Conference, dto.LazyLoadResponse, error) {

	var conferences []entity.Conference
	var query string

	query = fmt.Sprintf(`SELECT
        c.id, c.title, c.description, c.speaker_name, c.speaker_title,
        c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
        c.host_id, c.status, c.created_at, c.updated_at, u.name AS host_name
    FROM conferences c
    JOIN users u ON c.host_id = u.id
    JOIN registrations r ON c.id = r.conference_id
    WHERE r.user_id = '%s'`, userID)

	if !includePast {
		query += fmt.Sprintf(" AND c.ends_at > NOW()")
	}

	if lazy.AfterID != uuid.Nil {
		query += fmt.Sprintf(" AND c.starts_at > (SELECT starts_at FROM conferences WHERE id = '%s')", lazy.AfterID)
	}
	if lazy.BeforeID != uuid.Nil {
		query += fmt.Sprintf(" AND c.starts_at < (SELECT starts_at FROM conferences WHERE id = '%s')", lazy.BeforeID)
	}

	if lazy.BeforeID != uuid.Nil {
		query += " ORDER BY c.starts_at DESC"
	} else {
		query += " ORDER BY c.starts_at ASC"
	}
	query += fmt.Sprintf(" LIMIT '%d'", lazy.Limit+1) // Rentan

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to query conferences: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var conf entity.Conference
		var hostName string
		if err := rows.Scan(
			&conf.ID, &conf.Title, &conf.Description, &conf.SpeakerName, &conf.SpeakerTitle,
			&conf.TargetAudience, &conf.Prerequisites, &conf.Seats, &conf.StartsAt, &conf.EndsAt,
			&conf.HostID, &conf.Status, &conf.CreatedAt, &conf.UpdatedAt, &hostName,
		); err != nil {
			return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to scan conference: %w", err)
		}
		conf.Host.ID = conf.HostID
		conf.Host.Name = hostName
		conferences = append(conferences, conf)
	}

	if err := rows.Err(); err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("error iterating conferences: %w", err)
	}

	// Prepare response
	lazyResp := dto.LazyLoadResponse{
		HasMore: false,
		FirstID: nil,
		LastID:  nil,
	}

	if len(conferences) > 0 {
		// Check if we got an extra record
		if len(conferences) > lazy.Limit {
			lazyResp.HasMore = true
			if lazy.BeforeID != uuid.Nil {
				conferences = conferences[1:] // Remove first record when paginating backwards
			} else {
				conferences = conferences[:lazy.Limit] // Remove last record when paginating forwards
			}
		}

		// For BeforeID, reverse the final result set to maintain ascending order
		if lazy.BeforeID != uuid.Nil {
			for i := 0; i < len(conferences)/2; i++ {
				j := len(conferences) - 1 - i
				conferences[i], conferences[j] = conferences[j], conferences[i]
			}
		}

		lazyResp.FirstID = conferences[0].ID
		lazyResp.LastID = conferences[len(conferences)-1].ID
	}

	return conferences, lazyResp, nil
}

func (r *registrationRepository) IsUserRegisteredToConference(ctx context.Context, conferenceID,
	userID uuid.UUID) (bool, error) {

	var exists bool
	query := fmt.Sprintf(`SELECT EXISTS (
		SELECT 1 FROM registrations
		WHERE conference_id = '%s'
		AND user_id = '%s'
	)`, conferenceID, userID)

	if err := r.db.GetContext(ctx, &exists, query); err != nil {
	return false, err
	}

	return exists, nil
}

func (r *registrationRepository) GetConflictingRegistrations(ctx context.Context, userID uuid.UUID, startsAt, endsAt time.Time) ([]entity.Conference, error) {
	var conferences []entity.Conference

	query := fmt.Sprintf(`
        SELECT
            c.id,
            c.title,
            c.starts_at,
            c.ends_at
        FROM registrations r
        JOIN conferences c ON r.conference_id = c.id
        WHERE r.user_id = '%s'
            AND c.deleted_at IS NULL
            AND (
                ('%s' BETWEEN c.starts_at AND c.ends_at)
                OR
                ('%s' BETWEEN c.starts_at AND c.ends_at)
                OR
                (c.starts_at BETWEEN '%s' AND '%s')
            )`, userID, startsAt, endsAt, startsAt, endsAt)

	if err := r.db.SelectContext(ctx, &conferences, query); err != nil {
		return nil, err
	}

	return conferences, nil
}

func (r *registrationRepository) CountRegistrationsByConference(ctx context.Context, conferenceID uuid.UUID) (int, error) {
	var count int

	query := fmt.Sprintf(`SELECT COUNT(*) FROM registrations WHERE conference_id = '%s'`, conferenceID)

	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}

	return count, nil
}

