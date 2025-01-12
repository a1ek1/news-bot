package fetcher

import (
	"context"
	"go.tomakado.io/containers/set"
	"log"
	"news-bot/internal/model"
	"news-bot/internal/source"
	"strings"
	"sync"
	"time"
)

type ArticleStorage interface {
	Store(ctx context.Context, article model.Article) error
}

type SourceStorage interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	articles ArticleStorage
	sources  SourceStorage

	fetchInterval  time.Duration
	filterKeywords []string
}

func New(
	articles ArticleStorage,
	sources SourceStorage,
	fetchInterval time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		articles:       articles,
		sources:        sources,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, src := range sources {
		wg.Add(1)

		rssSource := source.NewRSSSourceFromModel(src)
		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("fetch articles failed: %v", err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("process items failed: %v", err)
				return
			}
		}(rssSource)
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []model.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldBeSkipped(item) {
			log.Printf("skipped item: %v", item)
			continue
		}

		if err := f.articles.Store(ctx, model.Article{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fetcher) itemShouldBeSkipped(item model.Item) bool {
	categoriesSet := set.New(item.Categories...)

	for _, keyword := range f.filterKeywords {
		titleContainsKeywords := strings.Contains(strings.ToLower(item.Title), keyword)

		if categoriesSet.Contains(keyword) || titleContainsKeywords {
			return true
		}
	}

	return false
}
