import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';
import { roleGuard } from './core/guards/role.guard';

export const routes: Routes = [
  {
    path: '',
    pathMatch: 'full',
    loadComponent: () => import('./features/landing/landing').then((m) => m.Landing),
  },
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login/login').then((m) => m.Login),
  },
  {
    path: 'signup',
    loadComponent: () => import('./features/auth/signup/signup').then((m) => m.Signup),
  },
  {
    path: 'register/:token',
    loadComponent: () =>
      import('./features/spin-wheel/register/public-register').then((m) => m.PublicRegister),
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () => import('./core/layout/app-shell').then((m) => m.AppShell),
    children: [
      {
        path: 'dashboard',
        loadComponent: () => import('./features/dashboard/dashboard').then((m) => m.Dashboard),
      },
      {
        path: 'spin-wheel',
        loadComponent: () =>
          import('./features/spin-wheel/list/spin-wheel-list').then((m) => m.SpinWheelList),
      },
      {
        path: 'spin-wheel/new',
        loadComponent: () =>
          import('./features/spin-wheel/create/create-tournament').then(
            (m) => m.CreateTournament
          ),
      },
      {
        path: 'spin-wheel/:id/spin',
        loadComponent: () =>
          import('./features/spin-wheel/spin/spin-wheel-spin').then((m) => m.SpinWheelSpin),
      },
      {
        path: 'spin-wheel/:id',
        loadComponent: () =>
          import('./features/spin-wheel/detail/spin-wheel-detail').then(
            (m) => m.SpinWheelDetail
          ),
      },
      {
        path: 'admin/subadmins',
        canActivate: [roleGuard(['admin'])],
        loadComponent: () =>
          import('./features/admin/subadmin-list/subadmin-list').then((m) => m.SubadminList),
      },
      {
        path: 'admin/subadmins/:id',
        canActivate: [roleGuard(['admin'])],
        loadComponent: () =>
          import('./features/admin/subadmin-detail/subadmin-detail').then(
            (m) => m.SubadminDetailComponent
          ),
      },
    ],
  },
];
