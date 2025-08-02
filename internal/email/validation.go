package email

import (
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// Basic email regex (RFC 5322 compliant)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	// Domain validation
	domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	// Local part validation (before @)
	localPartRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+$`)
)

type EmailValidator struct {
	AllowUTF8       bool
	MaxLocalLength  int
	MaxDomainLength int
	MaxTotalLength  int
	RequireTLD      bool
	AllowIPDomains  bool
	ValidDomains    []string // Whitelist of allowed domains
	InvalidDomains  []string // Blacklist of forbidden domains
}

// NewEmailValidator creates a new validator with default settings
func NewEmailValidator() *EmailValidator {
	return &EmailValidator{
		AllowUTF8:       true,
		MaxLocalLength:  64,  // RFC 5321 limit
		MaxDomainLength: 253, // RFC 5321 limit
		MaxTotalLength:  320, // RFC 5321 limit (64 + 1 + 255)
		RequireTLD:      true,
		AllowIPDomains:  false,
		ValidDomains:    []string{},
		InvalidDomains:  []string{},
	}
}

func (v *EmailValidator) ValidateAddress(address string) *ValidationResult {
	result := &ValidationResult{Valid: true, Errors: []ValidationError{}}

	// Basic checks
	if address == "" {
		result.addError("address", "Email address cannot be empty", address)
		return result
	}

	// Length check
	if len(address) > v.MaxTotalLength {
		result.addError("address", fmt.Sprintf("Email address too long (max %d chars)", v.MaxTotalLength), address)
	}

	// UTF-8 validation
	if !utf8.ValidString(address) {
		result.addError("address", "Invalid UTF-8 encoding", address)
		return result
	}

	// Check for UTF-8 characters if not allowed
	if !v.AllowUTF8 && !isASCII(address) {
		result.addError("address", "Non-ASCII characters not allowed", address)
	}

	// Parse using Go's mail package
	parsed, err := mail.ParseAddress(address)
	if err != nil {
		result.addError("address", "Invalid email format: "+err.Error(), address)
		return result
	}

	// Split into local and domain parts
	parts := strings.SplitN(parsed.Address, "@", 2)
	if len(parts) != 2 {
		result.addError("address", "Email must contain exactly one @ symbol", address)
		return result
	}

	local, domain := parts[0], parts[1]

	v.validateLocalPart(local, result)

	v.validateDomain(domain, result)

	v.checkDomainPolicy(domain, result)

	return result
}

func (v *EmailValidator) validateLocalPart(local string, result *ValidationResult) {
	if local == "" {
		result.addError("local", "Local part cannot be empty", local)
		return
	}

	if len(local) > v.MaxLocalLength {
		result.addError("local", fmt.Sprintf("Local part too long (max %d chars)", v.MaxLocalLength), local)
	}

	if strings.Contains(local, "..") {
		result.addError("local", "Consecutive dots not allowed", local)
	}

	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		result.addError("local", "Local part cannot start or end with dot", local)
	}

	// Basic character validation for non-quoted local parts
	if !strings.HasPrefix(local, "\"") {
		if !v.AllowUTF8 && !localPartRegex.MatchString(local) {
			result.addError("local", "Invalid characters in local part", local)
		}
	}
}

func (v *EmailValidator) validateDomain(domain string, result *ValidationResult) {
	if domain == "" {
		result.addError("domain", "Domain cannot be empty", domain)
		return
	}

	if len(domain) > v.MaxDomainLength {
		result.addError("domain", fmt.Sprintf("Domain too long (max %d chars)", v.MaxDomainLength), domain)
	}

	// Check for IP address domains
	if strings.HasPrefix(domain, "[") && strings.HasSuffix(domain, "]") {
		if !v.AllowIPDomains {
			result.addError("domain", "IP address domains not allowed", domain)
			return
		}
		v.validateIPDomain(domain[1:len(domain)-1], result)
		return
	}

	if !domainRegex.MatchString(domain) {
		result.addError("domain", "Invalid domain format", domain)
		return
	}

	// Check for TLD requirement
	if v.RequireTLD && !strings.Contains(domain, ".") {
		result.addError("domain", "Domain must have a top-level domain", domain)
	}

	labels := strings.Split(domain, ".")
	for i, label := range labels {
		if label == "" {
			result.addError("domain", "Domain cannot have empty labels", domain)
			continue
		}

		if len(label) > 63 {
			result.addError("domain", fmt.Sprintf("Domain label '%s' too long (max 63 chars)", label), domain)
		}

		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			result.addError("domain", fmt.Sprintf("Domain label '%s' cannot start or end with hyphen", label), domain)
		}

		if i == len(labels)-1 && v.RequireTLD {
			if !isValidTLD(label) {
				result.addError("domain", fmt.Sprintf("Invalid top-level domain: %s", label), domain)
			}
		}
	}
}

func (v *EmailValidator) validateIPDomain(ip string, result *ValidationResult) {
	// Remove IPv6: prefix if present
	if strings.HasPrefix(ip, "IPv6:") {
		ip = ip[5:]
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		result.addError("domain", "Invalid IP address format", ip)
	}
}

func (v *EmailValidator) checkDomainPolicy(domain string, result *ValidationResult) {
	domain = strings.ToLower(domain)

	for _, blocked := range v.InvalidDomains {
		if strings.EqualFold(domain, blocked) {
			result.addError("domain", "Domain is not allowed", domain)
			return
		}
	}

	if len(v.ValidDomains) > 0 {
		allowed := false
		for _, valid := range v.ValidDomains {
			if strings.EqualFold(domain, valid) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.addError("domain", "Domain is not in allowed list", domain)
		}
	}
}

func (v *EmailValidator) ParseEmailAddress(address string) (*EmailAddress, error) {
	result := v.ValidateAddress(address)
	if !result.Valid {
		return nil, fmt.Errorf("validation failed: %v", result.Errors)
	}

	parts := strings.SplitN(address, "@", 2)
	return &EmailAddress{
		Local:  parts[0],
		Domain: parts[1],
		Raw:    address,
		IsUTF8: !isASCII(address),
	}, nil
}

func (r *ValidationResult) addError(field, message, value string) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func isValidTLD(tld string) bool {
	if len(tld) < 2 {
		return false
	}

	for _, r := range tld {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func (v *EmailValidator) ValidateEmailList(addresses []string) map[string]*ValidationResult {
	results := make(map[string]*ValidationResult)
	for _, addr := range addresses {
		results[addr] = v.ValidateAddress(addr)
	}
	return results
}
