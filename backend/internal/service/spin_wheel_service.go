package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"

	"github.com/avanthika/efootball-backend/internal/models"
	"github.com/avanthika/efootball-backend/internal/repository"
)

var (
	ErrNotYourTournament    = errors.New("you do not own this tournament")
	ErrWrongTournamentType  = errors.New("tournament is not a spin wheel tournament")
	ErrRegistrationClosed   = errors.New("registration is closed for this tournament")
	ErrBracketNotFull       = errors.New("bracket is not full yet")
	ErrAlreadyStarted       = errors.New("tournament has already started")
	ErrTiedScore            = errors.New("scores cannot be tied in a knockout match")
	ErrMatchAlreadyComplete = errors.New("this match result has already been submitted")
	ErrMatchNotReady        = errors.New("one or both players for this match are not decided yet")
	ErrInvalidBracketSize   = errors.New("player count must be 2, 4, 8, 16, 32, or 64")
	ErrInvalidOrder         = errors.New("order must contain every registered player exactly once")
	ErrBracketSizeTooSmall  = errors.New("player count cannot be less than the number of players already registered")
)

type SpinWheelService interface {
	CreateTournament(ctx context.Context, createdBy, name string, bracketSize int) (*models.Tournament, error)
	ListMyTournaments(ctx context.Context, createdBy string) ([]models.Tournament, error)
	GetTournament(ctx context.Context, id, requesterID string) (*models.Tournament, error)
	GetPublicTournament(ctx context.Context, token string) (*models.Tournament, error)
	UpdateBracketSize(ctx context.Context, tournamentID, requesterID string, bracketSize int) (*models.Tournament, error)
	DeleteTournament(ctx context.Context, tournamentID, requesterID string) error

	Register(ctx context.Context, token, name, phone string) (*models.SpinWheelEntry, error)
	AddManualEntry(ctx context.Context, tournamentID, requesterID, name, phone string) (*models.SpinWheelEntry, error)
	ListEntries(ctx context.Context, tournamentID, requesterID string) ([]models.SpinWheelEntry, error)
	UpdateEntry(ctx context.Context, tournamentID, entryID, requesterID, name, phone string) (*models.SpinWheelEntry, error)
	DeleteEntry(ctx context.Context, tournamentID, entryID, requesterID string) error

	AssignColors(ctx context.Context, tournamentID, requesterID string) ([]models.SpinWheelEntry, error)
	StartSpin(ctx context.Context, tournamentID, requesterID string, order []string) ([]models.SpinWheelMatch, error)
	GetFixtures(ctx context.Context, tournamentID, requesterID string) ([]models.SpinWheelMatch, error)
	SubmitMatchResult(ctx context.Context, matchID, requesterID string, p1Score, p2Score int) (*models.SpinWheelMatch, error)
}

type spinWheelService struct {
	tournaments repository.TournamentRepository
	spin        repository.SpinWheelRepository
}

func NewSpinWheelService(tournaments repository.TournamentRepository, spin repository.SpinWheelRepository) SpinWheelService {
	return &spinWheelService{tournaments: tournaments, spin: spin}
}

const (
	minBracketSize = 2
	maxBracketSize = 64
)

// isValidBracketSize requires an exact power of two (2, 4, 8, 16, 32, 64) so
// a single-elimination bracket never needs byes — every registered player
// always has a real round-1 opponent.
func isValidBracketSize(n int) bool {
	if n < minBracketSize || n > maxBracketSize {
		return false
	}
	return n&(n-1) == 0
}

func (s *spinWheelService) CreateTournament(ctx context.Context, createdBy, name string, bracketSize int) (*models.Tournament, error) {
	if !isValidBracketSize(bracketSize) {
		return nil, ErrInvalidBracketSize
	}
	return s.tournaments.CreateSpinWheel(ctx, name, createdBy, bracketSize)
}

func (s *spinWheelService) ListMyTournaments(ctx context.Context, createdBy string) ([]models.Tournament, error) {
	return s.tournaments.ListByCreator(ctx, createdBy)
}

