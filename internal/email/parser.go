package email

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type EmailParser struct {
	MaxSize        int64
	MaxAttachments int
	AllowUTF8      bool
}

func NewEmailParser() *EmailParser {
	return &EmailParser{
		MaxSize:        25 * 1024 * 1024, // 25MB
		MaxAttachments: 10,
		AllowUTF8:      true,
	}
}

// ParseEmail parses a raw email message
func (p *EmailParser) ParseEmail(rawEmail string) (*ParseResult, error) {
	result := &ParseResult{
		Email:  &Email{},
		Errors: []ValidationError{},
	}

	if rawEmail == "" {
		result.addError("email", "Email content cannot be empty", "")
		return result, nil
	}

	if int64(len(rawEmail)) > p.MaxSize {
		result.addError("email", fmt.Sprintf("Email too large (max %d bytes)", p.MaxSize), "")
		return result, nil
	}

	if !utf8.ValidString(rawEmail) {
		result.addError("email", "Invalid UTF-8 encoding in email", "")
		return result, nil
	}

	reader := strings.NewReader(rawEmail)
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		result.addError("email", "Failed to parse email: "+err.Error(), "")
		return result, nil
	}

	result.Email.ID = p.generateEmailID()
	result.Email.ReceivedAt = time.Now()
	result.Email.Size = int64(len(rawEmail))
	result.Email.IsUTF8 = !isASCII(rawEmail)

	result.Email.Headers = make(map[string]string)
	for key, values := range msg.Header {
		result.Email.Headers[key] = strings.Join(values, ", ")
	}

	p.parseStandardHeaders(msg.Header, result)

	p.parseBody(msg, result)

	return result, nil
}

func (p *EmailParser) parseStandardHeaders(headers mail.Header, result *ParseResult) {
	// Parse From
	if from := headers.Get("From"); from != "" {
		if addr, err := mail.ParseAddress(from); err == nil {
			result.Email.From = addr
		} else {
			result.addError("from", "Invalid From address: "+err.Error(), from)
		}
	}

	// Parse To
	if to := headers.Get("To"); to != "" {
		if addrs, err := mail.ParseAddressList(to); err == nil {
			result.Email.To = addrs
		} else {
			result.addError("to", "Invalid To addresses: "+err.Error(), to)
		}
	}

	// Parse CC
	if cc := headers.Get("Cc"); cc != "" {
		if addrs, err := mail.ParseAddressList(cc); err == nil {
			result.Email.CC = addrs
		} else {
			result.addError("cc", "Invalid CC addresses: "+err.Error(), cc)
		}
	}

	// Parse BCC
	if bcc := headers.Get("Bcc"); bcc != "" {
		if addrs, err := mail.ParseAddressList(bcc); err == nil {
			result.Email.BCC = addrs
		} else {
			result.addError("bcc", "Invalid BCC addresses: "+err.Error(), bcc)
		}
	}

	// Parse Subject
	if subject := headers.Get("Subject"); subject != "" {
		decoded, err := p.decodeHeader(subject)
		if err != nil {
			result.addError("subject", "Failed to decode subject: "+err.Error(), subject)
		} else {
			result.Email.Subject = decoded
		}
	}
}

func (p *EmailParser) parseBody(msg *mail.Message, result *ParseResult) {
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		result.addError("body", "Invalid Content-Type: "+err.Error(), contentType)
		mediaType = "text/plain"
	}

	bodyBytes, err := io.ReadAll(msg.Body)
	if err != nil {
		result.addError("body", "Failed to read body: "+err.Error(), "")
		return
	}

	result.Email.Body.Raw = string(bodyBytes)

	switch {
	case strings.HasPrefix(mediaType, "multipart/"):
		p.parseMultipartBody(bodyBytes, params, result)
	case mediaType == "text/plain":
		result.Email.Body.Text = string(bodyBytes)
	case mediaType == "text/html":
		result.Email.Body.HTML = string(bodyBytes)
	default:
		// Treat as plain text for unknown types
		result.Email.Body.Text = string(bodyBytes)
	}
}

func (p *EmailParser) parseMultipartBody(bodyBytes []byte, params map[string]string, result *ParseResult) {
	boundary := params["boundary"]
	if boundary == "" {
		result.addError("body", "Missing boundary in multipart message", "")
		return
	}

	reader := multipart.NewReader(bytes.NewReader(bodyBytes), boundary)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.addError("body", "Error reading multipart: "+err.Error(), "")
			break
		}

		p.parseMultipartPart(part, result)
		part.Close()

		// Limit number of attachments
		if len(result.Email.Attachments) >= p.MaxAttachments {
			result.addError("attachments", fmt.Sprintf("Too many attachments (max %d)", p.MaxAttachments), "")
			break
		}
	}
}

