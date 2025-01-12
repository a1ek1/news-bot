package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"news-bot/internal/model"
	"time"
)

type ArticlePostgresStorage struct {
	db *sqlx.DB
}

func NewArticleStorage(db *sqlx.DB) *ArticlePostgresStorage {
	return &ArticlePostgresStorage{db: db}
}

// Store нужен для сохранения статьи в базу данных
func (s *ArticlePostgresStorage) Store(ctx context.Context, article model.Article) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		"insert into articles (source_id, title, link, summary, published_at) "+
			"values ($1, $2, $3, $4, $5)"+
			"on conflict do nothing",
		article.SourceID,
		article.Title,
		article.Link,
		article.Summary,
		article.PublishedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *ArticlePostgresStorage) AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var articles []dbArticle
	if err := conn.SelectContext(
		ctx, &articles,
		"select * from articles where posted_at is null and published_at >= $1::timestamp order by published_at desc limit $2",
		since.UTC().Format(time.RFC3339),
		limit,
	); err != nil {
		return nil, err
	}

	return lo.Map(articles, func(article dbArticle, _ int) model.Article {
		return model.Article{
			ID:          article.ID,
			Title:       article.Title,
			Link:        article.Link,
			Summary:     article.Summary,
			PostedAt:    article.PostedAt,
			PublishedAt: article.PublishedAt,
			CreatedAt:   article.CreatedAt,
		}
	}), nil
}

func (s *ArticlePostgresStorage) MarkPosted(ctx context.Context, id int64) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		"update articles set posted_at = $1::timestamp where id = $2 ",
		time.Now().UTC().Format(time.RFC3339),
		id,
	); err != nil {
		return err
	}

	return nil
}

type dbArticle struct {
	ID          int64     `db:"id"`
	SourceID    string    `db:"source_id"`
	Title       string    `db:"title"`
	Link        string    `db:"link"`
	Summary     string    `db:"summary"`
	PostedAt    time.Time `db:"posted_at"`
	PublishedAt time.Time `db:"published_at"`
	CreatedAt   time.Time `db:"created_at"`
}
