export interface Email {
  id: string;
  from: string;
  recipients?: string[];
  subject: string;
  body: string | EmailBody;
  headers?: Record<string, string>;
  attachments?: Attachment[];
  timestamp: string;
  received_at?: string;
  size?: number;
  is_utf8?: boolean;
  read: boolean;
  starred?: boolean;
  hasAttachments?: boolean;
}

export interface EmailBody {
  text: string;
  html: string;
  raw: string;
}

export interface Attachment {
  filename: string;
  content_type: string;
  size: number;
  headers: Record<string, string>;
}

export interface EmailStats {
  received: number;
  total_size: number;
  last_updated: string;
}

export interface APIError {
  message: string;
  status: number;
}