func (p *EmailParser) parseMultipartPart(part *multipart.Part, result *ParseResult) {
	contentType := part.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = "text/plain"
	}

	// Read part content
	content, err := io.ReadAll(part)
	if err != nil {
		result.addError("body", "Failed to read part content: "+err.Error(), "")
		return
	}

	// Decode if needed
	encoding := part.Header.Get("Content-Transfer-Encoding")
	if encoding != "" {
		content = p.decodeContent(content, encoding)
	}

	// Handle based on content type and disposition
	disposition := part.Header.Get("Content-Disposition")

	if disposition != "" && strings.HasPrefix(disposition, "attachment") {
		// This is an attachment
		p.parseAttachment(part, content, result)
	} else {
		// This is body content
		switch mediaType {
		case "text/plain":
			if result.Email.Body.Text == "" {
				result.Email.Body.Text = string(content)
			}
		case "text/html":
			if result.Email.Body.HTML == "" {
				result.Email.Body.HTML = string(content)
			}
		}
	}
}

// parseAttachment parses an email attachment
func (p *EmailParser) parseAttachment(part *multipart.Part, content []byte, result *ParseResult) {
	attachment := Attachment{
		Data:        content,
		Size:        int64(len(content)),
		ContentType: part.Header.Get("Content-Type"),
		Headers:     make(map[string]string),
	}

	// Copy headers
	for key, values := range part.Header {
		attachment.Headers[key] = strings.Join(values, ", ")
	}

	// Extract filename
	disposition := part.Header.Get("Content-Disposition")
	if disposition != "" {
		_, params, err := mime.ParseMediaType(disposition)
		if err == nil {
			if filename, ok := params["filename"]; ok {
				attachment.Filename = filename
			}
		}
	}

	// Try to get filename from Content-Type if not found
	if attachment.Filename == "" {
		contentType := part.Header.Get("Content-Type")
		if contentType != "" {
			_, params, err := mime.ParseMediaType(contentType)
			if err == nil {
				if name, ok := params["name"]; ok {
					attachment.Filename = name
				}
			}
		}
	}

	result.Email.Attachments = append(result.Email.Attachments, attachment)
}

// decodeContent decodes content based on transfer encoding
func (p *EmailParser) decodeContent(content []byte, encoding string) []byte {
	switch strings.ToLower(encoding) {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(string(content))
		if err != nil {
			return content // Return original if decode fails
		}
		return decoded
	case "quoted-printable":
		decoded, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(content)))
		if err != nil {
			return content
		}
		return decoded
	default:
		return content
	}
}

// decodeHeader decodes MIME-encoded headers
func (p *EmailParser) decodeHeader(header string) (string, error) {
	dec := new(mime.WordDecoder)
	return dec.DecodeHeader(header)
}

// generateEmailID generates a unique email ID
func (p *EmailParser) generateEmailID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// ParseEmailStats calculates statistics about an email
func (p *EmailParser) ParseEmailStats(rawEmail string) *EmailStats {
	stats := &EmailStats{}

	lines := strings.Split(rawEmail, "\n")
	stats.LineCount = len(lines)
	stats.TotalSize = int64(len(rawEmail))
	stats.CharCount = len(rawEmail)

	// Split headers and body
	headerEnd := strings.Index(rawEmail, "\r\n\r\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(rawEmail, "\n\n")
	}

	if headerEnd > 0 {
		stats.HeaderSize = int64(headerEnd)
		stats.BodySize = stats.TotalSize - stats.HeaderSize

		// Count words in body
		body := rawEmail[headerEnd:]
		stats.WordCount = len(strings.Fields(body))
	} else {
		stats.HeaderSize = 0
		stats.BodySize = stats.TotalSize
		stats.WordCount = len(strings.Fields(rawEmail))
	}

	return stats
}

// ValidateRawEmail performs basic validation on raw email content
func (p *EmailParser) ValidateRawEmail(rawEmail string) *ValidationResult {
	result := &ValidationResult{Valid: true, Errors: []ValidationError{}}

	if rawEmail == "" {
		result.addError("email", "Email content cannot be empty", "")
		return result
	}

	// Size check
	if int64(len(rawEmail)) > p.MaxSize {
		result.addError("email", fmt.Sprintf("Email too large (max %d bytes)", p.MaxSize), "")
	}

	// UTF-8 check
	if !utf8.ValidString(rawEmail) {
		result.addError("email", "Invalid UTF-8 encoding", "")
	}

	// Basic structure check - must have headers
	if !strings.Contains(rawEmail, "\n") && !strings.Contains(rawEmail, "\r\n") {
		result.addError("email", "Email must contain headers", "")
	}

	// Check for required headers by trying to parse
	reader := strings.NewReader(rawEmail)
	if _, err := mail.ReadMessage(reader); err != nil {
		result.addError("email", "Invalid email format: "+err.Error(), "")
	}

	return result
}

// Helper methods

func (r *ParseResult) addError(field, message, value string) {
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// ExtractPlainText extracts plain text from an email, trying HTML if plain text isn't available
func (p *EmailParser) ExtractPlainText(email *Email) string {
	if email.Body.Text != "" {
		return email.Body.Text
	}

	if email.Body.HTML != "" {
		// Basic HTML tag removal for plain text extraction
		re := regexp.MustCompile(`<[^>]*>`)
		text := re.ReplaceAllString(email.Body.HTML, "")
		// Clean up extra whitespace
		text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
		return strings.TrimSpace(text)
	}

	return email.Body.Raw
}
