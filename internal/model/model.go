package model

import (
	"time"
)

// Item описывает статью как элемент ленты
type Item struct {
	Title      string
	Categories []string
	Link       string
	Date       time.Time
	Summary    string
	SourceName string
}

// Source описывает источник в формате базы данных
type Source struct {
	ID        int64
	Name      string
	FeedURL   string
	CreatedAt time.Time
}

// Article описывает статью в формате базы данных
type Article struct {
	ID          int64
	SourceID    int64
	Title       string
	Link        string
	Summary     string
	PublishedAt time.Time
	PostedAt    time.Time
	CreatedAt   time.Time
}
