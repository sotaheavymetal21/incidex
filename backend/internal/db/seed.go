package db

import (
	"context"
	"incidex/internal/domain"
	"log"
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
			Email:    "admin@example.com",
			Name:     "管理者ユーザー",
			Role:     domain.RoleAdmin,
		},
		{
			Email:    "editor1@example.com",
			Name:     "編集者 太郎",
			Role:     domain.RoleEditor,
		},
		{
			Email:    "editor2@example.com",
			Name:     "編集者 花子",
			Role:     domain.RoleEditor,
		},
		{
			Email:    "viewer1@example.com",
			Name:     "閲覧者 一郎",
			Role:     domain.RoleViewer,
		},
		{
			Email:    "viewer2@example.com",
			Name:     "閲覧者 二郎",
			Role:     domain.RoleViewer,
		},
	}

	// Hash password "password123" for all test users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
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
		{Name: "ネットワーク障害", Color: "#ef4444"},     // red
		{Name: "サーバー障害", Color: "#f97316"},       // orange
		{Name: "アプリケーション", Color: "#3b82f6"},   // blue
		{Name: "データベース", Color: "#8b5cf6"},      // purple
		{Name: "セキュリティ", Color: "#dc2626"},      // dark red
		{Name: "パフォーマンス", Color: "#f59e0b"},    // amber
		{Name: "メンテナンス", Color: "#10b981"},      // green
		{Name: "ユーザー報告", Color: "#06b6d4"},      // cyan
		{Name: "バグ", Color: "#ec4899"},             // pink
		{Name: "設定変更", Color: "#6366f1"},         // indigo
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
	oneWeekAgo := now.Add(-7 * 24 * time.Hour)
	resolvedTime := now.Add(-2 * time.Hour)

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
	}

	// Assign tags to incidents
	incidentTags := map[int][]int{
		0: {1, 3}, // データベース接続エラー: サーバー障害, データベース
		1: {2, 5}, // APIレスポンス遅延: アプリケーション, パフォーマンス
		2: {4, 7}, // CSRF検証エラー: セキュリティ, ユーザー報告
		3: {2, 8}, // グラフ表示問題: アプリケーション, バグ
		4: {4},    // 不正アクセス: セキュリティ
		5: {3, 6}, // 定期メンテナンス: データベース, メンテナンス
		6: {2, 8}, // 画像アップロード: アプリケーション, バグ
		7: {2, 9}, // メール遅延: アプリケーション, 設定変更
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
