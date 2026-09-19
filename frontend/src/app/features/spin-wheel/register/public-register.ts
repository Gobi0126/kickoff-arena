import { Component, OnInit, inject, signal } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { SpinWheelService } from '../../../core/services/spin-wheel.service';
import { Tournament } from '../../../core/models/tournament.model';
import { COUNTRY_CODES, DEFAULT_COUNTRY_DIAL_CODE } from '../../../core/constants/country-codes';

@Component({
  selector: 'app-public-register',
  imports: [FormsModule],
  templateUrl: './public-register.html',
  styleUrl: './public-register.scss',
})
export class PublicRegister implements OnInit {
  private route = inject(ActivatedRoute);
  private service = inject(SpinWheelService);

  token = this.route.snapshot.paramMap.get('token')!;

  tournament = signal<Tournament | null>(null);
  loading = signal(true);
  notFound = signal(false);
  submitted = signal(false);
  submitting = signal(false);
  errorMessage = signal<string | null>(null);

  countryCodes = COUNTRY_CODES;
  name = '';
  countryCode = DEFAULT_COUNTRY_DIAL_CODE;
  phone = '';

  ngOnInit() {
    this.service.getPublicTournament(this.token).subscribe({
      next: (t) => {
        this.tournament.set(t);
        this.loading.set(false);
      },
      error: () => {
        this.notFound.set(true);
        this.loading.set(false);
      },
    });
  }

  submit() {
    if (!this.name.trim()) return;
    this.submitting.set(true);
    this.errorMessage.set(null);

    const fullPhone = this.phone.trim() ? `${this.countryCode}${this.phone.trim()}` : '';

    this.service.register(this.token, this.name.trim(), fullPhone).subscribe({
      next: () => {
        this.submitting.set(false);
        this.submitted.set(true);
      },
      error: (err) => {
        this.submitting.set(false);
        this.errorMessage.set(err.error?.error ?? 'Could not register, please try again');
      },
    });
  }
}