func (s *spinWheelService) GetTournament(ctx context.Context, id, requesterID string) (*models.Tournament, error) {
	t, err := s.tournaments.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.CreatedBy != requesterID {
		return nil, ErrNotYourTournament
	}
	return t, nil
}

func (s *spinWheelService) GetPublicTournament(ctx context.Context, token string) (*models.Tournament, error) {
	return s.tournaments.FindByRegistrationToken(ctx, token)
}

func (s *spinWheelService) UpdateBracketSize(ctx context.Context, tournamentID, requesterID string, bracketSize int) (*models.Tournament, error) {
	t, err := s.GetTournament(ctx, tournamentID, requesterID)
	if err != nil {
		return nil, err
	}
	if t.Status != models.StatusRegistrationOpen {
		return nil, ErrRegistrationClosed
	}
	if !isValidBracketSize(bracketSize) {
		return nil, ErrInvalidBracketSize
	}

	tx, err := s.tournaments.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Lock the tournament row so this read-count-then-write can't race
	// against a concurrent registration landing in between.
	locked, err := s.tournaments.FindByIDForUpdateTx(ctx, tx, tournamentID)
	if err != nil {
		return nil, err
	}
	if locked.Status != models.StatusRegistrationOpen {
		return nil, ErrRegistrationClosed
	}

	count, err := s.spin.CountEntriesTx(ctx, tx, tournamentID)
	if err != nil {
		return nil, err
	}
	if bracketSize < count {
		return nil, ErrBracketSizeTooSmall
	}

	if err := s.tournaments.UpdateBracketSizeTx(ctx, tx, tournamentID, bracketSize); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.tournaments.FindByID(ctx, tournamentID)
}

func (s *spinWheelService) DeleteTournament(ctx context.Context, tournamentID, requesterID string) error {
	if _, err := s.GetTournament(ctx, tournamentID, requesterID); err != nil {
		return err
	}
	return s.tournaments.Delete(ctx, tournamentID)
}

func (s *spinWheelService) Register(ctx context.Context, token, name, phone string) (*models.SpinWheelEntry, error) {
	t0, err := s.tournaments.FindByRegistrationToken(ctx, token)
	if err != nil {
		return nil, err
	}

	tx, err := s.tournaments.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Lock the tournament row so two submissions racing for the last open
	// slot serialize instead of both reading "still room" and both landing.
	t, err := s.tournaments.FindByIDForUpdateTx(ctx, tx, t0.ID)
	if err != nil {
		return nil, err
	}
	if t.Status != models.StatusRegistrationOpen {
		return nil, ErrRegistrationClosed
	}

	count, err := s.spin.CountEntriesTx(ctx, tx, t.ID)
	if err != nil {
		return nil, err
	}
	if count >= t.BracketSize {
		return nil, ErrRegistrationClosed
	}

	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}
	entry, err := s.spin.CreateEntryTx(ctx, tx, t.ID, name, phonePtr, models.EntrySourceForm, nil)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *spinWheelService) AddManualEntry(ctx context.Context, tournamentID, requesterID, name, phone string) (*models.SpinWheelEntry, error) {
	tx, err := s.tournaments.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	t, err := s.tournaments.FindByIDForUpdateTx(ctx, tx, tournamentID)
	if err != nil {
		return nil, err
	}
	if t.CreatedBy != requesterID {
		return nil, ErrNotYourTournament
	}
	if t.Status != models.StatusRegistrationOpen {
		return nil, ErrRegistrationClosed
	}

	count, err := s.spin.CountEntriesTx(ctx, tx, t.ID)
	if err != nil {
		return nil, err
	}
	if count >= t.BracketSize {
		return nil, ErrRegistrationClosed
	}

	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}
	entry, err := s.spin.CreateEntryTx(ctx, tx, t.ID, name, phonePtr, models.EntrySourceManual, &requesterID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *spinWheelService) ListEntries(ctx context.Context, tournamentID, requesterID string) ([]models.SpinWheelEntry, error) {
	if _, err := s.GetTournament(ctx, tournamentID, requesterID); err != nil {
		return nil, err
	}
	return s.spin.ListEntries(ctx, tournamentID)
}

