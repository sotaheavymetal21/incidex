package sanitizer

import (
	"encoding/json"
	"regexp"
	"strings"
)

// SanitizeSQL は SQL クエリ内の機密情報をマスクします
func SanitizeSQL(sql string) string {
	// パスワード hash をマスクします（bcrypt 形式: $2a$, $2b$, $2y$）
	bcryptPattern := regexp.MustCompile(`'\$2[aby]\$[^']*'`)
	sql = bcryptPattern.ReplaceAllString(sql, "'[MASKED_HASH]'")

	// メールアドレスをマスクします
	emailPattern := regexp.MustCompile(`'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'`)
	sql = emailPattern.ReplaceAllString(sql, "'[MASKED_EMAIL]'")

	// JWT token をマスクします（通常は長い base64 風の文字列）
	jwtPattern := regexp.MustCompile(`'eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*'`)
	sql = jwtPattern.ReplaceAllString(sql, "'[MASKED_TOKEN]'")

	// 電話番号をマスクします（様々な形式）
	// マッチ対象: +1-234-567-8900, (123) 456-7890, 123-456-7890, 1234567890
	phonePattern := regexp.MustCompile(`'(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}'`)
	sql = phonePattern.ReplaceAllString(sql, "'[MASKED_PHONE]'")

	// クレジットカード番号をマスクします（13-19桁、オプションでスペース/ダッシュ付き）
	ccPattern := regexp.MustCompile(`'[0-9]{4}[-\s]?[0-9]{4}[-\s]?[0-9]{4}[-\s]?[0-9]{3,7}'`)
	sql = ccPattern.ReplaceAllString(sql, "'[MASKED_CC]'")

	// 社会保障番号をマスクします（SSN 形式: 123-45-6789）
	ssnPattern := regexp.MustCompile(`'[0-9]{3}-[0-9]{2}-[0-9]{4}'`)
	sql = ssnPattern.ReplaceAllString(sql, "'[MASKED_SSN]'")

	// IPv4 アドレスをマスクします
	ipv4Pattern := regexp.MustCompile(`'(?:[0-9]{1,3}\.){3}[0-9]{1,3}'`)
	sql = ipv4Pattern.ReplaceAllString(sql, "'[MASKED_IP]'")

	// IPv6 アドレスをマスクします
	ipv6Pattern := regexp.MustCompile(`'(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}'`)
	sql = ipv6Pattern.ReplaceAllString(sql, "'[MASKED_IP]'")

	// シークレット/token に見える長い文字列をマスクします（32文字以上の英数字）
	secretPattern := regexp.MustCompile(`'[a-zA-Z0-9_-]{32,}'`)
	sql = secretPattern.ReplaceAllString(sql, "'[MASKED_SECRET]'")

	// API キーをマスクします（一般的なパターン: sk_, pk_, api_, key_ で始まる）
	apiKeyPattern := regexp.MustCompile(`'(?:sk_|pk_|api_|key_)[a-zA-Z0-9_-]+'`)
	sql = apiKeyPattern.ReplaceAllString(sql, "'[MASKED_API_KEY]'")

	return sql
}

// SanitizeJSON は JSON データ内の機密フィールドをマスクします
func SanitizeJSON(body string) string {
	if body == "" {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// 有効な JSON でない場合はそのまま返します
		return body
	}

	sanitizeMap(data)

	sanitized, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return string(sanitized)
}

// sanitizeMap はマップ内の機密フィールドを再帰的にサニタイズします
func sanitizeMap(data map[string]interface{}) {
	sensitiveKeys := []string{
		// 認証・認可
		"password", "old_password", "new_password", "confirm_password",
		"token", "access_token", "refresh_token", "api_token", "auth_token",
		"secret", "client_secret", "api_secret", "jwt_secret",
		"api_key", "apikey", "key", "private_key", "public_key",
		"authorization", "auth",

		// 個人情報
		"ssn", "social_security", "social_security_number",
		"credit_card", "creditcard", "card_number", "cvv", "cvc",
		"pin", "pin_code",
		"passport", "passport_number",
		"drivers_license", "driver_license",

		// ネットワーク・システム
		"ip_address", "ip", "ipaddress",
		"mac_address", "mac",

		// データベース・インフラストラクチャ
		"database_url", "db_url", "connection_string",
		"redis_url", "redis_password",
		"minio_secret_key", "aws_secret_access_key",

		// 決済・金融
		"bank_account", "account_number", "routing_number",
		"iban", "swift", "bic",
	}

	for key, value := range data {
		lowerKey := strings.ToLower(key)

		// このキーをサニタイズすべきか確認します
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

		// ネストされたマップを再帰的にサニタイズします
		if nestedMap, ok := value.(map[string]interface{}); ok {
			sanitizeMap(nestedMap)
		}

		// マップの配列を再帰的にサニタイズします
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

// SanitizeEmail はメールアドレスを部分的にマスクします
// 例: user@example.com -> u***@example.com
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

// SanitizeIP は GDPR 準拠のため IP アドレスをマスクします
// 例: 192.168.1.100 -> 192.168.xxx.xxx
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

// SanitizePhoneNumber は電話番号を部分的にマスクします
// 例: +1-234-567-8900 -> +1-234-xxx-xxxx
func SanitizePhoneNumber(phone string) string {
	if phone == "" {
		return ""
	}

	// 数字のみを抽出します
	digits := regexp.MustCompile(`[0-9]`).FindAllString(phone, -1)
	if len(digits) < 6 {
		return "[MASKED_PHONE]"
	}

	// 最初の3-4桁を保持し、残りをマスクします
	keepDigits := 3
	if len(digits) > 10 {
		keepDigits = 4 // 国コード
	}

	prefix := strings.Join(digits[:keepDigits], "")
	return prefix + "-xxx-xxxx"
}

// SanitizeCreditCard はクレジットカード番号をマスクします
// 例: 4111-1111-1111-1111 -> 4111-xxxx-xxxx-1111
func SanitizeCreditCard(cc string) string {
	if cc == "" {
		return ""
	}

	// 数字のみを抽出します
	digits := regexp.MustCompile(`[0-9]`).FindAllString(cc, -1)
	if len(digits) < 13 || len(digits) > 19 {
		return "[MASKED_CC]"
	}

	// 最初の4桁と最後の4桁を保持します
	first4 := strings.Join(digits[:4], "")
	last4 := strings.Join(digits[len(digits)-4:], "")

	return first4 + "-xxxx-xxxx-" + last4
}
