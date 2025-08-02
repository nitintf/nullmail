package smtp

const (
	// Success codes
	CodeOK             = "250"
	CodeServiceReady   = "220"
	CodeServiceClosing = "221"
	CodeStartMailInput = "354"
	CodeAuthSuccessful = "235"
	CodeAuthContinue   = "334"
	CodeStartTLS       = "220"

	// Error codes
	CodeSyntaxError            = "501"
	CodeCommandNotRecognized   = "500"
	CodeCommandNotImplemented  = "502"
	CodeRequestedActionAborted = "451"
	CodeAuthenticationFailed   = "535"
	CodeUserNotLocal           = "550"
	CodeCannotVerify           = "252"
	CodeMessageTooLarge        = "552"
	CodeTLSRequired            = "530"
)

const (
	MsgServiceReady           = "temp-smtp.local ESMTP Ready"
	MsgServiceClosing         = "Bye"
	MsgOK                     = "OK"
	MsgMessageAccepted        = "OK: Message accepted for delivery"
	MsgStartMailInput         = "Start mail input; end with <CRLF>.<CRLF>"
	MsgSyntaxError            = "Syntax error"
	MsgCommandNotRecognized   = "Command not recognized"
	MsgCommandNotImplemented  = "Command not implemented"
	MsgRequestedActionAborted = "Requested action aborted: local error in processing"
	MsgAuthSuccessful         = "Authentication successful"
	MsgAuthFailed             = "Authentication failed"
	MsgAuthContinue           = ""
	MsgUserNotLocal           = "User not local"
	MsgCannotVerify           = "Cannot verify user, but will accept message"
	MsgHelpMessage            = "Commands: HELO EHLO MAIL RCPT DATA VRFY EXPN HELP RSET NOOP QUIT AUTH STARTTLS"
	MsgTurnNotSupported       = "Turn not supported"
	MsgStartTLS               = "Ready to start TLS"
	MsgMessageTooLarge        = "Message too large"
	MsgTLSRequired            = "Must issue STARTTLS first"
	MsgInvalidUTF             = "Invalid UTF-8"
)

const (
	EHLOGreetingTemplate = "250-temp-smtp.local\r\n250-8BITMIME\r\n250-AUTH PLAIN LOGIN\r\n250-STARTTLS\r\n250-SIZE %d\r\n250-SMTPUTF8\r\n250 HELP\r\n"
)

const (
	DefaultHostname = "temp-smtp.local"
	DefaultPort     = ":2525"
	MaxMessageSize  = 25000000 // 25MB
)
