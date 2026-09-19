import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';
import { User } from '../models/user.model';
import { Tournament } from '../models/tournament.model';

export interface SubadminDetail {
  subadmin: User;
  spinWheels: Tournament[];
  auctionTours: Tournament[];
}

@Injectable({ providedIn: 'root' })
export class AdminService {
  private http = inject(HttpClient);

  listSubadmins() {
    return this.http.get<{ subadmins: User[] }>(`${environment.apiUrl}/admin/subadmins`);
  }

  getSubadmin(id: string) {
    return this.http.get<SubadminDetail>(`${environment.apiUrl}/admin/subadmins/${id}`);
  }

  deleteSubadmin(id: string) {
    return this.http.delete<void>(`${environment.apiUrl}/admin/subadmins/${id}`);
  }
}
