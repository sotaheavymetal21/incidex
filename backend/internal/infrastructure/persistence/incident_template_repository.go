package persistence

import (
	"context"
	"incidex/internal/domain"

	"gorm.io/gorm"
)

type incidentTemplateRepository struct {
	db *gorm.DB
}

func NewIncidentTemplateRepository(db *gorm.DB) domain.IncidentTemplateRepository {
	return &incidentTemplateRepository{db: db}
}

func (r *incidentTemplateRepository) Create(ctx context.Context, template *domain.IncidentTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *incidentTemplateRepository) FindAll(ctx context.Context, userID uint) ([]*domain.IncidentTemplate, error) {
	var templates []*domain.IncidentTemplate

	// ユーザー自身のテンプレート + 公開テンプレートを取得
	err := r.db.WithContext(ctx).
		Preload("Creator").
		Preload("Tags").
		Where("creator_id = ? OR is_public = ?", userID, true).
		Order("usage_count DESC, created_at DESC").
		Find(&templates).Error

	if err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *incidentTemplateRepository) FindByID(ctx context.Context, id uint) (*domain.IncidentTemplate, error) {
	var template domain.IncidentTemplate

	err := r.db.WithContext(ctx).
		Preload("Creator").
		Preload("Tags").
		First(&template, id).Error

	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (r *incidentTemplateRepository) Update(ctx context.Context, template *domain.IncidentTemplate) error {
	// タグの関連付けを更新するために、既存の関連を削除してから再作成
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存のタグ関連を削除
		if err := tx.Model(template).Association("Tags").Clear(); err != nil {
			return err
		}

		// テンプレートを更新（タグも含む）
		return tx.Save(template).Error
	})
}

func (r *incidentTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.IncidentTemplate{}, id).Error
}

func (r *incidentTemplateRepository) IncrementUsageCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.IncidentTemplate{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}
