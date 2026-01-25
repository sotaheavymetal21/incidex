package domain

import (
	"context"
	"time"
)

// PostMortem はインシデントのポストモーテム分析を表します
type PostMortem struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	IncidentID       uint       `gorm:"uniqueIndex;not null" json:"incident_id"` // 1対1の関係
	AuthorID         uint       `gorm:"not null;index" json:"author_id"`
	RootCause        string     `gorm:"type:text" json:"root_cause"`
	ImpactAnalysis   string     `gorm:"type:text" json:"impact_analysis"`
	WhatWentWell     string     `gorm:"type:text" json:"what_went_well"`
	WhatWentWrong    string     `gorm:"type:text" json:"what_went_wrong"`
	LessonsLearned   string     `gorm:"type:text" json:"lessons_learned"`
	FiveWhysAnalysis string     `gorm:"type:json" json:"five_whys_analysis"` // JSON形式で保存
	Status           PMStatus   `gorm:"size:20;not null;default:'draft';index" json:"status"`
	CreatedAt        time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	PublishedAt      *time.Time `json:"published_at"`

	// リレーション
	Incident    *Incident    `gorm:"foreignKey:IncidentID" json:"incident,omitempty"`
	Author      *User        `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	ActionItems []ActionItem `gorm:"foreignKey:PostMortemID" json:"action_items,omitempty"`
}

// PMStatus はポストモーテムのステータスを表します
type PMStatus string

const (
	PMStatusDraft     PMStatus = "draft"
	PMStatusPublished PMStatus = "published"
)

// FiveWhysAnalysis は「なぜなぜ分析」の構造を表します
type FiveWhysAnalysis struct {
	Why1 string `json:"why1" binding:"max=1000"`
	Why2 string `json:"why2" binding:"max=1000"`
	Why3 string `json:"why3" binding:"max=1000"`
	Why4 string `json:"why4" binding:"max=1000"`
	Why5 string `json:"why5" binding:"max=1000"`
}

// PostMortemRepository はポストモーテムデータアクセスのインターフェースを定義します
type PostMortemRepository interface {
	Create(ctx context.Context, pm *PostMortem) error
	FindByID(ctx context.Context, id uint) (*PostMortem, error)
	FindByIncidentID(ctx context.Context, incidentID uint) (*PostMortem, error)
	Update(ctx context.Context, pm *PostMortem) error
	Delete(ctx context.Context, id uint) error
	FindAll(ctx context.Context, filters PostMortemFilters, pagination Pagination) ([]*PostMortem, *PaginationResult, error)
}

// PostMortemFilters はポストモーテムのフィルタリングオプションを表します
type PostMortemFilters struct {
	Status   string
	AuthorID uint
	Search   string
	SortBy   string
	Order    string
}
