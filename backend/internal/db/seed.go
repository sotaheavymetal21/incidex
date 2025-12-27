package db

import (
	"context"
	"fmt"
	"incidex/internal/domain"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed populates the database with test data
func Seed(db *gorm.DB) error {
	log.Println("Starting database seeding...")
	// Check if data already exists
	var userCount int64
	db.Model(&domain.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("Database already contains data. Skipping seed.")
		return nil
	}

	ctx := context.Background()

	// Create users
	users, err := seedUsers(db, ctx)
	if err != nil {
		return err
	}

	// Create tags
	tags, err := seedTags(db)
	if err != nil {
		return err
	}

	// Create incidents
	if err := seedIncidents(db, ctx, users, tags); err != nil {
		return err
	}

	log.Println("Database seeding completed successfully!")
	return nil
}

func seedUsers(db *gorm.DB, ctx context.Context) ([]*domain.User, error) {
	log.Println("Seeding users...")

	users := []*domain.User{
		{
			Email: "admin@example.com",
			Name:  "管理者ユーザー",
			Role:  domain.RoleAdmin,
		},
		{
			Email: "editor1@example.com",
			Name:  "編集者 太郎",
			Role:  domain.RoleEditor,
		},
		{
			Email: "editor2@example.com",
			Name:  "編集者 花子",
			Role:  domain.RoleEditor,
		},
		{
			Email: "viewer1@example.com",
			Name:  "閲覧者 一郎",
			Role:  domain.RoleViewer,
		},
		{
			Email: "viewer2@example.com",
			Name:  "閲覧者 二郎",
			Role:  domain.RoleViewer,
		},
	}

	// Get test user password from environment variable
	// TEST_USER_PASSWORD環境変数が必須です
	testUserPassword := os.Getenv("TEST_USER_PASSWORD")
	if testUserPassword == "" {
		return nil, fmt.Errorf("TEST_USER_PASSWORD environment variable is required")
	}

	// Hash password for all test users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testUserPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.PasswordHash = string(hashedPassword)
		if err := db.WithContext(ctx).Create(user).Error; err != nil {
			return nil, err
		}
		log.Printf("Created user: %s (%s)", user.Email, user.Role)
	}

	return users, nil
}

func seedTags(db *gorm.DB) ([]*domain.Tag, error) {
	log.Println("Seeding tags...")

	tags := []*domain.Tag{
		{Name: "ネットワーク障害", Color: "#ef4444"}, // red
		{Name: "サーバー障害", Color: "#f97316"},   // orange
		{Name: "アプリケーション", Color: "#3b82f6"}, // blue
		{Name: "データベース", Color: "#8b5cf6"},   // purple
		{Name: "セキュリティ", Color: "#dc2626"},   // dark red
		{Name: "パフォーマンス", Color: "#f59e0b"},  // amber
		{Name: "メンテナンス", Color: "#10b981"},   // green
		{Name: "ユーザー報告", Color: "#06b6d4"},   // cyan
		{Name: "バグ", Color: "#ec4899"},       // pink
		{Name: "設定変更", Color: "#6366f1"},     // indigo
	}

	for _, tag := range tags {
		if err := db.Create(tag).Error; err != nil {
			return nil, err
		}
		log.Printf("Created tag: %s", tag.Name)
	}

	return tags, nil
}

