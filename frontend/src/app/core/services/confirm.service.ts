import { Injectable, signal } from '@angular/core';

export interface ConfirmRequest {
  title: string;
  message: string;
  confirmLabel: string;
  resolve: (confirmed: boolean) => void;
}

@Injectable({ providedIn: 'root' })
export class ConfirmService {
  request = signal<ConfirmRequest | null>(null);

  ask(message: string, options?: { title?: string; confirmLabel?: string }): Promise<boolean> {
    return new Promise((resolve) => {
      this.request.set({
        title: options?.title ?? 'Please confirm',
        message,
        confirmLabel: options?.confirmLabel ?? 'Yes, Delete',
        resolve,
      });
    });
  }

  respond(confirmed: boolean) {
    this.request()?.resolve(confirmed);
    this.request.set(null);
  }
}
