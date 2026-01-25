package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationLimits は各種フィールドの最大長を定義します
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

// Limits はバリデーション制限値を含みます
var Limits = ValidationLimits{
	NameMaxLength:             50,
	EmailMaxLength:            254,
	PasswordMinLength:         8,
	PasswordMinLengthStrict:   12,
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

// ValidateName は名前フィールドをバリデーションします
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

// ValidateEmail はメールフィールドをバリデーションします
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

// ValidatePassword はパスワードフィールドをバリデーションします
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

// ValidateEmployeeNumber は社員番号をバリデーションします
func ValidateEmployeeNumber(employeeNumber string) error {
	if employeeNumber == "" {
		return nil // 任意フィールド
	}
	if utf8.RuneCountInString(employeeNumber) > Limits.EmployeeNumberMaxLength {
		return ErrEmployeeNumberTooLong
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9\-]*$`).MatchString(employeeNumber) {
		return ErrInvalidEmployeeNumber
	}
	return nil
}

// ValidateDepartment は部署フィールドをバリデーションします
func ValidateDepartment(department string) error {
	if department == "" {
		return nil // 任意フィールド
	}
	if utf8.RuneCountInString(department) > Limits.DepartmentMaxLength {
		return ErrDepartmentTooLong
	}
	if ContainsEmoji(department) {
		return ErrDepartmentContainsEmoji
	}
	return nil
}

// ValidateTitle はインシデント/アクションアイテムのタイトルをバリデーションします
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

// ValidateDescription は説明フィールドをバリデーションします
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

// ValidateImpactScope は影響範囲フィールドをバリデーションします
func ValidateImpactScope(impactScope string) error {
	if impactScope == "" {
		return nil // 任意フィールド
	}
	if utf8.RuneCountInString(impactScope) > Limits.ImpactScopeMaxLength {
		return ErrImpactScopeTooLong
	}
	return nil
}

// ValidateComment はコメントフィールドをバリデーションします
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

// ValidateTagName はタグ名フィールドをバリデーションします
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

// ValidateTagColor は16進数カラーコードをバリデーションします
func ValidateTagColor(color string) error {
	if color == "" {
		return nil // デフォルト値を持つ任意フィールド
	}
	if !regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`).MatchString(color) {
		return ErrInvalidColorFormat
	}
	return nil
}

// ValidatePostMortemText はポストモーテムのテキストフィールドをバリデーションします
func ValidatePostMortemText(text string) error {
	if text == "" {
		return nil // 任意フィールド
	}
	if utf8.RuneCountInString(text) > Limits.PostMortemTextMaxLength {
		return ErrPostMortemTextTooLong
	}
	return nil
}

// ValidateActionItemTitle はアクションアイテムのタイトルをバリデーションします
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

// ValidateActionItemDescription はアクションアイテムの説明をバリデーションします
func ValidateActionItemDescription(description string) error {
	if description == "" {
		return nil // 任意フィールド
	}
	if utf8.RuneCountInString(description) > Limits.ActionItemDescMaxLength {
		return ErrActionItemDescTooLong
	}
	return nil
}

// ValidateRelatedLinks は関連リンクフィールドをバリデーションします
func ValidateRelatedLinks(links string) error {
	if links == "" {
		return nil // 任意フィールド
	}
	if utf8.RuneCountInString(links) > Limits.ActionItemLinksMaxLength {
		return ErrActionItemLinksTooLong
	}
	return nil
}

// ValidateSlackWebhook は Slack webhook URL をバリデーションします
func ValidateSlackWebhook(url string) error {
	if url == "" {
		return nil // 任意フィールド
	}
	if utf8.RuneCountInString(url) > Limits.SlackWebhookMaxLength {
		return ErrSlackWebhookTooLong
	}
	if !regexp.MustCompile(`^https://hooks\.slack\.com/services/[A-Z0-9]+/[A-Z0-9]+/[a-zA-Z0-9]+$`).MatchString(url) {
		return ErrInvalidSlackWebhook
	}
	return nil
}

// ヘルパー関数

// IsValidEmail はメールが有効かをチェックします
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ContainsEmoji は文字列に絵文字が含まれているかをチェックします
func ContainsEmoji(s string) bool {
	emojiRanges := []struct {
		start, end rune
	}{
		{0x1F000, 0x1F9FF}, // その他の記号と絵文字
		{0x2600, 0x27BF},   // その他の記号
		{0x1F600, 0x1F64F}, // 顔文字
		{0x1F680, 0x1F6FF}, // 交通と地図の記号
		{0x2702, 0x27B0},   // 装飾記号
		{0xFE00, 0xFE0F},   // 異体字セレクタ
		{0x1F300, 0x1F5FF}, // その他の記号と絵文字
		{0x1F900, 0x1F9FF}, // 補助記号と絵文字
		{0x1FA00, 0x1FA6F}, // チェス記号
		{0x1FA70, 0x1FAFF}, // 記号と絵文字拡張A
		{0x2300, 0x23FF},   // その他の技術記号
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

// ContainsDangerousChars は XSS およびインジェクション脆弱性をチェックします
func ContainsDangerousChars(s string) bool {
	// スクリプトタグ、イベントハンドラ、その他の XSS ベクターをチェックします
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

// ValidateSeverity は重要度の列挙値をバリデーションします
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

// ValidateStatus はステータスの列挙値をバリデーションします
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

// ValidatePriority は優先度の列挙値をバリデーションします
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

// ValidateActionStatus はアクションアイテムステータスの列挙値をバリデーションします
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

// ValidateTimelineEventType はタイムラインイベントタイプの列挙値をバリデーションします
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