func (s *spinWheelService) UpdateEntry(ctx context.Context, tournamentID, entryID, requesterID, name, phone string) (*models.SpinWheelEntry, error) {
	t, err := s.GetTournament(ctx, tournamentID, requesterID)
	if err != nil {
		return nil, err
	}
	if t.Status != models.StatusRegistrationOpen {
		return nil, ErrRegistrationClosed
	}

	entry, err := s.spin.FindEntry(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if entry.TournamentID != tournamentID {
		return nil, repository.ErrEntryNotFound
	}

	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}
	return s.spin.UpdateEntry(ctx, entryID, name, phonePtr)
}

func (s *spinWheelService) DeleteEntry(ctx context.Context, tournamentID, entryID, requesterID string) error {
	t, err := s.GetTournament(ctx, tournamentID, requesterID)
	if err != nil {
		return err
	}
	if t.Status != models.StatusRegistrationOpen {
		return ErrRegistrationClosed
	}

	entry, err := s.spin.FindEntry(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.TournamentID != tournamentID {
		return repository.ErrEntryNotFound
	}

	return s.spin.DeleteEntry(ctx, entryID)
}

// distinctColors generates n colors evenly spaced around the hue wheel, so
// visual separation between any two colors stays maximal no matter how many
// are needed — unlike picking a random subset of a fixed palette, which can
// land on two similar shades (e.g. two purples) for small n. A random hue
// offset keeps the set from looking identical across every spin.
func distinctColors(n int) []string {
	if n == 0 {
		return nil
	}
	offset := rand.Float64() * 360
	colors := make([]string, n)
	for i := 0; i < n; i++ {
		hue := math.Mod(offset+float64(i)*(360.0/float64(n)), 360)
		colors[i] = hslToHex(hue, 0.65, 0.48)
	}
	return colors
}

func hslToHex(h, s, l float64) string {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	toHex := func(v float64) int {
		return int(math.Round((v + m) * 255))
	}
	return fmt.Sprintf("#%02X%02X%02X", toHex(r), toHex(g), toHex(b))
}

// AssignColors gives every registered entry a unique, evenly-spaced color.
// Called when the subadmin opens the spin wheel screen, before any spinning
// happens, so the wheel can render each player's slice with its final color
// from the very first frame.
func (s *spinWheelService) AssignColors(ctx context.Context, tournamentID, requesterID string) ([]models.SpinWheelEntry, error) {
	t, err := s.GetTournament(ctx, tournamentID, requesterID)
	if err != nil {
		return nil, err
	}
	if t.Status != models.StatusRegistrationOpen {
		return nil, ErrAlreadyStarted
	}

	entries, err := s.spin.ListEntries(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if len(entries) != t.BracketSize {
		return nil, ErrBracketNotFull
	}

	palette := distinctColors(len(entries))
	rand.Shuffle(len(palette), func(i, j int) {
		palette[i], palette[j] = palette[j], palette[i]
	})

	for i, entry := range entries {
		if err := s.spin.AssignColor(ctx, entry.ID, palette[i]); err != nil {
			return nil, err
		}
	}

	return s.spin.ListEntries(ctx, tournamentID)
}

// StartSpin creates round 1. If order is provided (the sequence the subadmin
// picked players off the wheel, one click at a time), that exact order
// decides byes and pairings. If order is nil, players are paired randomly —
// used as a fallback for a "quick start" without the wheel animation.
func (s *spinWheelService) StartSpin(ctx context.Context, tournamentID, requesterID string, order []string) ([]models.SpinWheelMatch, error) {
	tx, err := s.tournaments.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	// Lock the tournament row so two simultaneous "Start Spin" clicks can't
	// both pass the status check before either one flips it — the second
	// blocks here until the first commits, then sees status = in_progress.
	t, err := s.tournaments.FindByIDForUpdateTx(ctx, tx, tournamentID)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}
	if t.CreatedBy != requesterID {
		tx.Rollback(ctx)
		return nil, ErrNotYourTournament
	}
	if t.Status != models.StatusRegistrationOpen {
		tx.Rollback(ctx)
		return nil, ErrAlreadyStarted
	}

	entries, err := s.spin.ListEntries(ctx, tournamentID)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}
	if len(entries) != t.BracketSize {
		tx.Rollback(ctx)
		return nil, ErrBracketNotFull
	}

	if err := s.tournaments.UpdateStatusTx(ctx, tx, tournamentID, models.StatusInProgress); err != nil {
		tx.Rollback(ctx)
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Status is now safely flipped to in_progress and committed — no other
	// call can get past the check above for this tournament anymore, so the
	// rest of the bracket setup can proceed with the regular (non-locking)
	// repository calls.
	shuffled, err := resolveOrder(entries, order)
	if err != nil {
		return nil, err
	}

	// If entries haven't had colors assigned yet (e.g. this is the random
	// "quick start" fallback path), assign them now.
	if shuffled[0].ColorHex == nil {
		palette := distinctColors(len(shuffled))
		for i, entry := range shuffled {
			if err := s.spin.AssignColor(ctx, entry.ID, palette[i]); err != nil {
				return nil, err
			}
		}
	}

	// If the entrant count isn't a power of two, the extra slots up to the
	// next power of two are filled with "byes" — those players advance to
	// round 2 automatically without playing a round 1 match. This is the
	// standard way real single-elimination brackets handle odd sizes.
	fullBracket := nextPowerOfTwo(len(shuffled))
	numByes := fullBracket - len(shuffled)
	totalSlots := fullBracket / 2

	// Spread the bye slots evenly across round 1 instead of bunching them
	// all at the front. Two adjacent match slots (2k-1, 2k) become each
	// other's round-2 opponent (see tryAdvanceRound), so bunching byes
	// together meant two players who never played round 1 would end up
	// facing each other in round 2 anyway — effectively wasting a round
	// for them. Spacing byes out means each one is paired against an
	// actual round-1 winner next round whenever that's mathematically
	// possible (i.e. as long as byes are at most half of the slots).
	isByeSlot := make([]bool, totalSlots)
	for i := 0; i < totalSlots; i++ {
		if (i+1)*numByes/totalSlots > i*numByes/totalSlots {
			isByeSlot[i] = true
		}
	}

	var matches []models.SpinWheelMatch
	idx := 0

	for slot := 0; slot < totalSlots; slot++ {
		matchNumber := slot + 1
		if isByeSlot[slot] {
			playerID := shuffled[idx].ID
			idx++
			m, err := s.spin.CreateByeMatch(ctx, tournamentID, 1, matchNumber, playerID)
			if err != nil {
				return nil, err
			}
			matches = append(matches, *m)
			if err := s.tryAdvanceRound(ctx, tournamentID, 1, matchNumber, playerID); err != nil {
				return nil, err
			}
			continue
		}

		p1 := shuffled[idx].ID
		p2 := shuffled[idx+1].ID
		idx += 2
		m, err := s.spin.CreateMatch(ctx, tournamentID, 1, matchNumber, &p1, &p2)
		if err != nil {
			return nil, err
		}
		matches = append(matches, *m)
	}

	return matches, nil
}

func nextPowerOfTwo(n int) int {
	p := 1
	for p < n {
		p *= 2
	}
	return p
}

// resolveOrder returns entries arranged in the requested order. If order is
// empty, entries are shuffled randomly instead. order must be exactly a
// permutation of the entry IDs — no duplicates, nothing missing, nothing
// foreign to this tournament.
func resolveOrder(entries []models.SpinWheelEntry, order []string) ([]models.SpinWheelEntry, error) {
	if len(order) == 0 {
		shuffled := make([]models.SpinWheelEntry, len(entries))
		copy(shuffled, entries)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		return shuffled, nil
	}

	if len(order) != len(entries) {
		return nil, ErrInvalidOrder
	}

	byID := make(map[string]models.SpinWheelEntry, len(entries))
	for _, e := range entries {
		byID[e.ID] = e
	}

	seen := make(map[string]bool, len(order))
	resolved := make([]models.SpinWheelEntry, len(order))
	for i, id := range order {
		entry, ok := byID[id]
		if !ok || seen[id] {
			return nil, ErrInvalidOrder
		}
		seen[id] = true
		resolved[i] = entry
	}

	return resolved, nil
}

func (s *spinWheelService) GetFixtures(ctx context.Context, tournamentID, requesterID string) ([]models.SpinWheelMatch, error) {
	if _, err := s.GetTournament(ctx, tournamentID, requesterID); err != nil {
		return nil, err
	}
	return s.spin.ListMatches(ctx, tournamentID)
}

func (s *spinWheelService) SubmitMatchResult(ctx context.Context, matchID, requesterID string, p1Score, p2Score int) (*models.SpinWheelMatch, error) {
	match, err := s.spin.FindMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	t, err := s.GetTournament(ctx, match.TournamentID, requesterID)
	if err != nil {
		return nil, err
	}

	if match.Status == models.MatchCompleted {
		return nil, ErrMatchAlreadyComplete
	}
	if match.Player1EntryID == nil || match.Player2EntryID == nil {
		return nil, ErrMatchNotReady
	}
	if p1Score == p2Score {
		return nil, ErrTiedScore
	}

	winnerID := *match.Player1EntryID
	if p2Score > p1Score {
		winnerID = *match.Player2EntryID
	}

	if err := s.spin.SubmitResult(ctx, matchID, p1Score, p2Score, winnerID); err != nil {
		return nil, err
	}

	matchesThisRound := nextPowerOfTwo(t.BracketSize) / (1 << match.Round)
	if matchesThisRound <= 1 {
		// This was the final.
		if err := s.tournaments.UpdateStatus(ctx, match.TournamentID, models.StatusCompleted); err != nil {
			return nil, err
		}
	} else {
		if err := s.tryAdvanceRound(ctx, t.ID, match.Round, match.MatchNumber, winnerID); err != nil {
			return nil, err
		}
	}

	return s.spin.FindMatch(ctx, matchID)
}

// tryAdvanceRound checks whether the sibling match (the one this match is
// paired with for the next round) is also complete. If so, it creates the
// next round's match with both winners. If not, it just waits — the sibling
// match's own completion will trigger creation instead.
func (s *spinWheelService) tryAdvanceRound(ctx context.Context, tournamentID string, round, matchNumber int, winnerID string) error {
	var siblingNumber int
	if matchNumber%2 == 1 {
		siblingNumber = matchNumber + 1
	} else {
		siblingNumber = matchNumber - 1
	}

	sibling, err := s.spin.FindMatchByPosition(ctx, tournamentID, round, siblingNumber)
	if errors.Is(err, repository.ErrMatchNotFound) || sibling.Status != models.MatchCompleted {
		// Sibling not done yet — nothing to do until it finishes.
		return nil
	}
	if err != nil {
		return err
	}

	nextRound := round + 1
	nextMatchNumber := (matchNumber + 1) / 2

	// Idempotency guard: if this next-round match already exists (e.g. a
	// retry), don't create it twice.
	if _, err := s.spin.FindMatchByPosition(ctx, tournamentID, nextRound, nextMatchNumber); err == nil {
		return nil
	} else if !errors.Is(err, repository.ErrMatchNotFound) {
		return err
	}

	p1 := winnerID
	p2 := *sibling.WinnerEntryID
	// Keep a stable order: whichever match_number is lower feeds player1.
	if matchNumber > siblingNumber {
		p1, p2 = p2, p1
	}

	_, err = s.spin.CreateMatch(ctx, tournamentID, nextRound, nextMatchNumber, &p1, &p2)
	return err
}
