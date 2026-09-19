import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ConfirmDialog } from './core/components/confirm-dialog/confirm-dialog';

@Component({
  imports: [RouterOutlet, ConfirmDialog],
  selector: 'app-root',
  styleUrl: './app.scss',
  templateUrl: './app.html',
})
export class App {}
