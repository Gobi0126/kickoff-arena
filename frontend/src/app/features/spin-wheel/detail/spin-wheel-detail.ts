import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { SpinWheelService } from '../../../core/services/spin-wheel.service';
import { ConfirmService } from '../../../core/services/confirm.service';
import { SpinWheelEntry, SpinWheelMatch, Tournament } from '../../../core/models/tournament.model';

@Component({
  selector: 'app-spin-wheel-detail',
  imports: [FormsModule, RouterLink],
  templateUrl: './spin-wheel-detail.html',
  styleUrl: './spin-wheel-detail.scss',
})
export class SpinWheelDetail implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private service = inject(SpinWheelService);
  private confirmService = inject(ConfirmService);

  tournamentId = this.route.snapshot.paramMap.get('id')!;

  tournament = signal<Tournament | null>(null);
  entries = signal<SpinWheelEntry[]>([]);
  matches = signal<SpinWheelMatch[]>([]);
  loading = signal(true);
  errorMessage = signal<string | null>(null);
  // Separate from errorMessage so player-registration errors (add/edit/
  // delete/full-bracket) only ever show once, right next to the add-player
  // form — not also duplicated in the page-top banner.
  entryError = signal<string | null>(null);
  linkCopied = signal(false);

  manualName = '';
  manualPhone = '';
  addingEntry = signal(false);

  scoreDrafts = new Map<string, { p1: number; p2: number }>();
  submittingMatch = signal<string | null>(null);

  entryMap = computed(() => {
    const map = new Map<string, SpinWheelEntry>();
    for (const e of this.entries()) map.set(e.id, e);
    return map;
  });

  rounds = computed(() => {
    const byRound = new Map<number, SpinWheelMatch[]>();
    for (const m of this.matches()) {
      const list = byRound.get(m.round) ?? [];
      list.push(m);
      byRound.set(m.round, list);
    }
    return [...byRound.entries()]
      .sort((a, b) => a[0] - b[0])
      .map(([round, matches]) => ({
        round,
        matches: matches.sort((a, b) => a.match_number - b.match_number),
      }));
  });

  registrationUrl = computed(() => {
    const t = this.tournament();
    if (!t) return '';
    return `${window.location.origin}/register/${t.registration_link_token}`;
  });

  canStartSpin = computed(() => {
    const t = this.tournament();
    return !!t && t.status === 'registration_open' && this.entries().length === t.bracket_size;
  });

  ngOnInit() {
    this.load();
  }

  private load() {
    this.service.get(this.tournamentId).subscribe({
      next: (t) => {
        this.tournament.set(t);
        this.loadEntries();
        if (t.status !== 'registration_open') {
          this.loadFixtures();
        } else {
          this.loading.set(false);
        }
      },
      error: () => {
        this.errorMessage.set('Could not load tournament');
        this.loading.set(false);
      },
    });
  }

  private loadEntries() {
    this.service.listEntries(this.tournamentId).subscribe({
      next: (res) => {
        this.entries.set(res.entries ?? []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }

  refreshingEntries = signal(false);

  refreshEntries() {
    this.refreshingEntries.set(true);
    this.service.listEntries(this.tournamentId).subscribe({
      next: (res) => {
        this.entries.set(res.entries ?? []);
        this.refreshingEntries.set(false);
      },
      error: () => this.refreshingEntries.set(false),
    });
  }

  private loadFixtures() {
    this.service.getFixtures(this.tournamentId).subscribe({
      next: (res) => this.matches.set(res.matches ?? []),
    });
  }

  addEntry() {
    if (!this.manualName.trim()) return;
    this.addingEntry.set(true);
    this.entryError.set(null);
    this.service.addManualEntry(this.tournamentId, this.manualName.trim(), this.manualPhone.trim()).subscribe({
      next: (entry) => {
        this.entries.update((list) => [...list, entry]);
        this.manualName = '';
        this.manualPhone = '';
        this.addingEntry.set(false);
      },
      error: (err) => {
        this.addingEntry.set(false);
        this.entryError.set(err.error?.error ?? 'Could not add player');
      },
    });
  }

  editingEntryId = signal<string | null>(null);
  editName = '';
  editPhone = '';
  savingEdit = signal(false);
  deletingEntryId = signal<string | null>(null);

  startEdit(entry: SpinWheelEntry) {
    this.editingEntryId.set(entry.id);
    this.editName = entry.player_name;
    this.editPhone = entry.phone ?? '';
  }

  cancelEdit() {
    this.editingEntryId.set(null);
  }

  saveEdit(entry: SpinWheelEntry) {
    if (!this.editName.trim()) return;
    this.savingEdit.set(true);
    this.entryError.set(null);
    this.service.updateEntry(this.tournamentId, entry.id, this.editName.trim(), this.editPhone.trim()).subscribe({
      next: (updated) => {
        this.entries.update((list) => list.map((e) => (e.id === updated.id ? updated : e)));
        this.savingEdit.set(false);
        this.editingEntryId.set(null);
      },
      error: (err) => {
        this.savingEdit.set(false);
        this.entryError.set(err.error?.error ?? 'Could not update player');
      },
    });
  }

  deleteEntry(entry: SpinWheelEntry) {
    this.deletingEntryId.set(entry.id);
    this.entryError.set(null);
    this.service.deleteEntry(this.tournamentId, entry.id).subscribe({
      next: () => {
        this.entries.update((list) => list.filter((e) => e.id !== entry.id));
        this.deletingEntryId.set(null);
      },
      error: (err) => {
        this.deletingEntryId.set(null);
        this.entryError.set(err.error?.error ?? 'Could not remove player');
      },
    });
  }

  deletingTournament = signal(false);

  async deleteTournament() {
    const confirmed = await this.confirmService.ask(
      `Delete "${this.tournament()?.name}"? This cannot be undone.`,
      { title: 'Delete tournament' }
    );
    if (!confirmed) return;
    this.deletingTournament.set(true);
    this.errorMessage.set(null);
    this.service.delete(this.tournamentId).subscribe({
      next: () => {
        this.router.navigate(['/spin-wheel']);
      },
      error: (err) => {
        this.deletingTournament.set(false);
        this.errorMessage.set(err.error?.error ?? 'Could not delete tournament');
      },
    });
  }

  copyLink() {
    navigator.clipboard?.writeText(this.registrationUrl()).then(() => {
      this.linkCopied.set(true);
      setTimeout(() => this.linkCopied.set(false), 2000);
    });
  }

  goToSpinWheel() {
    this.router.navigate(['/spin-wheel', this.tournamentId, 'spin']);
  }

  draftFor(matchId: string) {
    if (!this.scoreDrafts.has(matchId)) {
      this.scoreDrafts.set(matchId, { p1: 0, p2: 0 });
    }
    return this.scoreDrafts.get(matchId)!;
  }

  submitResult(match: SpinWheelMatch) {
    this.errorMessage.set(null);
    const draft = this.draftFor(match.id);
    if (draft.p1 === draft.p2) {
      this.errorMessage.set('Scores cannot be tied — enter a winner');
      return;
    }

    this.submittingMatch.set(match.id);
    this.service.submitResult(this.tournamentId, match.id, draft.p1, draft.p2).subscribe({
      next: () => {
        this.submittingMatch.set(null);
        this.load();
      },
      error: (err) => {
        this.submittingMatch.set(null);
        this.errorMessage.set(err.error?.error ?? 'Could not submit result');
      },
    });
  }

  roundLabel(round: number): string {
    const bracketSize = this.tournament()?.bracket_size ?? 0;
    const fullBracket = Math.pow(2, Math.ceil(Math.log2(Math.max(bracketSize, 1))));
    // Number of players entering this particular round (halves each round).
    const participantsInRound = fullBracket / Math.pow(2, round - 1);

    if (participantsInRound === 2) return 'Final';
    if (participantsInRound === 4) return 'Semifinal';
    if (participantsInRound === 8) return 'Quarterfinal';
    return `Round of ${participantsInRound}`;
  }

  playerName(id: string | null): string {
    if (!id) return 'TBD';
    return this.entryMap().get(id)?.player_name ?? 'Unknown';
  }

  playerColor(id: string | null): string {
    if (!id) return '#e2e8f0';
    return this.entryMap().get(id)?.color_hex ?? '#e2e8f0';
  }

  championName = computed(() => {
    const t = this.tournament();
    if (!t || t.status !== 'completed') return null;
    const rounds = this.rounds();
    const finalRound = rounds[rounds.length - 1];
    const finalMatch = finalRound?.matches[0];
    return finalMatch?.winner_entry_id ? this.playerName(finalMatch.winner_entry_id) : null;
  });

  shareOnWhatsApp() {
    const t = this.tournament();
    if (!t) return;

    const lines: string[] = [`*${t.name}* - Fixtures`, ''];

    for (const r of this.rounds()) {
      lines.push(`*${this.roundLabel(r.round)}*`);
      for (const m of r.matches) {
        let p1 = this.playerName(m.player1_entry_id);
        let p2 = m.player2_entry_id ? this.playerName(m.player2_entry_id) : 'BYE';
        if (m.status === 'completed') {
          if (m.winner_entry_id === m.player1_entry_id) p1 = `*${p1}*`;
          if (m.winner_entry_id === m.player2_entry_id) p2 = `*${p2}*`;
          lines.push(`${p1} ${m.player1_score} - ${m.player2_score} ${p2}`);
        } else {
          lines.push(`${p1} vs ${p2}`);
        }
      }
      lines.push('');
    }

    const champion = this.championName();
    if (champion) {
      lines.push(`Champion: *${champion}*`);
    }

    const text = encodeURIComponent(lines.join('\n'));
    window.open(`https://wa.me/?text=${text}`, '_blank');
  }
}
