package models

import "time"

type EntrySource string

const (
	EntrySourceForm   EntrySource = "form"
	EntrySourceManual EntrySource = "manual"
)

type SpinWheelEntry struct {
	ID           string      `json:"id"`
	TournamentID string      `json:"tournament_id"`
	PlayerName   string      `json:"player_name"`
	Phone        *string     `json:"phone,omitempty"`
	Source       EntrySource `json:"source"`
	AddedBy      *string     `json:"added_by,omitempty"`
	ColorHex     *string     `json:"color_hex,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

type MatchStatus string

const (
	MatchPending   MatchStatus = "pending"
	MatchCompleted MatchStatus = "completed"
)

type SpinWheelMatch struct {
	ID             string      `json:"id"`
	TournamentID   string      `json:"tournament_id"`
	Round          int         `json:"round"`
	MatchNumber    int         `json:"match_number"`
	Player1EntryID *string     `json:"player1_entry_id"`
	Player2EntryID *string     `json:"player2_entry_id"`
	Player1Score   *int        `json:"player1_score"`
	Player2Score   *int        `json:"player2_score"`
	WinnerEntryID  *string     `json:"winner_entry_id"`
	Status         MatchStatus `json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
}
