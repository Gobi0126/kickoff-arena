import { Component, inject, signal } from '@angular/core';
import { AbstractControl, FormBuilder, ReactiveFormsModule, ValidationErrors, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { SpinWheelService } from '../../../core/services/spin-wheel.service';

function evenNumberValidator(control: AbstractControl): ValidationErrors | null {
  const value = control.value;
  if (value === null || value === '') return null;
  return Number(value) % 2 === 0 ? null : { notEven: true };
}

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

  form = this.fb.nonNullable.group({
    name: ['', [Validators.required]],
    bracket_size: [8, [Validators.required, Validators.min(2), Validators.max(64), evenNumberValidator]],
  });

  get bracketSizeErrors(): string | null {
    const control = this.form.controls.bracket_size;
    if (!control.touched || !control.errors) return null;
    if (control.errors['required']) return 'Player count is required';
    if (control.errors['min']) return 'Must be at least 2 players';
    if (control.errors['max']) return 'Must be 64 players or fewer';
    if (control.errors['notEven']) return 'Player count must be an even number';
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
