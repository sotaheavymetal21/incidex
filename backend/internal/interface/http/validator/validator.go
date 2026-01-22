package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationLimits defines the maximum lengths for various fields
type ValidationLimits struct {
	NameMaxLength             int
	EmailMaxLength            int
	PasswordMinLength         int
	PasswordMinLengthStrict   int
	EmployeeNumberMaxLength   int
	DepartmentMaxLength       int
	TitleMaxLength            int
	DescriptionMaxLength      int
	ImpactScopeMaxLength      int
	CommentMaxLength          int
	TagNameMaxLength          int
	PostMortemTextMaxLength   int
	ActionItemTitleMaxLength  int
	ActionItemDescMaxLength   int
	ActionItemLinksMaxLength  int
	SlackWebhookMaxLength     int
}

// Limits contains the validation limits
var Limits = ValidationLimits{
	NameMaxLength:             50,
	EmailMaxLength:            254,
	PasswordMinLength:         6,
	PasswordMinLengthStrict:   8,
	EmployeeNumberMaxLength:   20,
	DepartmentMaxLength:       50,
	TitleMaxLength:            500,
	DescriptionMaxLength:      10000,
	ImpactScopeMaxLength:      500,
	CommentMaxLength:          5000,
	TagNameMaxLength:          50,
	PostMortemTextMaxLength:   10000,
	ActionItemTitleMaxLength:  500,
	ActionItemDescMaxLength:   5000,
	ActionItemLinksMaxLength:  2000,
	SlackWebhookMaxLength:     500,
}

// ValidateName validates a name field
func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ErrNameRequired
	}
	if utf8.RuneCountInString(trimmed) > Limits.NameMaxLength {
		return ErrNameTooLong
	}
	if ContainsEmoji(trimmed) {
		return ErrNameContainsEmoji
	}
	if ContainsDangerousChars(trimmed) {
		return ErrContainsDangerousChars
	}
	return nil
}

// ValidateEmail validates an email field
func ValidateEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return ErrEmailRequired
	}
	if utf8.RuneCountInString(trimmed) > Limits.EmailMaxLength {
		return ErrEmailTooLong
	}
	if !IsValidEmail(trimmed) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword validates a password field
func ValidatePassword(password string, strict bool) error {
	if password == "" {
		return ErrPasswordRequired
	}

	minLength := Limits.PasswordMinLength
	if strict {
		minLength = Limits.PasswordMinLengthStrict
	}

	if len(password) < minLength {
		return ErrPasswordTooShort
	}

	if strict {
		if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
			return ErrPasswordNeedsUppercase
		}
		if !regexp.MustCompile(`[a-z]`).MatchString(password) {
			return ErrPasswordNeedsLowercase
		}
		if !regexp.MustCompile(`[0-9]`).MatchString(password) {
			return ErrPasswordNeedsDigit
		}
	}

	return nil
}

// ValidateEmployeeNumber validates an employee number
func ValidateEmployeeNumber(employeeNumber string) error {
	if employeeNumber == "" {
		return nil // Optional field
	}
	if utf8.RuneCountInString(employeeNumber) > Limits.EmployeeNumberMaxLength {
		return ErrEmployeeNumberTooLong
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9\-]*$`).MatchString(employeeNumber) {
		return ErrInvalidEmployeeNumber
	}
	return nil
}

// ValidateDepartment validates a department field
func ValidateDepartment(department string) error {
	if department == "" {
		return nil // Optional field
	}
	if utf8.RuneCountInString(department) > Limits.DepartmentMaxLength {
		return ErrDepartmentTooLong
	}
	if ContainsEmoji(department) {
		return ErrDepartmentContainsEmoji
	}
	return nil
}

// ValidateTitle validates an incident/action item title
func ValidateTitle(title string) error {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return ErrTitleRequired
	}
	if utf8.RuneCountInString(trimmed) > Limits.TitleMaxLength {
		return ErrTitleTooLong
	}
	if ContainsDangerousChars(trimmed) {
		return ErrContainsDangerousChars
	}
	return nil
}

// ValidateDescription validates a description field
func ValidateDescription(description string) error {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return ErrDescriptionRequired
	}
	if utf8.RuneCountInString(trimmed) > Limits.DescriptionMaxLength {
		return ErrDescriptionTooLong
	}
	return nil
}

// ValidateImpactScope validates an impact scope field
func ValidateImpactScope(impactScope string) error {
	if impactScope == "" {
		return nil // Optional field
	}
	if utf8.RuneCountInString(impactScope) > Limits.ImpactScopeMaxLength {
		return ErrImpactScopeTooLong
	}
	return nil
}

// ValidateComment validates a comment field
func ValidateComment(comment string) error {
	trimmed := strings.TrimSpace(comment)
	if trimmed == "" {
		return ErrCommentRequired
	}
	if utf8.RuneCountInString(trimmed) > Limits.CommentMaxLength {
		return ErrCommentTooLong
	}
	return nil
}

// ValidateTagName validates a tag name field
func ValidateTagName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ErrTagNameRequired
	}
	if utf8.RuneCountInString(trimmed) > Limits.TagNameMaxLength {
		return ErrTagNameTooLong
	}
	if ContainsEmoji(trimmed) {
		return ErrTagNameContainsEmoji
	}
	if ContainsDangerousChars(trimmed) {
		return ErrContainsDangerousChars
	}
	return nil
}

// ValidateTagColor validates a hex color code
func ValidateTagColor(color string) error {
	if color == "" {
		return nil // Optional field with default
	}
	if !regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`).MatchString(color) {
		return ErrInvalidColorFormat
	}
	return nil
}

