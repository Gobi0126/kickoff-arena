package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avanthika/efootball-backend/internal/models"
)

var ErrDuplicateColor = errors.New("color already assigned in this tournament")
var ErrEntryNotFound = errors.New("player not found")

type SpinWheelRepository interface {
	CreateEntry(ctx context.Context, tournamentID, name string, phone *string, source models.EntrySource, addedBy *string) (*models.SpinWheelEntry, error)
	ListEntries(ctx context.Context, tournamentID string) ([]models.SpinWheelEntry, error)
	CountEntries(ctx context.Context, tournamentID string) (int, error)

	// Tx variants run within a transaction that already holds a row lock on
	// the parent tournament (see TournamentRepository.FindByIDForUpdateTx),
	// so the count-check-then-insert sequence is atomic across concurrent
	// requests for the same tournament.
	CountEntriesTx(ctx context.Context, tx pgx.Tx, tournamentID string) (int, error)
	CreateEntryTx(ctx context.Context, tx pgx.Tx, tournamentID, name string, phone *string, source models.EntrySource, addedBy *string) (*models.SpinWheelEntry, error)
	AssignColor(ctx context.Context, entryID, colorHex string) error
	FindEntry(ctx context.Context, entryID string) (*models.SpinWheelEntry, error)
	UpdateEntry(ctx context.Context, entryID, name string, phone *string) (*models.SpinWheelEntry, error)
	DeleteEntry(ctx context.Context, entryID string) error

	CreateMatch(ctx context.Context, tournamentID string, round, matchNumber int, player1ID, player2ID *string) (*models.SpinWheelMatch, error)
	CreateByeMatch(ctx context.Context, tournamentID string, round, matchNumber int, playerID string) (*models.SpinWheelMatch, error)
	ListMatches(ctx context.Context, tournamentID string) ([]models.SpinWheelMatch, error)
	FindMatch(ctx context.Context, matchID string) (*models.SpinWheelMatch, error)
	FindMatchByPosition(ctx context.Context, tournamentID string, round, matchNumber int) (*models.SpinWheelMatch, error)
	SubmitResult(ctx context.Context, matchID string, p1Score, p2Score int, winnerEntryID string) error
	MaxRound(ctx context.Context, tournamentID string) (int, error)
}

type spinWheelRepository struct {
	db *pgxpool.Pool
}

func NewSpinWheelRepository(db *pgxpool.Pool) SpinWheelRepository {
	return &spinWheelRepository{db: db}
}

func (r *spinWheelRepository) CreateEntry(ctx context.Context, tournamentID, name string, phone *string, source models.EntrySource, addedBy *string) (*models.SpinWheelEntry, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO spin_wheel_entries (tournament_id, player_name, phone, source, added_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, tournament_id, player_name, phone, source, added_by, color_hex, created_at`,
		tournamentID, name, phone, source, addedBy,
	)
	var e models.SpinWheelEntry
	err := row.Scan(&e.ID, &e.TournamentID, &e.PlayerName, &e.Phone, &e.Source, &e.AddedBy, &e.ColorHex, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *spinWheelRepository) ListEntries(ctx context.Context, tournamentID string) ([]models.SpinWheelEntry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, tournament_id, player_name, phone, source, added_by, color_hex, created_at
		 FROM spin_wheel_entries WHERE tournament_id = $1 ORDER BY created_at ASC`,
		tournamentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.SpinWheelEntry
	for rows.Next() {
		var e models.SpinWheelEntry
		if err := rows.Scan(&e.ID, &e.TournamentID, &e.PlayerName, &e.Phone, &e.Source, &e.AddedBy, &e.ColorHex, &e.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *spinWheelRepository) CountEntries(ctx context.Context, tournamentID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM spin_wheel_entries WHERE tournament_id = $1`, tournamentID).Scan(&count)
	return count, err
}

func (r *spinWheelRepository) CountEntriesTx(ctx context.Context, tx pgx.Tx, tournamentID string) (int, error) {
	var count int
	err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM spin_wheel_entries WHERE tournament_id = $1`, tournamentID).Scan(&count)
	return count, err
}

func (r *spinWheelRepository) CreateEntryTx(ctx context.Context, tx pgx.Tx, tournamentID, name string, phone *string, source models.EntrySource, addedBy *string) (*models.SpinWheelEntry, error) {
	row := tx.QueryRow(ctx,
		`INSERT INTO spin_wheel_entries (tournament_id, player_name, phone, source, added_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, tournament_id, player_name, phone, source, added_by, color_hex, created_at`,
		tournamentID, name, phone, source, addedBy,
	)
	var e models.SpinWheelEntry
	err := row.Scan(&e.ID, &e.TournamentID, &e.PlayerName, &e.Phone, &e.Source, &e.AddedBy, &e.ColorHex, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *spinWheelRepository) AssignColor(ctx context.Context, entryID, colorHex string) error {
	_, err := r.db.Exec(ctx, `UPDATE spin_wheel_entries SET color_hex = $1 WHERE id = $2`, colorHex, entryID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateColor
		}
		return err
	}
	return nil
}

func (r *spinWheelRepository) FindEntry(ctx context.Context, entryID string) (*models.SpinWheelEntry, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, tournament_id, player_name, phone, source, added_by, color_hex, created_at
		 FROM spin_wheel_entries WHERE id = $1`,
		entryID,
	)
	var e models.SpinWheelEntry
	err := row.Scan(&e.ID, &e.TournamentID, &e.PlayerName, &e.Phone, &e.Source, &e.AddedBy, &e.ColorHex, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEntryNotFound
	}
	return &e, err
}

func (r *spinWheelRepository) UpdateEntry(ctx context.Context, entryID, name string, phone *string) (*models.SpinWheelEntry, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE spin_wheel_entries SET player_name = $1, phone = $2 WHERE id = $3
		 RETURNING id, tournament_id, player_name, phone, source, added_by, color_hex, created_at`,
		name, phone, entryID,
	)
	var e models.SpinWheelEntry
	err := row.Scan(&e.ID, &e.TournamentID, &e.PlayerName, &e.Phone, &e.Source, &e.AddedBy, &e.ColorHex, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEntryNotFound
	}
	return &e, err
}

func (r *spinWheelRepository) DeleteEntry(ctx context.Context, entryID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM spin_wheel_entries WHERE id = $1`, entryID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrEntryNotFound
	}
	return nil
}

func (r *spinWheelRepository) CreateMatch(ctx context.Context, tournamentID string, round, matchNumber int, player1ID, player2ID *string) (*models.SpinWheelMatch, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO spin_wheel_matches (tournament_id, round, match_number, player1_entry_id, player2_entry_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, tournament_id, round, match_number, player1_entry_id, player2_entry_id,
		           player1_score, player2_score, winner_entry_id, status, created_at`,
		tournamentID, round, matchNumber, player1ID, player2ID,
	)
	return scanMatch(row)
}

