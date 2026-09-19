import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { DatePipe } from '@angular/common';
import { SpinWheelService } from '../../../core/services/spin-wheel.service';
import { ConfirmService } from '../../../core/services/confirm.service';
import { Tournament } from '../../../core/models/tournament.model';

type StatusFilter = 'all' | Tournament['status'];
type SortOrder = 'newest' | 'oldest';

const PAGE_SIZE_OPTIONS = [30, 60, 90, 120];

@Component({
  selector: 'app-spin-wheel-list',
  imports: [RouterLink, FormsModule, DatePipe],
  templateUrl: './spin-wheel-list.html',
  styleUrl: './spin-wheel-list.scss',
})
export class SpinWheelList implements OnInit {
  private service = inject(SpinWheelService);
  private confirmService = inject(ConfirmService);

  tournaments = signal<Tournament[]>([]);
  loading = signal(true);
  deletingId = signal<string | null>(null);

  searchQuery = signal('');
  statusFilter = signal<StatusFilter>('all');
  createdAtSort = signal<SortOrder>('newest');

  pageSizeOptions = PAGE_SIZE_OPTIONS;
  pageSize = signal(PAGE_SIZE_OPTIONS[0]);
  currentPage = signal(1);

  editingTournament = signal<Tournament | null>(null);
  playerCountDraft = 0;
  savingEdit = signal(false);
  editError = signal<string | null>(null);

  filteredTournaments = computed(() => {
    const query = this.searchQuery().trim().toLowerCase();
    const filter = this.statusFilter();
    const order = this.createdAtSort();

    let list = this.tournaments();
    if (query) {
      list = list.filter((t) => t.name.toLowerCase().includes(query));
    }
    if (filter !== 'all') {
      list = list.filter((t) => t.status === filter);
    }

    return [...list].sort((a, b) => {
      const diff = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      return order === 'newest' ? -diff : diff;
    });
  });

  totalPages = computed(() => Math.max(1, Math.ceil(this.filteredTournaments().length / this.pageSize())));

  pagedTournaments = computed(() => {
    const page = this.currentPage();
    const size = this.pageSize();
    const start = (page - 1) * size;
    return this.filteredTournaments().slice(start, start + size);
  });

  rangeStart = computed(() => {
    if (this.filteredTournaments().length === 0) return 0;
    return (this.currentPage() - 1) * this.pageSize() + 1;
  });

  rangeEnd = computed(() => {
    return Math.min(this.currentPage() * this.pageSize(), this.filteredTournaments().length);
  });

  // Windowed pagination — always current ± 2, plus first/last page, with
  // '…' gap markers in between, so the row never grows wide enough to wrap
  // onto a second line no matter how many pages there are.
  pageNumbers = computed<(number | '…')[]>(() => {
    const total = this.totalPages();
    const current = this.currentPage();

    if (total <= 7) {
      return Array.from({ length: total }, (_, i) => i + 1);
    }

    const pages = new Set<number>([1, total, current, current - 1, current - 2, current + 1, current + 2]);
    const sorted = [...pages].filter((p) => p >= 1 && p <= total).sort((a, b) => a - b);

    const result: (number | '…')[] = [];
    let prev = 0;
    for (const p of sorted) {
      if (prev && p - prev > 1) result.push('…');
      result.push(p);
      prev = p;
    }
    return result;
  });

  refreshing = signal(false);

  ngOnInit() {
    this.service.listMine().subscribe({
      next: (res) => {
        this.tournaments.set(res.tournaments ?? []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }

  refresh() {
    this.refreshing.set(true);
    this.service.listMine().subscribe({
      next: (res) => {
        this.tournaments.set(res.tournaments ?? []);
        this.refreshing.set(false);
      },
      error: () => this.refreshing.set(false),
    });
  }

  statusLabel(status: Tournament['status']) {
    return status.replace('_', ' ');
  }

  toggleCreatedAtSort() {
    this.createdAtSort.update((s) => (s === 'newest' ? 'oldest' : 'newest'));
    this.currentPage.set(1);
  }

  resetToFirstPage() {
    this.currentPage.set(1);
  }

  changePageSize(size: number) {
    this.pageSize.set(size);
    this.currentPage.set(1);
  }

  goToPage(page: number) {
    if (page < 1 || page > this.totalPages()) return;
    this.currentPage.set(page);
  }

  openEdit(t: Tournament, event: Event) {
    event.preventDefault();
    event.stopPropagation();
    this.playerCountDraft = t.bracket_size;
    this.editError.set(null);
    this.editingTournament.set(t);
  }

  closeEdit() {
    this.editingTournament.set(null);
  }

  saveEdit() {
    const t = this.editingTournament();
    if (!t) return;

    this.savingEdit.set(true);
    this.editError.set(null);
    this.service.updateBracketSize(t.id, this.playerCountDraft).subscribe({
      next: (updated) => {
        this.tournaments.update((list) => list.map((x) => (x.id === updated.id ? updated : x)));
        this.savingEdit.set(false);
        this.editingTournament.set(null);
      },
      error: (err) => {
        this.savingEdit.set(false);
        this.editError.set(err.error?.error ?? 'Could not update player count');
      },
    });
  }

  async deleteTournament(t: Tournament, event: Event) {
    event.preventDefault();
    event.stopPropagation();
    const confirmed = await this.confirmService.ask(`Delete "${t.name}"? This cannot be undone.`, {
      title: 'Delete tournament',
    });
    if (!confirmed) return;

    this.deletingId.set(t.id);
    this.service.delete(t.id).subscribe({
      next: () => {
        this.tournaments.update((list) => list.filter((x) => x.id !== t.id));
        this.deletingId.set(null);
      },
      error: () => this.deletingId.set(null),
    });
  }
}
