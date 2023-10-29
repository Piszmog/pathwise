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

func (s *JobApplicationNoteStore) GetByID(ctx context.Context, id int) (types.JobApplicationNote, error) {
	row := s.Database.DB().QueryRowContext(
		ctx,
		`
		SELECT
		    n.id, n.job_application_id, n.note, n.created_at
		FROM 
		    job_application_notes n
		WHERE
		    n.id = ?`,
		id,
	)
	return scanJobApplicationNote(row)
}

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

func (s *JobApplicationNoteStore) Get(ctx context.Context, opts GetOpts) ([]types.JobApplicationNote, error) {
	rows, err := s.Database.DB().QueryContext(
		ctx,
		`
		SELECT
		    n.id, n.job_application_id, n.note, n.created_at
		FROM 
		    job_application_notes n
		ORDER BY n.created_at DESC
		LIMIT ? OFFSET ?`,
		opts.PerPage,
		opts.Page*opts.PerPage,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobApplicationNotes(rows)
}

func (s *JobApplicationNoteStore) GetAllByID(ctx context.Context, id int) ([]types.JobApplicationNote, error) {
	rows, err := s.Database.DB().QueryContext(
		ctx,
		`
		SELECT
		    n.id, n.job_application_id, n.note, n.created_at
		FROM 
		    job_application_notes n
		WHERE
		    n.job_application_id = ?
		ORDER BY n.created_at DESC`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobApplicationNotes(rows)
}

func scanJobApplicationNotes(rows *sql.Rows) ([]types.JobApplicationNote, error) {
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
	res, err := tx.ExecContext(
		ctx,
		`INSERT INTO job_application_notes (job_application_id, note) VALUES (?, ?)`,
		rec.JobApplicationID,
		rec.Note,
	)
	if err != nil {
		tx.Rollback()
		return types.JobApplicationNote{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return types.JobApplicationNote{}, err
	}
	row := tx.QueryRowContext(
		ctx,
		`SELECT id, job_application_id, note, created_at FROM job_application_notes WHERE id = ?`,
		id,
	)
	note, err := scanJobApplicationNote(row)
	if err != nil {
		tx.Rollback()
		return types.JobApplicationNote{}, err
	}
	return note, tx.Commit()
}

func (s *JobApplicationNoteStore) Update(ctx context.Context, rec types.JobApplicationNote) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`UPDATE job_application_notes SET note = ? WHERE id = ?`,
		rec.Note,
		rec.ID,
	)
	return err
}

func (s *JobApplicationNoteStore) Delete(ctx context.Context, id int) error {
	_, err := s.Database.DB().ExecContext(
		ctx,
		`DELETE FROM job_application_notes WHERE id = ?`,
		id,
	)
	return err
}
