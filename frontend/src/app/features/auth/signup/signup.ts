import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { AuthService } from '../../../core/services/auth.service';
import { COUNTRY_CODES, DEFAULT_COUNTRY_DIAL_CODE } from '../../../core/constants/country-codes';

@Component({
  selector: 'app-signup',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './signup.html',
  styleUrl: './signup.scss',
})
export class Signup {
  private fb = inject(FormBuilder);
  private auth = inject(AuthService);
  private router = inject(Router);

  loading = signal(false);
  errorMessage = signal<string | null>(null);
  countryCodes = COUNTRY_CODES;

  form = this.fb.nonNullable.group({
    name: ['', [Validators.required]],
    countryCode: [DEFAULT_COUNTRY_DIAL_CODE, [Validators.required]],
    phone: ['', [Validators.required]],
    password: ['', [Validators.required, Validators.minLength(8)]],
  });

  submit() {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);
    this.errorMessage.set(null);

    const { name, countryCode, phone, password } = this.form.getRawValue();
    const fullPhone = `${countryCode}${phone.trim()}`;

    this.auth.signup(name, fullPhone, password).subscribe({
      next: () => {
        this.loading.set(false);
        this.router.navigate(['/dashboard']);
      },
      error: (err: HttpErrorResponse) => {
        this.loading.set(false);
        this.errorMessage.set(
          err.status === 409
            ? 'This phone number is already registered'
            : err.error?.error ?? 'Something went wrong, please try again'
        );
      },
    });
  }
}
