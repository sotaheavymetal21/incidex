package validator

import "errors"

// Validation errors for user-related fields
var (
	ErrNameRequired            = errors.New("name is required")
	ErrNameTooLong             = errors.New("name is too long")
	ErrNameContainsEmoji       = errors.New("name cannot contain emoji characters")
	ErrEmailRequired           = errors.New("email is required")
	ErrEmailTooLong            = errors.New("email is too long")
	ErrInvalidEmail            = errors.New("invalid email format")
	ErrPasswordRequired        = errors.New("password is required")
	ErrPasswordTooShort        = errors.New("password is too short")
	ErrPasswordNeedsUppercase  = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNeedsLowercase  = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNeedsDigit      = errors.New("password must contain at least one digit")
	ErrEmployeeNumberTooLong   = errors.New("employee number is too long")
	ErrInvalidEmployeeNumber   = errors.New("employee number can only contain alphanumeric characters and hyphens")
	ErrDepartmentTooLong       = errors.New("department name is too long")
	ErrDepartmentContainsEmoji = errors.New("department name cannot contain emoji characters")
)

// Validation errors for incident-related fields
var (
	ErrTitleRequired        = errors.New("title is required")
	ErrTitleTooLong         = errors.New("title is too long")
	ErrDescriptionRequired  = errors.New("description is required")
	ErrDescriptionTooLong   = errors.New("description is too long")
	ErrImpactScopeTooLong   = errors.New("impact scope is too long")
	ErrInvalidSeverity      = errors.New("invalid severity value")
	ErrInvalidStatus        = errors.New("invalid status value")
)

// Validation errors for tag-related fields
var (
	ErrTagNameRequired       = errors.New("tag name is required")
	ErrTagNameTooLong        = errors.New("tag name is too long")
	ErrTagNameContainsEmoji  = errors.New("tag name cannot contain emoji characters")
	ErrInvalidColorFormat    = errors.New("invalid color format (expected #RRGGBB)")
)

// Validation errors for comment/activity fields
var (
	ErrCommentRequired    = errors.New("comment is required")
	ErrCommentTooLong     = errors.New("comment is too long")
	ErrInvalidEventType   = errors.New("invalid event type")
)

// Validation errors for post-mortem fields
var (
	ErrPostMortemTextTooLong = errors.New("post-mortem text is too long")
)

// Validation errors for action item fields
var (
	ErrActionItemTitleRequired  = errors.New("action item title is required")
	ErrActionItemTitleTooLong   = errors.New("action item title is too long")
	ErrActionItemDescTooLong    = errors.New("action item description is too long")
	ErrActionItemLinksTooLong   = errors.New("action item related links is too long")
	ErrInvalidPriority          = errors.New("invalid priority value")
	ErrInvalidActionStatus      = errors.New("invalid action status value")
)

// Validation errors for notification settings
var (
	ErrSlackWebhookTooLong  = errors.New("slack webhook URL is too long")
	ErrInvalidSlackWebhook  = errors.New("invalid slack webhook URL format")
)

// General validation errors
var (
	ErrContainsDangerousChars = errors.New("input contains dangerous characters")
)
