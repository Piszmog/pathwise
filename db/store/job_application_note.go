package store

import (
	"context"
	"database/sql"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
)

type JobApplicationNoteStore struct {
	Database db.Database
}

func (s *JobApplicationNoteStore) GetAllByID(ctx context.Context, id int) ([]types.JobApplicationNote, error) {
	rows, err := s.Database.DB().QueryContext(ctx, noteGetAllByIDQuery, id)
	if err != nil {
		return nil, err
	}
	return scanJobApplicationNotes(rows)
}

const noteGetAllByIDQuery = `
SELECT
	n.id, n.job_application_id, n.note, n.created_at
FROM 
	job_application_notes n
WHERE
	n.job_application_id = ?
ORDER BY n.created_at DESC
`

func scanJobApplicationNotes(rows *sql.Rows) ([]types.JobApplicationNote, error) {
	defer rows.Close()
	var notes []types.JobApplicationNote
	for rows.Next() {
		var note types.JobApplicationNote
		err := rows.Scan(
			&note.ID,
			&note.JobApplicationID,
			&note.Note,
			&note.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (s *JobApplicationNoteStore) Insert(ctx context.Context, rec types.JobApplicationNote) (types.JobApplicationNote, error) {
	tx, err := s.Database.DB().BeginTx(ctx, nil)
	if err != nil {
		return types.JobApplicationNote{}, err
	}
	res, err := tx.ExecContext(ctx, noteInsertQuery, rec.JobApplicationID, rec.Note)
	if err != nil {
		tx.Rollback()
		return types.JobApplicationNote{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return types.JobApplicationNote{}, err
	}
	note, err := scanJobApplicationNote(tx.QueryRowContext(ctx, noteGetByIDQuery, id))
	if err != nil {
		tx.Rollback()
		return types.JobApplicationNote{}, err
	}
	return note, tx.Commit()
}

const noteInsertQuery = `INSERT INTO job_application_notes (job_application_id, note) VALUES (?, ?)`

const noteGetByIDQuery = `
SELECT
	n.id, n.job_application_id, n.note, n.created_at
FROM 
	job_application_notes n
WHERE
	n.id = ?
`

func scanJobApplicationNote(row *sql.Row) (types.JobApplicationNote, error) {
	var note types.JobApplicationNote
	err := row.Scan(
		&note.ID,
		&note.JobApplicationID,
		&note.Note,
		&note.CreatedAt,
	)
	return note, err
}
