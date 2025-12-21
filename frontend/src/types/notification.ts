export interface NotificationSetting {
  id?: number;
  user_id: number;

  // 通知チャネル
  email_enabled: boolean;
  slack_enabled: boolean;
  slack_webhook?: string;

  // 通知イベントの有効/無効
  notify_on_incident_created: boolean;
  notify_on_assigned: boolean;
  notify_on_comment: boolean;
  notify_on_status_change: boolean;
  notify_on_severity_change: boolean;
  notify_on_resolved: boolean;
  notify_on_escalation: boolean;

  created_at?: string;
  updated_at?: string;
}
