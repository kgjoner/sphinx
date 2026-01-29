package authrepo

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/lib/pq"
)

func (q DAO) InsertSession(session *auth.Session) error {
	raw, exists := rawQueries["CreateSession"]
	if !exists {
		return ErrNoQuery
	}

	q.dbtx.QueryRowContext(q.ctx, raw,
		session.ID,
		session.SubjectID,
		session.AudienceID,
		session.RefreshToken,
		session.Device,
		session.IP,
	)
	return nil
}

func (q DAO) UpdateSession(session auth.Session) error {
	raw, exists := rawQueries["UpdateSession"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		session.RefreshToken,
		session.RefreshedAt,
		session.ElapsedMinutesBetweenRefreshes,
		session.RefreshesCount,
		session.IsActive,
		session.TerminatedAt,
		session.UpdatedAt,
		session.ID,
	)
	return err
}

func (q DAO) GetSessionByID(id uuid.UUID) (*auth.Session, error) {
	raw, exists := rawQueries["GetSessionByID"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, id)

	var session auth.Session
	err := row.Scan(
		&session.ID,
		&session.SubjectID,
		&session.SubjectEmail,
		&session.SubjectName,
		&session.AudienceID,
		pq.Array(&session.Roles),
		&session.IP,
		&session.Device,
		&session.RefreshToken,
		&session.RefreshedAt,
		pq.Array(&session.ElapsedMinutesBetweenRefreshes),
		&session.RefreshesCount,
		&session.IsActive,
		&session.TerminatedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (q DAO) TerminateAllSubjectSessions(subjectID uuid.UUID) error {
	raw, exists := rawQueries["TerminateAllSubjectSessions"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw, subjectID)
	return err
}