func seedIncidents(db *gorm.DB, ctx context.Context, users []*domain.User, tags []*domain.Tag) error {
	log.Println("Seeding incidents...")

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)
	threeDaysAgo := now.Add(-72 * time.Hour)
	fourDaysAgo := now.Add(-96 * time.Hour)
	fiveDaysAgo := now.Add(-120 * time.Hour)
	oneWeekAgo := now.Add(-7 * 24 * time.Hour)
	resolvedTime := now.Add(-2 * time.Hour)
	resolvedYesterday := yesterday.Add(-3 * time.Hour)
	resolvedTwoDaysAgo := twoDaysAgo.Add(-1 * time.Hour)

	incidents := []*domain.Incident{
		{
			Title:       "本番環境でのデータベース接続エラー",
			Description: "本番環境のアプリケーションサーバーから、メインデータベースへの接続が間欠的に失敗しています。接続タイムアウトエラーが多数発生しており、ユーザーに影響が出ています。",
			Summary:     "DB接続エラーによりサービスが断続的に利用不可",
			Severity:    domain.SeverityCritical,
			Status:      domain.StatusInvestigating,
			ImpactScope: "全ユーザー（推定5000人）に影響",
			DetectedAt:  now.Add(-3 * time.Hour),
			CreatorID:   users[1].ID,
			AssigneeID:  &users[0].ID,
		},
		{
			Title:       "APIレスポンスタイムの著しい遅延",
			Description: "午前9時頃から、ユーザーAPIのレスポンスタイムが通常の10倍以上に増加しています。特に検索機能に影響が大きく出ています。データベースクエリの最適化が必要と思われます。",
			Summary:     "APIレスポンスが大幅に遅延、検索機能に影響大",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusOpen,
			ImpactScope: "検索機能を使用する約30%のユーザーに影響",
			DetectedAt:  now.Add(-5 * time.Hour),
			CreatorID:   users[2].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "ログイン画面でのCSRF検証エラー",
			Description: "一部のユーザーから、ログイン画面でCSRFトークンの検証エラーが発生しているとの報告がありました。キャッシュの設定ミスが原因と思われます。",
			Summary:     "ログイン時のCSRFエラーが散発的に発生",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusResolved,
			ImpactScope: "約50名のユーザーから報告あり",
			DetectedAt:  yesterday,
			ResolvedAt:  &resolvedTime,
			CreatorID:   users[3].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "ダッシュボードのグラフ表示が崩れる問題",
			Description: "特定のブラウザ（Firefox 120以降）で、ダッシュボードのグラフが正しく表示されない問題が報告されています。CSSの互換性問題の可能性があります。",
			Summary:     "Firefox最新版でグラフ表示に不具合",
			Severity:    domain.SeverityLow,
			Status:      domain.StatusOpen,
			ImpactScope: "Firefox利用者（全体の約15%）に影響",
			DetectedAt:  twoDaysAgo,
			CreatorID:   users[4].ID,
			AssigneeID:  &users[2].ID,
		},
		{
			Title:       "不正アクセスの試行を検知",
			Description: "複数のIPアドレスから、パスワード総当たり攻撃と思われるアクセスが検知されました。現在はレート制限により影響は限定的ですが、監視を強化する必要があります。",
			Summary:     "ブルートフォース攻撃を検知、レート制限で対応中",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusInvestigating,
			ImpactScope: "現時点でユーザーへの直接的な影響なし",
			DetectedAt:  now.Add(-1 * time.Hour),
			CreatorID:   users[0].ID,
			AssigneeID:  &users[0].ID,
		},
		{
			Title:       "定期メンテナンス：インデックス再構築",
			Description: "データベースのパフォーマンス改善のため、インデックスの再構築を実施します。メンテナンスウィンドウは深夜2時〜4時を予定しています。",
			Summary:     "定期メンテナンス作業の実施予定",
			Severity:    domain.SeverityLow,
			Status:      domain.StatusOpen,
			ImpactScope: "メンテナンス時間中は一部機能が制限される可能性",
			DetectedAt:  now,
			CreatorID:   users[1].ID,
		},
		{
			Title:       "画像アップロード機能の間欠的なエラー",
			Description: "ユーザープロフィールの画像アップロード時に、ファイルサイズに関わらずエラーが発生するケースが報告されています。ストレージサービスとの通信に問題がある可能性があります。",
			Summary:     "画像アップロードが間欠的に失敗",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusOpen,
			ImpactScope: "画像アップロードを試みる約10%のユーザーに影響",
			DetectedAt:  yesterday,
			CreatorID:   users[2].ID,
		},
		{
			Title:       "メール通知の送信遅延",
			Description: "パスワードリセットメールやアラート通知メールの配信が大幅に遅延しています。メール送信キューに問題があり、調査中です。",
			Summary:     "メール通知が最大30分遅延",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusClosed,
			ImpactScope: "全ユーザーのメール通知機能に影響",
			DetectedAt:  oneWeekAgo,
			ResolvedAt:  &yesterday,
			CreatorID:   users[0].ID,
			AssigneeID:  &users[1].ID,
		},
		// 追加のインシデント（9-30）
		{
			Title:       "CDNキャッシュの不整合",
			Description: "静的コンテンツのCDNキャッシュが古い状態で配信されており、CSSやJavaScriptの最新版が反映されていません。",
			Summary:     "CDNキャッシュ問題で古いコンテンツが表示",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusResolved,
			ImpactScope: "全ユーザーに影響（視覚的な不具合）",
			DetectedAt:  twoDaysAgo,
			ResolvedAt:  &resolvedYesterday,
			CreatorID:   users[1].ID,
			AssigneeID:  &users[2].ID,
		},
		{
			Title:       "Webサーバーのメモリリーク",
			Description: "アプリケーションサーバーのメモリ使用量が徐々に増加し、定期的な再起動が必要になっています。メモリリークの調査が必要です。",
			Summary:     "サーバーメモリ使用量が継続的に増加",
			Severity:    domain.SeverityCritical,
			Status:      domain.StatusInvestigating,
			ImpactScope: "全サービスの安定性に影響",
			DetectedAt:  threeDaysAgo,
			CreatorID:   users[0].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "決済処理の二重実行",
			Description: "一部のユーザーから決済が二重に実行されたとの報告がありました。決済フロー内の冪等性チェックに問題があるようです。",
			Summary:     "決済が重複実行される不具合",
			Severity:    domain.SeverityCritical,
			Status:      domain.StatusResolved,
			ImpactScope: "決済機能を使用した約5名のユーザーに影響",
			DetectedAt:  fiveDaysAgo,
			ResolvedAt:  &resolvedTwoDaysAgo,
			CreatorID:   users[2].ID,
			AssigneeID:  &users[0].ID,
		},
		{
			Title:       "検索結果の精度低下",
			Description: "全文検索エンジンのインデックスが破損しており、検索結果の精度が著しく低下しています。再インデックスが必要です。",
			Summary:     "検索機能の精度が大幅に低下",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusOpen,
			ImpactScope: "検索機能利用者に影響",
			DetectedAt:  yesterday,
			CreatorID:   users[3].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "モバイルアプリのクラッシュ頻発",
			Description: "iOS版アプリで起動時のクラッシュが頻発しています。最新のOSアップデート後に発生し始めました。",
			Summary:     "iOSアプリが起動時にクラッシュ",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusOpen,
			ImpactScope: "iOS 18ユーザー約200名に影響",
			DetectedAt:  now.Add(-6 * time.Hour),
			CreatorID:   users[4].ID,
			AssigneeID:  &users[2].ID,
		},
		{
			Title:       "SSL証明書の期限切れアラート",
			Description: "開発環境のSSL証明書が1週間後に期限切れを迎えます。更新作業が必要です。",
			Summary:     "開発環境SSL証明書の更新が必要",
			Severity:    domain.SeverityLow,
			Status:      domain.StatusOpen,
			ImpactScope: "開発環境のみ（本番環境への影響なし）",
			DetectedAt:  now,
			CreatorID:   users[1].ID,
		},
		{
			Title:       "バックアップジョブの失敗",
			Description: "過去3日間、日次バックアップジョブが失敗しています。ストレージ容量不足が原因と思われます。",
			Summary:     "日次バックアップが3日連続で失敗",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusInvestigating,
			ImpactScope: "データ保護に影響（緊急時の復旧リスク）",
			DetectedAt:  threeDaysAgo,
			CreatorID:   users[0].ID,
			AssigneeID:  &users[0].ID,
		},
		{
			Title:       "ログ出力の過剰増加",
			Description: "アプリケーションログの出力量が通常の10倍に増加し、ディスク容量を圧迫しています。デバッグログの無効化が必要です。",
			Summary:     "ログファイルが異常に肥大化",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusResolved,
			ImpactScope: "ディスク容量に影響",
			DetectedAt:  fourDaysAgo,
			ResolvedAt:  &resolvedTwoDaysAgo,
			CreatorID:   users[1].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "APIレート制限の誤動作",
			Description: "正規ユーザーのAPIリクエストがレート制限によって拒否されるケースが報告されています。制限値の見直しが必要です。",
			Summary:     "レート制限が正規ユーザーをブロック",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusOpen,
			ImpactScope: "API頻繁利用者約30名に影響",
			DetectedAt:  yesterday,
			CreatorID:   users[2].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "データ同期の遅延",
			Description: "マスターデータベースからレプリカへの同期が大幅に遅延しており、読み取り専用クエリで古いデータが返されています。",
			Summary:     "DBレプリケーション遅延が発生",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusInvestigating,
			ImpactScope: "レポート機能で古いデータが表示",
			DetectedAt:  now.Add(-4 * time.Hour),
			CreatorID:   users[0].ID,
			AssigneeID:  &users[0].ID,
		},
		{
			Title:       "ファイルダウンロード機能のタイムアウト",
			Description: "大きなファイルのダウンロード時にタイムアウトエラーが発生しています。プロキシ設定の調整が必要です。",
			Summary:     "大容量ファイルのダウンロードが失敗",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusOpen,
			ImpactScope: "大容量ファイル利用者約15%に影響",
			DetectedAt:  twoDaysAgo,
			CreatorID:   users[3].ID,
		},
		{
			Title:       "ユーザー認証の突然のログアウト",
			Description: "操作中に突然ログアウトされる現象が報告されています。セッション管理に問題がある可能性があります。",
			Summary:     "セッション切れで突然ログアウト",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusOpen,
			ImpactScope: "約100名のユーザーから報告",
			DetectedAt:  yesterday,
			CreatorID:   users[4].ID,
			AssigneeID:  &users[2].ID,
		},
		{
			Title:       "管理画面の表示エラー",
			Description: "管理画面の一部ページで500エラーが発生しています。権限チェックのロジックに不具合があるようです。",
			Summary:     "管理画面で500エラーが発生",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusResolved,
			ImpactScope: "管理者ユーザー約10名に影響",
			DetectedAt:  threeDaysAgo,
			ResolvedAt:  &resolvedYesterday,
			CreatorID:   users[0].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "通知プッシュの未配信",
			Description: "モバイルアプリへのプッシュ通知が一部のユーザーに届いていません。通知サービスのトークン管理に問題がありそうです。",
			Summary:     "プッシュ通知が一部ユーザーに未配信",
			Severity:    domain.SeverityLow,
			Status:      domain.StatusOpen,
			ImpactScope: "モバイルユーザーの約5%に影響",
			DetectedAt:  twoDaysAgo,
			CreatorID:   users[2].ID,
		},
		{
			Title:       "外部API連携の認証エラー",
			Description: "サードパーティAPIとの連携で認証エラーが頻発しています。APIキーのローテーションが必要です。",
			Summary:     "外部API認証エラーで連携失敗",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusInvestigating,
			ImpactScope: "外部連携機能利用者に影響",
			DetectedAt:  yesterday,
			CreatorID:   users[1].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "キャッシュサーバーの接続切断",
			Description: "Redisキャッシュサーバーへの接続が間欠的に切断されています。ネットワーク設定の確認が必要です。",
			Summary:     "Redisへの接続が不安定",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusOpen,
			ImpactScope: "パフォーマンス全体に影響",
			DetectedAt:  now.Add(-2 * time.Hour),
			CreatorID:   users[0].ID,
			AssigneeID:  &users[0].ID,
		},
		{
			Title:       "CSVエクスポート機能の文字化け",
			Description: "データをCSV形式でエクスポートすると、日本語が文字化けする問題が報告されています。文字コードの設定が必要です。",
			Summary:     "CSVエクスポート時に日本語が文字化け",
			Severity:    domain.SeverityLow,
			Status:      domain.StatusOpen,
			ImpactScope: "CSVエクスポート機能利用者に影響",
			DetectedAt:  threeDaysAgo,
			CreatorID:   users[3].ID,
			AssigneeID:  &users[2].ID,
		},
		{
			Title:       "定期ジョブの実行時刻ズレ",
			Description: "夜間バッチ処理の実行時刻が徐々にずれており、営業時間中に実行される事態が発生しています。",
			Summary:     "バッチジョブのスケジュールが不安定",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusResolved,
			ImpactScope: "バッチ処理のタイミングに影響",
			DetectedAt:  oneWeekAgo,
			ResolvedAt:  &resolvedTwoDaysAgo,
			CreatorID:   users[1].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "Webhook配信の失敗",
			Description: "Webhook通知の配信成功率が通常の80%に低下しています。リトライ機構の改善が必要です。",
			Summary:     "Webhook配信の失敗率が上昇",
			Severity:    domain.SeverityMedium,
			Status:      domain.StatusOpen,
			ImpactScope: "Webhook利用中の連携先に影響",
			DetectedAt:  yesterday,
			CreatorID:   users[2].ID,
			AssigneeID:  &users[1].ID,
		},
		{
			Title:       "パスワードリセット機能の無効化",
			Description: "パスワードリセットメールのリンクをクリックしてもエラーになる問題が報告されています。トークン検証ロジックの不具合です。",
			Summary:     "パスワードリセットリンクが機能しない",
			Severity:    domain.SeverityCritical,
			Status:      domain.StatusResolved,
			ImpactScope: "パスワード忘れユーザー約20名に影響",
			DetectedAt:  fourDaysAgo,
			ResolvedAt:  &resolvedTwoDaysAgo,
			CreatorID:   users[4].ID,
			AssigneeID:  &users[0].ID,
		},
		{
			Title:       "監視アラートの誤検知",
			Description: "システム監視ツールが正常動作中にも関わらずアラートを頻発しています。閾値の調整が必要です。",
			Summary:     "監視システムが過剰にアラート発報",
			Severity:    domain.SeverityLow,
			Status:      domain.StatusOpen,
			ImpactScope: "運用チームの作業効率に影響",
			DetectedAt:  twoDaysAgo,
			CreatorID:   users[0].ID,
		},
		{
			Title:       "タイムゾーン設定の不整合",
			Description: "ユーザーのタイムゾーン設定が一部の機能で正しく反映されず、時刻表示にズレが生じています。",
			Summary:     "タイムゾーン表示が不正確",
			Severity:    domain.SeverityLow,
			Status:      domain.StatusOpen,
			ImpactScope: "海外ユーザー約50名に影響",
			DetectedAt:  fiveDaysAgo,
			CreatorID:   users[3].ID,
			AssigneeID:  &users[2].ID,
		},
	}

	// Assign tags to incidents
	incidentTags := map[int][]int{
		0:  {1, 3},    // データベース接続エラー: サーバー障害, データベース
		1:  {2, 5},    // APIレスポンス遅延: アプリケーション, パフォーマンス
		2:  {4, 7},    // CSRF検証エラー: セキュリティ, ユーザー報告
		3:  {2, 8},    // グラフ表示問題: アプリケーション, バグ
		4:  {4},       // 不正アクセス: セキュリティ
		5:  {3, 6},    // 定期メンテナンス: データベース, メンテナンス
		6:  {2, 8},    // 画像アップロード: アプリケーション, バグ
		7:  {2, 9},    // メール遅延: アプリケーション, 設定変更
		8:  {0, 9},    // CDNキャッシュ: ネットワーク障害, 設定変更
		9:  {1, 5},    // メモリリーク: サーバー障害, パフォーマンス
		10: {2, 8},    // 決済二重実行: アプリケーション, バグ
		11: {2, 5},    // 検索精度低下: アプリケーション, パフォーマンス
		12: {2, 8},    // モバイルクラッシュ: アプリケーション, バグ
		13: {4, 6},    // SSL証明書: セキュリティ, メンテナンス
		14: {1, 3},    // バックアップ失敗: サーバー障害, データベース
		15: {1, 9},    // ログ過剰: サーバー障害, 設定変更
		16: {2, 9},    // レート制限: アプリケーション, 設定変更
		17: {3, 5},    // データ同期: データベース, パフォーマンス
		18: {0, 2},    // ダウンロード: ネットワーク障害, アプリケーション
		19: {2, 4},    // 突然ログアウト: アプリケーション, セキュリティ
		20: {2, 8},    // 管理画面エラー: アプリケーション, バグ
		21: {2, 7},    // プッシュ未配信: アプリケーション, ユーザー報告
		22: {2, 4},    // 外部API: アプリケーション, セキュリティ
		23: {0, 3},    // キャッシュ切断: ネットワーク障害, データベース
		24: {2, 8},    // CSV文字化け: アプリケーション, バグ
		25: {9},       // ジョブ時刻ズレ: 設定変更
		26: {2, 0},    // Webhook失敗: アプリケーション, ネットワーク障害
		27: {2, 4, 8}, // パスワードリセット: アプリケーション, セキュリティ, バグ
		28: {6, 9},    // 監視誤検知: メンテナンス, 設定変更
		29: {2, 8},    // タイムゾーン: アプリケーション, バグ
	}

	for i, incident := range incidents {
		if err := db.WithContext(ctx).Create(incident).Error; err != nil {
			return err
		}

		// Associate tags
		if tagIndices, ok := incidentTags[i]; ok {
			var incidentTagsList []domain.Tag
			for _, tagIdx := range tagIndices {
				incidentTagsList = append(incidentTagsList, *tags[tagIdx])
			}
			if err := db.Model(incident).Association("Tags").Append(incidentTagsList); err != nil {
				return err
			}
		}

		log.Printf("Created incident: %s (Severity: %s, Status: %s)", incident.Title, incident.Severity, incident.Status)
	}

	return nil
}
