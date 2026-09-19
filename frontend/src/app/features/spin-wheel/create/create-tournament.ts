import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { SpinWheelService } from '../../../core/services/spin-wheel.service';

// Bracket sizes are restricted to powers of two so a knockout bracket never
// needs byes — every registered player always has a real round-1 opponent.
export const BRACKET_SIZE_OPTIONS = [2, 4, 8, 16, 32, 64];

@Component({
  selector: 'app-create-tournament',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './create-tournament.html',
  styleUrl: './create-tournament.scss',
})
export class CreateTournament {
  private fb = inject(FormBuilder);
  private service = inject(SpinWheelService);
  private router = inject(Router);

  loading = signal(false);
  errorMessage = signal<string | null>(null);
  bracketSizeOptions = BRACKET_SIZE_OPTIONS;

  form = this.fb.nonNullable.group({
    name: ['', [Validators.required]],
    bracket_size: [8, [Validators.required]],
  });

  get bracketSizeErrors(): string | null {
    const control = this.form.controls.bracket_size;
    if (!control.touched || !control.errors) return null;
    if (control.errors['required']) return 'Player count is required';
    return null;
  }

  submit() {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);
    this.errorMessage.set(null);

    const { name, bracket_size } = this.form.getRawValue();

    this.service.create(name, bracket_size).subscribe({
      next: (t) => {
        this.loading.set(false);
        this.router.navigate(['/spin-wheel', t.id]);
      },
      error: (err) => {
        this.loading.set(false);
        this.errorMessage.set(err.error?.error ?? 'Could not create tournament, please try again');
      },
    });
  }
}
