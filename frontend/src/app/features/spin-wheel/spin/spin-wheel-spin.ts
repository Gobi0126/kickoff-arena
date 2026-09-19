import { AfterViewInit, Component, ElementRef, OnInit, ViewChild, effect, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { SpinWheelService } from '../../../core/services/spin-wheel.service';
import { SpinWheelEntry, Tournament } from '../../../core/models/tournament.model';

const SPIN_DURATION_MS = 4200;
const EXTRA_SPINS = 5;

@Component({
  selector: 'app-spin-wheel-spin',
  imports: [RouterLink],
  templateUrl: './spin-wheel-spin.html',
  styleUrl: './spin-wheel-spin.scss',
})
export class SpinWheelSpin implements OnInit, AfterViewInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private service = inject(SpinWheelService);

  @ViewChild('canvas') canvasRef!: ElementRef<HTMLCanvasElement>;

  tournamentId = this.route.snapshot.paramMap.get('id')!;

  tournament = signal<Tournament | null>(null);
  remaining = signal<SpinWheelEntry[]>([]);
  pickOrder = signal<SpinWheelEntry[]>([]);
  loading = signal(true);
  errorMessage = signal<string | null>(null);
  spinning = signal(false);
  confirming = signal(false);
  rotationDeg = signal(0);

  private viewReady = signal(false);

  constructor() {
    // Redraw whenever the remaining player set changes, or once the view
    // (and therefore the canvas element, which only exists once loading
    // flips to false and the @if reveals it) becomes ready. requestAnimationFrame
    // defers the actual paint until after Angular has committed the DOM
    // update, so we never draw against a stale/missing canvas reference.
    effect(() => {
      this.remaining();
      this.viewReady();
      requestAnimationFrame(() => this.drawWheel());
    });
  }

  ngOnInit() {
    this.service.get(this.tournamentId).subscribe({
      next: (t) => (this.tournament.set(t)),
    });

    this.service.assignColors(this.tournamentId).subscribe({
      next: (res) => {
        this.remaining.set(res.entries ?? []);
        this.loading.set(false);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.error ?? 'Could not load players for the wheel');
        this.loading.set(false);
      },
    });
  }

  ngAfterViewInit() {
    this.viewReady.set(true);
  }

  get allPicked() {
    return this.remaining().length === 0 && this.pickOrder().length > 0;
  }

  spin() {
    if (this.spinning() || this.remaining().length === 0) return;

    const entries = this.remaining();
    const n = entries.length;
    const sliceAngle = 360 / n;
    const winnerIndex = Math.floor(Math.random() * n);

    // Degrees, measured clockwise from 12 o'clock, of the winning slice's center.
    const winnerCenter = (winnerIndex + 0.5) * sliceAngle;
    // Rotate the wheel so that slice ends up under the fixed top pointer,
    // plus a few extra full turns for visual effect.
    const targetRotation =
      this.rotationDeg() -
      (this.rotationDeg() % 360) +
      EXTRA_SPINS * 360 +
      (360 - winnerCenter);

    this.spinning.set(true);
    this.rotationDeg.set(targetRotation);

    setTimeout(() => {
      const winner = entries[winnerIndex];
      this.remaining.set(entries.filter((_, i) => i !== winnerIndex));
      this.pickOrder.set([...this.pickOrder(), winner]);
      this.spinning.set(false);
    }, SPIN_DURATION_MS);
  }

  directMatchup() {
    if (this.spinning() || this.remaining().length !== 2) return;
    this.pickOrder.set([...this.pickOrder(), ...this.remaining()]);
    this.remaining.set([]);
  }

  confirmFixtures() {
    this.confirming.set(true);
    const order = this.pickOrder().map((e) => e.id);
    this.service.startSpin(this.tournamentId, order).subscribe({
      next: () => {
        this.router.navigate(['/spin-wheel', this.tournamentId]);
      },
      error: (err) => {
        this.confirming.set(false);
        this.errorMessage.set(err.error?.error ?? 'Could not create the fixtures');
      },
    });
  }

  private drawWheel() {
    const canvas = this.canvasRef?.nativeElement;
    if (!canvas) return;

    const entries = this.remaining();
    const size = canvas.width;
    const radius = size / 2;
    const ctx = canvas.getContext('2d')!;
    ctx.clearRect(0, 0, size, size);

    if (entries.length === 0) return;

    const sliceAngleRad = (2 * Math.PI) / entries.length;

    entries.forEach((entry, i) => {
      const start = -Math.PI / 2 + i * sliceAngleRad;
      const end = start + sliceAngleRad;

      ctx.beginPath();
      ctx.moveTo(radius, radius);
      ctx.arc(radius, radius, radius - 4, start, end);
      ctx.closePath();
      ctx.fillStyle = entry.color_hex ?? '#94a3b8';
      ctx.fill();
      ctx.strokeStyle = 'rgba(255,255,255,0.6)';
      ctx.lineWidth = 2;
      ctx.stroke();

      ctx.save();
      ctx.translate(radius, radius);
      ctx.rotate(start + sliceAngleRad / 2);
      ctx.textAlign = 'right';
      ctx.fillStyle = '#fff';
      ctx.font = '600 16px -apple-system, sans-serif';
      ctx.fillText(entry.player_name, radius - 16, 6);
      ctx.restore();
    });
  }
}
