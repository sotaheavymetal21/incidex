export interface Attachment {
  id: number;
  incident_id: number;
  user_id: number;
  file_name: string;
  file_size: number;
  mime_type: string;
  storage_key: string;
  created_at: string;
  user?: {
    id: number;
    name: string;
    email: string;
  };
}