// ValidatePostMortemText validates post-mortem text fields
func ValidatePostMortemText(text string) error {
	if text == "" {
		return nil // Optional field
	}
	if utf8.RuneCountInString(text) > Limits.PostMortemTextMaxLength {
		return ErrPostMortemTextTooLong
	}
	return nil
}

// ValidateActionItemTitle validates an action item title
func ValidateActionItemTitle(title string) error {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return ErrActionItemTitleRequired
	}
	if utf8.RuneCountInString(trimmed) > Limits.ActionItemTitleMaxLength {
		return ErrActionItemTitleTooLong
	}
	return nil
}

// ValidateActionItemDescription validates an action item description
func ValidateActionItemDescription(description string) error {
	if description == "" {
		return nil // Optional field
	}
	if utf8.RuneCountInString(description) > Limits.ActionItemDescMaxLength {
		return ErrActionItemDescTooLong
	}
	return nil
}

// ValidateRelatedLinks validates related links field
func ValidateRelatedLinks(links string) error {
	if links == "" {
		return nil // Optional field
	}
	if utf8.RuneCountInString(links) > Limits.ActionItemLinksMaxLength {
		return ErrActionItemLinksTooLong
	}
	return nil
}

// ValidateSlackWebhook validates a Slack webhook URL
func ValidateSlackWebhook(url string) error {
	if url == "" {
		return nil // Optional field
	}
	if utf8.RuneCountInString(url) > Limits.SlackWebhookMaxLength {
		return ErrSlackWebhookTooLong
	}
	if !regexp.MustCompile(`^https://hooks\.slack\.com/services/[A-Z0-9]+/[A-Z0-9]+/[a-zA-Z0-9]+$`).MatchString(url) {
		return ErrInvalidSlackWebhook
	}
	return nil
}

// Helper functions

// IsValidEmail checks if an email is valid
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ContainsEmoji checks if a string contains emoji characters
func ContainsEmoji(s string) bool {
	emojiRanges := []struct {
		start, end rune
	}{
		{0x1F000, 0x1F9FF}, // Miscellaneous Symbols and Pictographs
		{0x2600, 0x27BF},   // Miscellaneous Symbols
		{0x1F600, 0x1F64F}, // Emoticons
		{0x1F680, 0x1F6FF}, // Transport and Map Symbols
		{0x2702, 0x27B0},   // Dingbats
		{0xFE00, 0xFE0F},   // Variation Selectors
		{0x1F300, 0x1F5FF}, // Miscellaneous Symbols and Pictographs
		{0x1F900, 0x1F9FF}, // Supplemental Symbols and Pictographs
		{0x1FA00, 0x1FA6F}, // Chess Symbols
		{0x1FA70, 0x1FAFF}, // Symbols and Pictographs Extended-A
		{0x2300, 0x23FF},   // Miscellaneous Technical
	}

	for _, r := range s {
		for _, emojiRange := range emojiRanges {
			if r >= emojiRange.start && r <= emojiRange.end {
				return true
			}
		}
	}
	return false
}

// ContainsDangerousChars checks for XSS and injection vulnerabilities
func ContainsDangerousChars(s string) bool {
	// Check for script tags, event handlers, and other XSS vectors
	dangerousPatterns := []string{
		`<script`,
		`javascript:`,
		`onerror=`,
		`onload=`,
		`onclick=`,
		`onmouseover=`,
		`onfocus=`,
		`onblur=`,
		`data:text/html`,
		`vbscript:`,
	}

	lower := strings.ToLower(s)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// ValidateSeverity validates severity enum
func ValidateSeverity(severity string) error {
	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}
	if !validSeverities[severity] {
		return ErrInvalidSeverity
	}
	return nil
}

// ValidateStatus validates status enum
func ValidateStatus(status string) error {
	validStatuses := map[string]bool{
		"open":          true,
		"investigating": true,
		"resolved":      true,
		"closed":        true,
	}
	if !validStatuses[status] {
		return ErrInvalidStatus
	}
	return nil
}

// ValidatePriority validates priority enum
func ValidatePriority(priority string) error {
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}
	if !validPriorities[priority] {
		return ErrInvalidPriority
	}
	return nil
}

// ValidateActionStatus validates action item status enum
func ValidateActionStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
	}
	if !validStatuses[status] {
		return ErrInvalidActionStatus
	}
	return nil
}

// ValidateTimelineEventType validates timeline event type enum
func ValidateTimelineEventType(eventType string) error {
	validTypes := map[string]bool{
		"detected":                true,
		"investigation_started":   true,
		"root_cause_identified":   true,
		"mitigation":              true,
		"timeline_resolved":       true,
		"other":                   true,
	}
	if !validTypes[eventType] {
		return ErrInvalidEventType
	}
	return nil
}
