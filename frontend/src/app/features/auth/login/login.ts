import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { COUNTRY_CODES, DEFAULT_COUNTRY_DIAL_CODE } from '../../../core/constants/country-codes';

@Component({
  selector: 'app-login',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './login.html',
  styleUrl: './login.scss',
})
export class Login {
  private fb = inject(FormBuilder);
  private auth = inject(AuthService);
  private router = inject(Router);

  loading = signal(false);
  errorMessage = signal<string | null>(null);
  countryCodes = COUNTRY_CODES;

  form = this.fb.nonNullable.group({
    countryCode: [DEFAULT_COUNTRY_DIAL_CODE, [Validators.required]],
    phone: ['', [Validators.required]],
    password: ['', [Validators.required]],
  });

  submit() {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);
    this.errorMessage.set(null);

    const { countryCode, phone, password } = this.form.getRawValue();
    const fullPhone = `${countryCode}${phone.trim()}`;

    this.auth.login(fullPhone, password).subscribe({
      next: (res) => {
        this.loading.set(false);
        const destination = res.user.role === 'admin' ? '/admin/subadmins' : '/dashboard';
        this.router.navigate([destination]);
      },
      error: () => {
        this.loading.set(false);
        this.errorMessage.set('Invalid phone or password');
      },
    });
  }
}
