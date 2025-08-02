package email

import (
	"net/mail"
	"time"
)

type Email struct {
	ID          string            `json:"id"`
	From        *mail.Address     `json:"from"`
	To          []*mail.Address   `json:"to"`
	CC          []*mail.Address   `json:"cc,omitempty"`
	BCC         []*mail.Address   `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	Body        EmailBody         `json:"body"`
	Headers     map[string]string `json:"headers"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	ReceivedAt  time.Time         `json:"received_at"`
	Size        int64             `json:"size"`
	IsUTF8      bool              `json:"is_utf8"`
}

type EmailBody struct {
	Text string `json:"text,omitempty"`
	HTML string `json:"html,omitempty"`
	Raw  string `json:"raw"`
}

type Attachment struct {
	Filename    string            `json:"filename"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Data        []byte            `json:"-"` // Don't serialize binary data in JSON
	Headers     map[string]string `json:"headers,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Value != "" {
		return e.Field + ": " + e.Message + " (value: " + e.Value + ")"
	}
	return e.Field + ": " + e.Message
}

type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

type ParseResult struct {
	Email  *Email            `json:"email"`
	Errors []ValidationError `json:"errors,omitempty"`
}

type EmailAddress struct {
	Local  string `json:"local"`  // Part before @
	Domain string `json:"domain"` // Part after @
	Raw    string `json:"raw"`    // Original string
	IsUTF8 bool   `json:"is_utf8"`
}

func (ea EmailAddress) String() string {
	return ea.Raw
}

type EmailStats struct {
	TotalSize      int64 `json:"total_size"`
	HeaderSize     int64 `json:"header_size"`
	BodySize       int64 `json:"body_size"`
	AttachmentSize int64 `json:"attachment_size"`
	LineCount      int   `json:"line_count"`
	WordCount      int   `json:"word_count"`
	CharCount      int   `json:"char_count"`
}
