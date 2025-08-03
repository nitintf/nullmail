package smtp

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"

	"nullmail/internal/email"
	"nullmail/internal/redis"
)

type SMTPServer struct {
	listener    net.Listener
	quit        chan struct{}
	tlsConfig   *tls.Config
	emailParser *email.EmailParser
	validator   *email.EmailValidator
	redisClient *redis.Client
}

type SMTPSession struct {
	isTLS       bool
	isUTF8      bool
	messageSize int64
	from        string
	recipients  []string
}

func NewSMTPServer(port string) *SMTPServer {
	redisClient := redis.NewClientFromEnv()

	// Test Redis connection
	if err := redisClient.Ping(); err != nil {
		slog.Warn("Redis connection failed, continuing without Redis", "error", err)
		redisClient = nil
	}

	return &SMTPServer{
		quit:        make(chan struct{}),
		tlsConfig:   loadOrGenerateTLSConfig(),
		emailParser: email.NewEmailParser(),
		validator:   email.NewEmailValidator(),
		redisClient: redisClient,
	}
}

func (s *SMTPServer) Start(port string) error {
	var err error

	s.listener, err = net.Listen("tcp", port)

	if err != nil {
		return fmt.Errorf("failed to start listener on port %s: %v", port, err)
	}

	slog.Info("SMTP server started", "port", port)

	// handle gracefull shutdown
	go s.handleShutdown()

	for {
		select {
		case <-s.quit:
			return nil
		default:
			conn, err := s.listener.Accept()

			if err != nil {
				select {
				case <-s.quit:
					return nil
				default:
					slog.Error("Error accepting connection", "error", err)
					continue
				}
			}

			// handle connection
			go s.handleConnection(conn, &SMTPSession{})
		}
	}

}

func (s *SMTPServer) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("SMTP server shutting down")
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
}

func (s *SMTPServer) handleConnection(conn net.Conn, session *SMTPSession) {
	s.handleConnectionWithoutClose(conn, session)
	conn.Close()
}

func (s *SMTPServer) handleConnectionWithoutClose(conn net.Conn, session *SMTPSession) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	clientAddr := conn.RemoteAddr().String()

	if !session.isTLS {
		slog.Info("New SMTP connection", "client", clientAddr)
		s.sendResponse(writer, CodeServiceReady, MsgServiceReady)
	}

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			// Handle different types of connection errors more gracefully
			if err == io.EOF {
				slog.Debug("Client disconnected", "client", clientAddr)
			} else if netErr, ok := err.(*net.OpError); ok && netErr.Err == syscall.ECONNRESET {
				slog.Debug("Connection reset by client (likely health check)", "client", clientAddr)
			} else {
				slog.Error("Error reading from client", "error", err, "client", clientAddr)
			}
			break
		}

		command := strings.TrimSpace(line)
		slog.Debug("Received SMTP command", "client", clientAddr, "command", command)

		result := s.handleSMTPCommand(command, reader, writer, clientAddr, session, conn)
		if result == -1 {
			// STARTTLS upgrade - connection will be handled by new TLS handler
			return
		} else if result == 0 {
			// Normal connection close
			break
		}
		// result == 1 means continue
	}

	slog.Info("SMTP connection closed", "client", clientAddr)
}

func (s *SMTPServer) handleSMTPCommand(command string, reader *bufio.Reader, writer *bufio.Writer, clientAddr string, session *SMTPSession, conn net.Conn) int {
	parts := strings.Fields(strings.ToUpper(command))

	if len(parts) == 0 {
		s.sendResponse(writer, CodeCommandNotRecognized, MsgCommandNotRecognized)
		return 1
	}

	cmd := parts[0]

	switch cmd {
	case "HELO", "EHLO":
		s.handleHelo(cmd, parts, writer, session)
	case "MAIL":
		s.handleMail(command, writer, session)
	case "RCPT":
		s.handleRcpt(command, writer, session)
	case "DATA":
		s.handleData(reader, writer, clientAddr, session)
	case "QUIT":
		s.sendResponse(writer, CodeServiceClosing, MsgServiceClosing)
		return 0
	case "RSET":
		s.sendResponse(writer, CodeOK, MsgOK)
	case "NOOP":
		s.sendResponse(writer, CodeOK, MsgOK)
	case "AUTH":
		s.handleAuth(command, writer)
	case "VRFY":
		s.handleVrfy(command, writer)
	case "EXPN":
		s.handleExpn(command, writer)
	case "HELP":
		s.handleHelp(writer)
	case "STARTTLS":
		return s.handleStartTLS(writer, conn, session)
	default:
		s.sendResponse(writer, CodeCommandNotImplemented, MsgCommandNotImplemented)
	}

	return 1
}

func (s *SMTPServer) sendResponse(writer *bufio.Writer, code, message string) {
	response := fmt.Sprintf("%s %s\r\n", code, message)
	writer.WriteString(response)
	writer.Flush()
	slog.Debug("Sent SMTP response", "code", code, "message", message)
}

