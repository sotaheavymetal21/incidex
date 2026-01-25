package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedFileType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fileName string
		expected bool
	}{
		// Images
		{"jpg is allowed", "test.jpg", true},
		{"jpeg is allowed", "test.jpeg", true},
		{"png is allowed", "test.png", true},
		{"gif is allowed", "test.gif", true},
		{"webp is allowed", "test.webp", true},

		// PDF
		{"pdf is allowed", "document.pdf", true},

		// Text files
		{"txt is allowed", "readme.txt", true},
		{"log is allowed", "error.log", true},
		{"md is allowed", "README.md", true},

		// Config files
		{"json is allowed", "config.json", true},
		{"xml is allowed", "data.xml", true},
		{"yaml is allowed", "config.yaml", true},
		{"yml is allowed", "config.yml", true},

		// Archives
		{"zip is allowed", "archive.zip", true},
		{"tar is allowed", "archive.tar", true},
		{"gz is allowed", "archive.gz", true},

		// Not allowed
		{"exe is not allowed", "program.exe", false},
		{"sh is not allowed", "script.sh", false},
		{"js is not allowed", "script.js", false},
		{"html is not allowed", "page.html", false},
		{"php is not allowed", "page.php", false},
		{"no extension is not allowed", "noextension", false},

		// Case insensitive
		{"JPG uppercase is allowed", "test.JPG", true},
		{"PDF uppercase is allowed", "test.PDF", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isAllowedFileType(tt.fileName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Note: Full integration tests for AttachmentUsecase require MinIO storage mock
// which is complex to set up. The core business logic validation is tested above.
// Handler-level tests with HTTP mocking will provide additional coverage.
