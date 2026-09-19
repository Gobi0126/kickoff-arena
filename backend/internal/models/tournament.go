package models

import "time"

type TournamentType string
type TournamentStatus string

const (
	TournamentTypeSpinWheel TournamentType = "spin_wheel"
	TournamentTypeAuction   TournamentType = "auction"

	StatusRegistrationOpen   TournamentStatus = "registration_open"
	StatusRegistrationClosed TournamentStatus = "registration_closed"
	StatusInProgress         TournamentStatus = "in_progress"
	StatusCompleted          TournamentStatus = "completed"
)

type Tournament struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	Type                   TournamentType   `json:"type"`
	Status                 TournamentStatus `json:"status"`
	CreatedBy              string           `json:"created_by"`
	RegistrationLinkToken  string           `json:"registration_link_token"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
	BracketSize            int              `json:"bracket_size,omitempty"`
}