func (s *SMTPServer) handleHelo(cmd string, parts []string, writer *bufio.Writer, session *SMTPSession) {
	if len(parts) < 2 {
		s.sendResponse(writer, CodeSyntaxError, MsgSyntaxError)
		return
	}

	if cmd == "EHLO" {
		// EHLO multi-line response format with dynamic size
		ehloResponse := fmt.Sprintf(EHLOGreetingTemplate, MaxMessageSize)
		writer.WriteString(ehloResponse)
		writer.Flush()
		slog.Debug("Sent EHLO response")
	} else {
		s.sendResponse(writer, CodeOK, DefaultHostname)
	}
}

func (s *SMTPServer) handleMail(cmd string, writer *bufio.Writer, session *SMTPSession) {
	upper := strings.ToUpper(cmd)
	if !strings.Contains(upper, "FROM:") {
		s.sendResponse(writer, CodeSyntaxError, MsgSyntaxError)
		return
	}

	emailAddr := s.extractEmailFromCommand(cmd, "FROM:")
	if emailAddr == "" {
		s.sendResponse(writer, CodeSyntaxError, "Invalid MAIL FROM syntax")
		return
	}

	// Validate the email address
	if result := s.validator.ValidateAddress(emailAddr); !result.Valid {
		slog.Warn("Invalid FROM address", "address", emailAddr, "errors", result.Errors)
		s.sendResponse(writer, CodeSyntaxError, "Invalid FROM address: "+result.Errors[0].Message)
		return
	}

	// Parse SIZE parameter
	if strings.Contains(upper, "SIZE=") {
		parts := strings.Fields(cmd)
		for _, part := range parts {
			if sizeStr, found := strings.CutPrefix(strings.ToUpper(part), "SIZE="); found {
				size, err := strconv.ParseInt(sizeStr, 10, 64)
				if err != nil {
					s.sendResponse(writer, CodeSyntaxError, MsgSyntaxError)
					return
				}
				if size > MaxMessageSize {
					s.sendResponse(writer, CodeMessageTooLarge, MsgMessageTooLarge)
					return
				}
				session.messageSize = size
			}
		}
	}

	// Check for SMTPUTF8
	if strings.Contains(upper, "SMTPUTF8") {
		session.isUTF8 = true
	}

	// Validate UTF-8 if SMTPUTF8 is enabled
	if session.isUTF8 {
		if !utf8.ValidString(cmd) {
			s.sendResponse(writer, CodeSyntaxError, "Invalid UTF-8")
			return
		}
	}

	// Store the validated FROM address in session
	session.from = emailAddr
	slog.Debug("MAIL FROM accepted", "address", emailAddr)
	s.sendResponse(writer, CodeOK, MsgOK)
}

func (s *SMTPServer) handleRcpt(cmd string, writer *bufio.Writer, session *SMTPSession) {
	if !strings.Contains(strings.ToUpper(cmd), "TO:") {
		s.sendResponse(writer, CodeSyntaxError, MsgSyntaxError)
		return
	}

	emailAddr := s.extractEmailFromCommand(cmd, "TO:")
	if emailAddr == "" {
		s.sendResponse(writer, CodeSyntaxError, "Invalid RCPT TO syntax")
		return
	}

	if result := s.validator.ValidateAddress(emailAddr); !result.Valid {
		slog.Warn("Invalid TO address", "address", emailAddr, "errors", result.Errors)
		s.sendResponse(writer, CodeSyntaxError, "Invalid TO address: "+result.Errors[0].Message)
		return
	}

	if session.recipients == nil {
		session.recipients = []string{}
	}
	session.recipients = append(session.recipients, emailAddr)
	slog.Debug("RCPT TO accepted", "address", emailAddr)
	s.sendResponse(writer, CodeOK, MsgOK)
}

func (s *SMTPServer) handleData(reader *bufio.Reader, writer *bufio.Writer, clientAddr string, session *SMTPSession) {
	s.sendResponse(writer, CodeStartMailInput, MsgStartMailInput)

	var emailContent strings.Builder
	var totalSize int64

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Error reading email data", "error", err, "client", clientAddr)
			s.sendResponse(writer, CodeRequestedActionAborted, MsgRequestedActionAborted)
			return
		}

		if strings.TrimSpace(line) == "." {
			break
		}

		if strings.HasPrefix(line, "..") {
			line = line[1:]
		}

		// Check size limits
		totalSize += int64(len(line))
		if totalSize > MaxMessageSize {
			slog.Error("Message too large", "size", totalSize, "limit", MaxMessageSize)
			s.sendResponse(writer, CodeMessageTooLarge, MsgMessageTooLarge)
			return
		}

		// Validate UTF-8 if SMTPUTF8 mode
		if session.isUTF8 && !utf8.ValidString(line) {
			slog.Error("Invalid UTF-8 in message", "client", clientAddr)
			s.sendResponse(writer, CodeSyntaxError, MsgInvalidUTF)
			return
		}

		emailContent.WriteString(line)
	}

	rawEmail := emailContent.String()
	parseResult, err := s.emailParser.ParseEmail(rawEmail)
	if err != nil {
		slog.Error("Failed to parse email", "error", err, "client", clientAddr)
		s.sendResponse(writer, CodeRequestedActionAborted, "Failed to parse email")
		return
	}

	if len(parseResult.Errors) > 0 {
		slog.Warn("Email parsing warnings", "errors", parseResult.Errors, "client", clientAddr)
		// Continue processing despite warnings
	}

	slog.Info("Email received and parsed",
		"client", clientAddr,
		"id", parseResult.Email.ID,
		"from", session.from,
		"recipients", session.recipients,
		"subject", parseResult.Email.Subject,
		"size", emailContent.Len(),
		"attachments", len(parseResult.Email.Attachments))

	if s.redisClient != nil {
		if err := s.storeEmailInRedis(parseResult.Email, session); err != nil {
			slog.Error("Failed to store email in Redis", "error", err, "id", parseResult.Email.ID)
		}
	} else {
		slog.Debug("Redis not available, email not stored")
	}

	slog.Debug("Parsed email structure", "email", parseResult.Email)
	s.sendResponse(writer, CodeOK, MsgMessageAccepted)
}

