import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';
import { SpinWheelEntry, SpinWheelMatch, Tournament } from '../models/tournament.model';

@Injectable({ providedIn: 'root' })
export class SpinWheelService {
  private http = inject(HttpClient);
  private base = `${environment.apiUrl}`;

  create(name: string, bracketSize: number) {
    return this.http.post<Tournament>(`${this.base}/tournaments`, {
      name,
      bracket_size: bracketSize,
    });
  }

  listMine() {
    return this.http.get<{ tournaments: Tournament[] }>(`${this.base}/tournaments`);
  }

  get(id: string) {
    return this.http.get<Tournament>(`${this.base}/tournaments/${id}`);
  }

  updateBracketSize(id: string, bracketSize: number) {
    return this.http.patch<Tournament>(`${this.base}/tournaments/${id}`, { bracket_size: bracketSize });
  }

  delete(id: string) {
    return this.http.delete<void>(`${this.base}/tournaments/${id}`);
  }

  listEntries(id: string) {
    return this.http.get<{ entries: SpinWheelEntry[] }>(`${this.base}/tournaments/${id}/entries`);
  }

  addManualEntry(id: string, name: string, phone: string) {
    return this.http.post<SpinWheelEntry>(`${this.base}/tournaments/${id}/entries`, { name, phone });
  }

  updateEntry(tournamentId: string, entryId: string, name: string, phone: string) {
    return this.http.patch<SpinWheelEntry>(
      `${this.base}/tournaments/${tournamentId}/entries/${entryId}`,
      { name, phone }
    );
  }

  deleteEntry(tournamentId: string, entryId: string) {
    return this.http.delete<void>(`${this.base}/tournaments/${tournamentId}/entries/${entryId}`);
  }

  assignColors(id: string) {
    return this.http.post<{ entries: SpinWheelEntry[] }>(`${this.base}/tournaments/${id}/assign-colors`, {});
  }

  startSpin(id: string, order?: string[]) {
    return this.http.post<{ matches: SpinWheelMatch[] }>(`${this.base}/tournaments/${id}/spin`, {
      order: order ?? [],
    });
  }

  getFixtures(id: string) {
    return this.http.get<{ matches: SpinWheelMatch[] }>(`${this.base}/tournaments/${id}/fixtures`);
  }

  submitResult(tournamentId: string, matchId: string, player1Score: number, player2Score: number) {
    return this.http.patch<SpinWheelMatch>(
      `${this.base}/tournaments/${tournamentId}/matches/${matchId}`,
      { player1_score: player1Score, player2_score: player2Score }
    );
  }

  // Public — no auth needed.
  getPublicTournament(token: string) {
    return this.http.get<Tournament>(`${this.base}/register/${token}`);
  }

  register(token: string, name: string, phone: string) {
    return this.http.post<SpinWheelEntry>(`${this.base}/register/${token}`, { name, phone });
  }
}
