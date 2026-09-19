export type TournamentStatus =
  | 'registration_open'
  | 'registration_closed'
  | 'in_progress'
  | 'completed';

export interface Tournament {
  id: string;
  name: string;
  type: 'spin_wheel' | 'auction';
  status: TournamentStatus;
  created_by: string;
  registration_link_token: string;
  created_at: string;
  updated_at: string;
  bracket_size: number;
}

export interface SpinWheelEntry {
  id: string;
  tournament_id: string;
  player_name: string;
  phone?: string;
  source: 'form' | 'manual';
  added_by?: string;
  color_hex?: string;
  created_at: string;
}

export type MatchStatus = 'pending' | 'completed';

export interface SpinWheelMatch {
  id: string;
  tournament_id: string;
  round: number;
  match_number: number;
  player1_entry_id: string | null;
  player2_entry_id: string | null;
  player1_score: number | null;
  player2_score: number | null;
  winner_entry_id: string | null;
  status: MatchStatus;
  created_at: string;
}