func (s *SMTPServer) handleAuth(_cmd string, writer *bufio.Writer) {
	s.sendResponse(writer, CodeAuthSuccessful, MsgAuthSuccessful)
}

func (s *SMTPServer) handleVrfy(cmd string, writer *bufio.Writer) {
	parts := strings.Fields(cmd)
	if len(parts) < 2 {
		s.sendResponse(writer, CodeSyntaxError, MsgSyntaxError)
		return
	}

	// For disposable email service, we can't really verify users
	// but we accept all addresses ending with our domain
	email := parts[1]
	if strings.Contains(email, "@nullmail.local") || strings.Contains(email, "@nullmail.nitin.sh") {
		s.sendResponse(writer, CodeOK, fmt.Sprintf("250 %s", email))
	} else {
		s.sendResponse(writer, CodeCannotVerify, MsgCannotVerify)
	}
}

func (s *SMTPServer) handleExpn(_cmd string, writer *bufio.Writer) {
	s.sendResponse(writer, CodeUserNotLocal, MsgUserNotLocal)
}

func (s *SMTPServer) handleHelp(writer *bufio.Writer) {
	s.sendResponse(writer, CodeOK, MsgHelpMessage)
}

func (s *SMTPServer) handleStartTLS(writer *bufio.Writer, conn net.Conn, session *SMTPSession) int {
	if session.isTLS {
		s.sendResponse(writer, CodeCommandNotImplemented, "Already using TLS")
		return 1
	}

	if s.tlsConfig == nil {
		s.sendResponse(writer, CodeCommandNotImplemented, "TLS not available")
		return 1
	}

	s.sendResponse(writer, CodeStartTLS, MsgStartTLS)

	tlsConn := tls.Server(conn, s.tlsConfig)
	err := tlsConn.Handshake()
	if err != nil {
		slog.Error("TLS handshake failed", "error", err)
		return 0
	}

	session.isTLS = true
	slog.Info("TLS connection established", "client", conn.RemoteAddr().String())

	// Continue with TLS connection
	s.handleConnectionWithoutClose(tlsConn, session)
	return -1
}

func (s *SMTPServer) extractEmailFromCommand(cmd, prefix string) string {
	upper := strings.ToUpper(cmd)
	prefixIndex := strings.Index(upper, prefix)
	if prefixIndex == -1 {
		return ""
	}

	remaining := strings.TrimSpace(cmd[prefixIndex+len(prefix):])

	if strings.HasPrefix(remaining, "<") && strings.HasSuffix(remaining, ">") {
		return remaining[1 : len(remaining)-1]
	}

	parts := strings.Fields(remaining)
	if len(parts) > 0 {
		addr := parts[0]
		if strings.HasPrefix(addr, "<") && strings.HasSuffix(addr, ">") {
			return addr[1 : len(addr)-1]
		}
		return addr
	}

	return ""
}

func (s *SMTPServer) storeEmailInRedis(parsedEmail *email.Email, session *SMTPSession) error {
	emailData := map[string]interface{}{
		"id":          parsedEmail.ID,
		"from":        session.from,
		"recipients":  session.recipients,
		"subject":     parsedEmail.Subject,
		"body":        parsedEmail.Body,
		"headers":     parsedEmail.Headers,
		"attachments": parsedEmail.Attachments,
		"received_at": parsedEmail.ReceivedAt,
		"size":        parsedEmail.Size,
		"is_utf8":     parsedEmail.IsUTF8,
	}

	if err := s.redisClient.StoreEmailWithRecipients(parsedEmail.ID, emailData, session.recipients); err != nil {
		return err
	}

	if err := s.redisClient.QueueEmail("inbound", emailData); err != nil {
		slog.Warn("Failed to queue email for processing", "error", err, "id", parsedEmail.ID)
	}

	if err := s.redisClient.IncrementEmailCount("received"); err != nil {
		slog.Warn("Failed to update email statistics", "error", err)
	}

	slog.Info("Email stored in Redis with recipient indexing", "id", parsedEmail.ID, "recipients", session.recipients)
	return nil
}
