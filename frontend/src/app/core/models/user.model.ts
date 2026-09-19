export type Role = 'admin' | 'subadmin' | 'user';

export interface User {
  id: string;
  name: string;
  phone: string;
  role: Role;
  created_at: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}