// CreateByeMatch creates a match that is already decided — used when a
// player has no opponent because the entrant count isn't a power of two.
// They advance automatically without playing.
func (r *spinWheelRepository) CreateByeMatch(ctx context.Context, tournamentID string, round, matchNumber int, playerID string) (*models.SpinWheelMatch, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO spin_wheel_matches (tournament_id, round, match_number, player1_entry_id, winner_entry_id, status)
		 VALUES ($1, $2, $3, $4, $4, 'completed')
		 RETURNING id, tournament_id, round, match_number, player1_entry_id, player2_entry_id,
		           player1_score, player2_score, winner_entry_id, status, created_at`,
		tournamentID, round, matchNumber, playerID,
	)
	return scanMatch(row)
}

func scanMatch(row pgx.Row) (*models.SpinWheelMatch, error) {
	var m models.SpinWheelMatch
	err := row.Scan(
		&m.ID, &m.TournamentID, &m.Round, &m.MatchNumber, &m.Player1EntryID, &m.Player2EntryID,
		&m.Player1Score, &m.Player2Score, &m.WinnerEntryID, &m.Status, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *spinWheelRepository) ListMatches(ctx context.Context, tournamentID string) ([]models.SpinWheelMatch, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, tournament_id, round, match_number, player1_entry_id, player2_entry_id,
		        player1_score, player2_score, winner_entry_id, status, created_at
		 FROM spin_wheel_matches WHERE tournament_id = $1 ORDER BY round ASC, match_number ASC`,
		tournamentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.SpinWheelMatch
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *m)
	}
	return list, rows.Err()
}

var ErrMatchNotFound = errors.New("match not found")

func (r *spinWheelRepository) FindMatch(ctx context.Context, matchID string) (*models.SpinWheelMatch, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, tournament_id, round, match_number, player1_entry_id, player2_entry_id,
		        player1_score, player2_score, winner_entry_id, status, created_at
		 FROM spin_wheel_matches WHERE id = $1`,
		matchID,
	)
	m, err := scanMatch(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMatchNotFound
	}
	return m, err
}

func (r *spinWheelRepository) FindMatchByPosition(ctx context.Context, tournamentID string, round, matchNumber int) (*models.SpinWheelMatch, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, tournament_id, round, match_number, player1_entry_id, player2_entry_id,
		        player1_score, player2_score, winner_entry_id, status, created_at
		 FROM spin_wheel_matches WHERE tournament_id = $1 AND round = $2 AND match_number = $3`,
		tournamentID, round, matchNumber,
	)
	m, err := scanMatch(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMatchNotFound
	}
	return m, err
}

func (r *spinWheelRepository) SubmitResult(ctx context.Context, matchID string, p1Score, p2Score int, winnerEntryID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE spin_wheel_matches
		 SET player1_score = $1, player2_score = $2, winner_entry_id = $3, status = 'completed'
		 WHERE id = $4`,
		p1Score, p2Score, winnerEntryID, matchID,
	)
	return err
}

func (r *spinWheelRepository) MaxRound(ctx context.Context, tournamentID string) (int, error) {
	var maxRound *int
	err := r.db.QueryRow(ctx,
		`SELECT MAX(round) FROM spin_wheel_matches WHERE tournament_id = $1`,
		tournamentID,
	).Scan(&maxRound)
	if err != nil {
		return 0, err
	}
	if maxRound == nil {
		return 0, nil
	}
	return *maxRound, nil
}
