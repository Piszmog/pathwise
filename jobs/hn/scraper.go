package hn

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Piszmog/hnclient"
	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/db/queries"
)

type Scraper struct {
	c        *hnclient.Client
	database db.Database
	logger   *slog.Logger
}

func NewScraper(logger *slog.Logger, database db.Database) *Scraper {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	client := hnclient.New(httpClient, hnclient.URLV0)
	return &Scraper{
		c:        client,
		database: database,
		logger:   logger,
	}
}

func (s *Scraper) Run(ctx context.Context) error {
	s.logger.DebugContext(ctx, "Running scraper")
	user, err := s.c.GetUser(ctx, "whoishiring")
	if err != nil {
		return err
	}

	s.logger.DebugContext(ctx, "Retrieved user data", "user", user)
	story, err := s.getStory(ctx, user.Submitted[0])
	if err != nil {
		return err
	}

	s.logger.DebugContext(ctx, "Retrieved story", "story", story)
	if !strings.HasPrefix(story.Title, "Ask HN: Who is hiring?") {
		return nil
	}

	exists, err := s.database.Queries().ExistsHNStory(ctx, story.ID)
	if err != nil {
		return err
	}

	if exists == 0 {
		s.logger.DebugContext(ctx, "Inserting story", "id", story.ID)
		err = s.database.Queries().InsertHNStory(ctx, queries.InsertHNStoryParams{
			PostedAt: story.Time.Time(),
			Title:    story.Title,
			ID:       story.ID,
		})
		if err != nil {
			return err
		}
	}

	for _, kidID := range story.Kids {
		commentExists, err := s.database.Queries().ExistsHNComment(ctx, kidID)
		if err != nil {
			return err
		}

		if commentExists == 1 {
			s.logger.DebugContext(ctx, "Skipping comment", "id", kidID)
			continue
		}

		comment, err := s.getComment(ctx, kidID)
		if err != nil {
			return err
		}

		s.logger.DebugContext(ctx, "Retrieved comment", "comment", comment)
		err = s.database.Queries().InsertHNComment(ctx, queries.InsertHNCommentParams{
			CommentedAt: comment.Time.Time(),
			Value:       comment.Text,
			ID:          comment.ID,
			HnStoryID:   story.ID,
		})
		if err != nil {
			return err
		}
	}
	s.logger.DebugContext(ctx, "Completed scraping")
	return nil
}

func (s *Scraper) getStory(ctx context.Context, id int64) (hnclient.Story, error) {
	var story hnclient.Story
	item, err := s.c.GetItem(ctx, id)
	if err != nil {
		return story, err
	}

	switch v := item.(type) {
	case hnclient.Story:
		story = v
	default:
		return story, errors.New("expected a story")
	}

	return story, nil
}

func (s *Scraper) getComment(ctx context.Context, id int64) (hnclient.Comment, error) {
	var comment hnclient.Comment
	kid, err := s.c.GetItem(ctx, id)
	if err != nil {
		return comment, err
	}

	switch v := kid.(type) {
	case hnclient.Comment:
		comment = v
	default:
		return comment, errors.New("expected a comment")
	}
	return comment, nil
}
