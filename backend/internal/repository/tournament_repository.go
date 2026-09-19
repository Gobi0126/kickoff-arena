package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avanthika/efootball-backend/internal/models"
)

var ErrTournamentNotFound = errors.New("tournament not found")

type TournamentRepository interface {
	CreateSpinWheel(ctx context.Context, name string, createdBy string, bracketSize int) (*models.Tournament, error)
	FindByID(ctx context.Context, id string) (*models.Tournament, error)
	FindByRegistrationToken(ctx context.Context, token string) (*models.Tournament, error)
	ListByCreator(ctx context.Context, createdBy string) ([]models.Tournament, error)
	UpdateStatus(ctx context.Context, id string, status models.TournamentStatus) error
	UpdateBracketSize(ctx context.Context, id string, bracketSize int) error
	Delete(ctx context.Context, id string) error

	// BeginTx and the *Tx methods below support locking the tournament row
	// (SELECT ... FOR UPDATE) for the duration of a check-then-act sequence,
	// so concurrent requests for the same tournament serialize instead of
	// racing (e.g. two "Start Spin" clicks, or two registrations landing on
	// the last open slot at the same time).
	BeginTx(ctx context.Context) (pgx.Tx, error)
	FindByIDForUpdateTx(ctx context.Context, tx pgx.Tx, id string) (*models.Tournament, error)
	UpdateStatusTx(ctx context.Context, tx pgx.Tx, id string, status models.TournamentStatus) error
	UpdateBracketSizeTx(ctx context.Context, tx pgx.Tx, id string, bracketSize int) error
}

type tournamentRepository struct {
	db *pgxpool.Pool
}

func NewTournamentRepository(db *pgxpool.Pool) TournamentRepository {
	return &tournamentRepository{db: db}
}

const tournamentSelectCols = `
	t.id, t.name, t.type, t.status, t.created_by, t.registration_link_token,
	t.created_at, t.updated_at, COALESCE(s.bracket_size, 0)
`

func scanTournament(row pgx.Row) (*models.Tournament, error) {
	var t models.Tournament
	err := row.Scan(
		&t.ID, &t.Name, &t.Type, &t.Status, &t.CreatedBy, &t.RegistrationLinkToken,
		&t.CreatedAt, &t.UpdatedAt, &t.BracketSize,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tournamentRepository) CreateSpinWheel(ctx context.Context, name string, createdBy string, bracketSize int) (*models.Tournament, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx,
		`INSERT INTO tournaments (name, type, created_by)
		 VALUES ($1, 'spin_wheel', $2)
		 RETURNING id, name, type, status, created_by, registration_link_token, created_at, updated_at`,
		name, createdBy,
	)

	var t models.Tournament
	if err := row.Scan(&t.ID, &t.Name, &t.Type, &t.Status, &t.CreatedBy, &t.RegistrationLinkToken, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO spin_wheel_settings (tournament_id, bracket_size) VALUES ($1, $2)`,
		t.ID, bracketSize,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	t.BracketSize = bracketSize
	return &t, nil
}

func (r *tournamentRepository) FindByID(ctx context.Context, id string) (*models.Tournament, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+tournamentSelectCols+`
		 FROM tournaments t
		 LEFT JOIN spin_wheel_settings s ON s.tournament_id = t.id
		 WHERE t.id = $1`,
		id,
	)
	t, err := scanTournament(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTournamentNotFound
	}
	return t, err
}

func (r *tournamentRepository) FindByRegistrationToken(ctx context.Context, token string) (*models.Tournament, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+tournamentSelectCols+`
		 FROM tournaments t
		 LEFT JOIN spin_wheel_settings s ON s.tournament_id = t.id
		 WHERE t.registration_link_token = $1`,
		token,
	)
	t, err := scanTournament(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTournamentNotFound
	}
	return t, err
}

func (r *tournamentRepository) ListByCreator(ctx context.Context, createdBy string) ([]models.Tournament, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+tournamentSelectCols+`
		 FROM tournaments t
		 LEFT JOIN spin_wheel_settings s ON s.tournament_id = t.id
		 WHERE t.created_by = $1
		 ORDER BY t.created_at DESC`,
		createdBy,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Tournament
	for rows.Next() {
		t, err := scanTournament(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *t)
	}
	return list, rows.Err()
}

func (r *tournamentRepository) UpdateStatus(ctx context.Context, id string, status models.TournamentStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tournaments SET status = $1, updated_at = now() WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *tournamentRepository) UpdateBracketSize(ctx context.Context, id string, bracketSize int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE spin_wheel_settings SET bracket_size = $1 WHERE tournament_id = $2`,
		bracketSize, id,
	)
	return err
}

func (r *tournamentRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *tournamentRepository) FindByIDForUpdateTx(ctx context.Context, tx pgx.Tx, id string) (*models.Tournament, error) {
	row := tx.QueryRow(ctx,
		`SELECT `+tournamentSelectCols+`
		 FROM tournaments t
		 LEFT JOIN spin_wheel_settings s ON s.tournament_id = t.id
		 WHERE t.id = $1
		 FOR UPDATE OF t`,
		id,
	)
	t, err := scanTournament(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTournamentNotFound
	}
	return t, err
}

func (r *tournamentRepository) UpdateStatusTx(ctx context.Context, tx pgx.Tx, id string, status models.TournamentStatus) error {
	_, err := tx.Exec(ctx,
		`UPDATE tournaments SET status = $1, updated_at = now() WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *tournamentRepository) UpdateBracketSizeTx(ctx context.Context, tx pgx.Tx, id string, bracketSize int) error {
	_, err := tx.Exec(ctx,
		`UPDATE spin_wheel_settings SET bracket_size = $1 WHERE tournament_id = $2`,
		bracketSize, id,
	)
	return err
}

func (r *tournamentRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM tournaments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTournamentNotFound
	}
	return nil
}
