import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-dashboard',
  imports: [RouterLink],
  template: `
    <div class="page">
      <h1>Welcome, {{ auth.user()?.name }}</h1>
      <p class="subtitle">What do you want to run today?</p>

      <div class="cards">
        <a class="card" routerLink="/spin-wheel">
          <span class="icon">🎡</span>
          <h2>Spin Wheel</h2>
          <p>Create random 1v1 matchups and run a knockout bracket.</p>
        </a>
        <div class="card disabled">
          <span class="icon">🏆</span>
          <h2>Player Auction</h2>
          <p>Coming soon.</p>
        </div>
      </div>
    </div>
  `,
  styles: `
    .page {
      padding: 2.5rem;
      max-width: 720px;
    }
    h1 {
      margin: 0;
      font-size: 1.5rem;
      color: #0f172a;
    }
    .subtitle {
      margin: 0.25rem 0 2rem;
      color: #667;
    }
    .cards {
      display: flex;
      gap: 1rem;
      flex-wrap: wrap;
    }
    .card {
      flex: 1;
      min-width: 220px;
      background: #fff;
      border-radius: 14px;
      padding: 1.5rem;
      text-decoration: none;
      color: inherit;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
      transition: transform 0.15s ease, box-shadow 0.15s ease;
    }
    .card:not(.disabled):hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 20px rgba(0, 0, 0, 0.1);
    }
    .card.disabled {
      opacity: 0.6;
    }
    .icon {
      font-size: 1.75rem;
      display: block;
      margin-bottom: 0.5rem;
    }
    .card h2 {
      margin: 0 0 0.5rem;
      font-size: 1.0625rem;
    }
    .card p {
      margin: 0;
      font-size: 0.8125rem;
      color: #667;
      line-height: 1.5;
    }
  `,
})
export class Dashboard {
  auth = inject(AuthService);
}
