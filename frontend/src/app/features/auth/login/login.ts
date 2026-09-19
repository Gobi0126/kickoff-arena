import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';

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

  form = this.fb.nonNullable.group({
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

    const { phone, password } = this.form.getRawValue();

    this.auth.login(phone, password).subscribe({
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
