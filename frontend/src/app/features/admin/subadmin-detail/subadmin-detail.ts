import { Component, inject, signal, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { DatePipe } from '@angular/common';
import { AdminService, SubadminDetail } from '../../../core/services/admin.service';

@Component({
  selector: 'app-subadmin-detail',
  imports: [RouterLink, DatePipe],
  templateUrl: './subadmin-detail.html',
  styleUrl: './subadmin-detail.scss',
})
export class SubadminDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private adminService = inject(AdminService);

  detail = signal<SubadminDetail | null>(null);
  loading = signal(true);
  errorMessage = signal<string | null>(null);

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id')!;
    this.adminService.getSubadmin(id).subscribe({
      next: (res) => {
        this.detail.set(res);
        this.loading.set(false);
      },
      error: () => {
        this.errorMessage.set('Subadmin not found');
        this.loading.set(false);
      },
    });
  }
}
