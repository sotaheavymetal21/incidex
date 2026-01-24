package sanitizer

import (
	"encoding/json"
	"regexp"
	"strings"
)

// SanitizeSQL masks sensitive information in SQL queries
func SanitizeSQL(sql string) string {
	// Mask password hashes (bcrypt format: $2a$, $2b$, $2y$)
	bcryptPattern := regexp.MustCompile(`'\$2[aby]\$[^']*'`)
	sql = bcryptPattern.ReplaceAllString(sql, "'[MASKED_HASH]'")

	// Mask email addresses
	emailPattern := regexp.MustCompile(`'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'`)
	sql = emailPattern.ReplaceAllString(sql, "'[MASKED_EMAIL]'")

	// Mask JWT tokens (typically long base64-like strings)
	jwtPattern := regexp.MustCompile(`'eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*'`)
	sql = jwtPattern.ReplaceAllString(sql, "'[MASKED_TOKEN]'")

	// Mask phone numbers (various formats)
	// Matches: +1-234-567-8900, (123) 456-7890, 123-456-7890, 1234567890
	phonePattern := regexp.MustCompile(`'(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}'`)
	sql = phonePattern.ReplaceAllString(sql, "'[MASKED_PHONE]'")

	// Mask credit card numbers (13-19 digits with optional spaces/dashes)
	ccPattern := regexp.MustCompile(`'[0-9]{4}[-\s]?[0-9]{4}[-\s]?[0-9]{4}[-\s]?[0-9]{3,7}'`)
	sql = ccPattern.ReplaceAllString(sql, "'[MASKED_CC]'")

	// Mask Social Security Numbers (SSN format: 123-45-6789)
	ssnPattern := regexp.MustCompile(`'[0-9]{3}-[0-9]{2}-[0-9]{4}'`)
	sql = ssnPattern.ReplaceAllString(sql, "'[MASKED_SSN]'")

	// Mask IPv4 addresses
	ipv4Pattern := regexp.MustCompile(`'(?:[0-9]{1,3}\.){3}[0-9]{1,3}'`)
	sql = ipv4Pattern.ReplaceAllString(sql, "'[MASKED_IP]'")

	// Mask IPv6 addresses
	ipv6Pattern := regexp.MustCompile(`'(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}'`)
	sql = ipv6Pattern.ReplaceAllString(sql, "'[MASKED_IP]'")

	// Mask long strings that look like secrets/tokens (>32 alphanumeric chars)
	secretPattern := regexp.MustCompile(`'[a-zA-Z0-9_-]{32,}'`)
	sql = secretPattern.ReplaceAllString(sql, "'[MASKED_SECRET]'")

	// Mask API keys (common patterns: starts with sk_, pk_, api_, key_)
	apiKeyPattern := regexp.MustCompile(`'(?:sk_|pk_|api_|key_)[a-zA-Z0-9_-]+'`)
	sql = apiKeyPattern.ReplaceAllString(sql, "'[MASKED_API_KEY]'")

	return sql
}

// SanitizeJSON masks sensitive fields in JSON data
func SanitizeJSON(body string) string {
	if body == "" {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// If not valid JSON, return as-is
		return body
	}

	sanitizeMap(data)

	sanitized, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return string(sanitized)
}

// sanitizeMap recursively sanitizes sensitive fields in a map
func sanitizeMap(data map[string]interface{}) {
	sensitiveKeys := []string{
		// Authentication & Authorization
		"password", "old_password", "new_password", "confirm_password",
		"token", "access_token", "refresh_token", "api_token", "auth_token",
		"secret", "client_secret", "api_secret", "jwt_secret",
		"api_key", "apikey", "key", "private_key", "public_key",
		"authorization", "auth",

		// Personal Information
		"ssn", "social_security", "social_security_number",
		"credit_card", "creditcard", "card_number", "cvv", "cvc",
		"pin", "pin_code",
		"passport", "passport_number",
		"drivers_license", "driver_license",

		// Network & System
		"ip_address", "ip", "ipaddress",
		"mac_address", "mac",

		// Database & Infrastructure
		"database_url", "db_url", "connection_string",
		"redis_url", "redis_password",
		"minio_secret_key", "aws_secret_access_key",

		// Payment & Financial
		"bank_account", "account_number", "routing_number",
		"iban", "swift", "bic",
	}

	for key, value := range data {
		lowerKey := strings.ToLower(key)

		// Check if this key should be sanitized
		shouldSanitize := false
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitiveKey) {
				shouldSanitize = true
				break
			}
		}

		if shouldSanitize {
			data[key] = "***REDACTED***"
			continue
		}

		// Recursively sanitize nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			sanitizeMap(nestedMap)
		}

		// Recursively sanitize arrays of maps
		if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					sanitizeMap(nestedMap)
					arr[i] = nestedMap
				}
			}
		}
	}
}

// SanitizeEmail partially masks an email address
// Example: user@example.com -> u***@example.com
func SanitizeEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "[INVALID_EMAIL]"
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 1 {
		return localPart + "***@" + domain
	}

	return string(localPart[0]) + "***@" + domain
}

// SanitizeIP masks an IP address for GDPR compliance
// Example: 192.168.1.100 -> 192.168.xxx.xxx
func SanitizeIP(ip string) string {
	if ip == "" {
		return ""
	}

	// IPv4
	if strings.Contains(ip, ".") {
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			return parts[0] + "." + parts[1] + ".xxx.xxx"
		}
	}

	// IPv6
	if strings.Contains(ip, ":") {
		parts := strings.Split(ip, ":")
		if len(parts) >= 2 {
			return parts[0] + ":" + parts[1] + ":xxxx:xxxx:xxxx:xxxx"
		}
	}

	return "[MASKED_IP]"
}

// SanitizePhoneNumber partially masks a phone number
// Example: +1-234-567-8900 -> +1-234-xxx-xxxx
func SanitizePhoneNumber(phone string) string {
	if phone == "" {
		return ""
	}

	// Extract only digits
	digits := regexp.MustCompile(`[0-9]`).FindAllString(phone, -1)
	if len(digits) < 6 {
		return "[MASKED_PHONE]"
	}

	// Keep first 3-4 digits, mask the rest
	keepDigits := 3
	if len(digits) > 10 {
		keepDigits = 4 // Country code
	}

	prefix := strings.Join(digits[:keepDigits], "")
	return prefix + "-xxx-xxxx"
}

// SanitizeCreditCard masks a credit card number
// Example: 4111-1111-1111-1111 -> 4111-xxxx-xxxx-1111
func SanitizeCreditCard(cc string) string {
	if cc == "" {
		return ""
	}

	// Extract only digits
	digits := regexp.MustCompile(`[0-9]`).FindAllString(cc, -1)
	if len(digits) < 13 || len(digits) > 19 {
		return "[MASKED_CC]"
	}

	// Keep first 4 and last 4 digits
	first4 := strings.Join(digits[:4], "")
	last4 := strings.Join(digits[len(digits)-4:], "")

	return first4 + "-xxxx-xxxx-" + last4
}
