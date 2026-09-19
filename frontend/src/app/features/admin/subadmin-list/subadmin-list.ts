import { Component, inject, signal, OnInit } from '@angular/core';
import { RouterLink } from '@angular/router';
import { DatePipe } from '@angular/common';
import { AdminService } from '../../../core/services/admin.service';
import { ConfirmService } from '../../../core/services/confirm.service';
import { User } from '../../../core/models/user.model';

@Component({
  selector: 'app-subadmin-list',
  imports: [RouterLink, DatePipe],
  templateUrl: './subadmin-list.html',
  styleUrl: './subadmin-list.scss',
})
export class SubadminList implements OnInit {
  private adminService = inject(AdminService);
  private confirmService = inject(ConfirmService);

  subadmins = signal<User[]>([]);
  loading = signal(true);
  errorMessage = signal<string | null>(null);
  deleteError = signal<string | null>(null);
  deletingId = signal<string | null>(null);

  ngOnInit() {
    this.adminService.listSubadmins().subscribe({
      next: (res) => {
        this.subadmins.set(res.subadmins);
        this.loading.set(false);
      },
      error: () => {
        this.errorMessage.set('Could not load subadmins');
        this.loading.set(false);
      },
    });
  }

  async deleteSubadmin(user: User, event: Event) {
    event.preventDefault();
    event.stopPropagation();
    const confirmed = await this.confirmService.ask(
      `Delete ${user.name}'s account? This will also delete every tournament they created. This cannot be undone.`,
      { title: 'Delete subadmin' }
    );
    if (!confirmed) return;

    this.deletingId.set(user.id);
    this.deleteError.set(null);
    this.adminService.deleteSubadmin(user.id).subscribe({
      next: () => {
        this.subadmins.update((list) => list.filter((u) => u.id !== user.id));
        this.deletingId.set(null);
      },
      error: (err) => {
        this.deletingId.set(null);
        this.deleteError.set(err.error?.error ?? 'Could not delete subadmin');
      },
    });
  }
}
