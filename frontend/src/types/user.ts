export type Role = 'admin' | 'editor' | 'viewer';

export interface User {
  id: number;
  name: string;
  email: string;
  role: Role;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface UpdateUserRequest {
  name: string;
  email: string;
  role: Role;
}

export interface UpdatePasswordRequest {
  old_password: string;
  new_password: string;
}
