package source

import (
	"context"
	"news-bot/internal/model"

	"github.com/SlyMarbo/rss"
	"github.com/samber/lo"
)

type RSSSource struct {
	URL        string
	SourceID   int64
	SourceName string
}

func NewRSSSource(s model.Source) RSSSource {
	return RSSSource{
		URL:        s.FeedURL,
		SourceID:   s.ID,
		SourceName: s.Name,
	}
}

func (s RSSSource) Fetch(ctx context.Context) ([]model.Item, error) {
	feed, err := s.loadFeed(ctx, s.URL)
	if err != nil {
		return nil, err
	}

	return lo.Map(feed.Items, func(item *rss.Item, _ int) model.Item {
		return model.Item{
			Title:      item.Title,
			Categories: item.Categories,
			Link:       item.Link,
			Date:       item.Date,
			Summary:    item.Summary,
			SourceName: s.SourceName,
		}
	}), nil
}

func (s RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {

	feedChan := make(chan *rss.Feed)
	errorChan := make(chan error)

	go func() {
		feed, err := rss.Fetch(url)
		if err != nil {
			errorChan <- err
			return
		}
		feedChan <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errorChan:
		return nil, err
	case feed := <-feedChan:
		return feed, nil
	}
}
