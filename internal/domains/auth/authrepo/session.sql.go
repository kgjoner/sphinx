package authrepo

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/lib/pq"
)

func (q DAO) InsertSession(session *auth.Session) error {
	raw, exists := rawQueries["CreateSession"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		session.ID,
		session.SubjectID,
		session.AudienceID,
		session.RefreshToken,
		pq.Array(session.ElapsedMinutesBetweenRefreshes),
		session.RefreshesCount,
		session.Device,
		session.IP,
		session.IsActive,
	)
	return err
}

func (q DAO) UpdateSession(session auth.Session) error {
	raw, exists := rawQueries["UpdateSession"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		session.ID,
		session.RefreshToken,
		session.RefreshedAt,
		pq.Array(session.ElapsedMinutesBetweenRefreshes),
		session.RefreshesCount,
		session.IsActive,
		session.TerminatedAt,
		session.UpdatedAt,
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
		dbhandler.IntArray(&session.ElapsedMinutesBetweenRefreshes),
		&session.RefreshesCount,
		&session.IsActive,
		&session.TerminatedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
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
