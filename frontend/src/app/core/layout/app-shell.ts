import { Component, computed, inject } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-shell',
  imports: [RouterLink, RouterLinkActive, RouterOutlet],
  templateUrl: './app-shell.html',
  styleUrl: './app-shell.scss',
})
export class AppShell {
  auth = inject(AuthService);

  // End users only ever see "Admin" — the subadmin/admin distinction
  // is an internal detail, never shown in the UI.
  displayRole = computed(() => (this.auth.role() ? 'Admin' : ''));
}
