package hn

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Piszmog/hnclient"
	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/db/queries"
)

var (
	ErrExpectedStory   = errors.New("expected a story")
	ErrExpectedComment = errors.New("expected a comment")
)

type Scraper struct {
	c        *hnclient.Client
	database db.Database
	logger   *slog.Logger
}

func NewScraper(logger *slog.Logger, database db.Database, httpClient *http.Client) *Scraper {
	client := hnclient.New(httpClient, hnclient.URLV0)
	return &Scraper{
		c:        client,
		database: database,
		logger:   logger,
	}
}

func (s *Scraper) Run(ctx context.Context, ids chan<- int64) error {
	s.logger.DebugContext(ctx, "running scraper")
	user, err := s.c.GetUser(ctx, "whoishiring")
	if err != nil {
		return err
	}

	s.logger.DebugContext(ctx, "retrieved user data", "user", user)

	var story hnclient.Story
	for i := range 3 {
		userStory, storyErr := s.getStory(ctx, user.Submitted[i])
		if storyErr != nil {
			return storyErr
		}

		s.logger.DebugContext(ctx, "retrieved story", "story", userStory)
		if strings.HasPrefix(userStory.Title, "Ask HN: Who is hiring?") {
			story = userStory
			break
		}
	}

	if story.ID == 0 {
		return nil
	}

	exists, err := s.database.Queries().ExistsHNStory(ctx, story.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) || exists == 0 {
		s.logger.DebugContext(ctx, "inserting story", "id", story.ID)
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
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if commentExists == 1 {
			s.logger.DebugContext(ctx, "skipping comment", "id", kidID)
			continue
		}

		comment, err := s.getComment(ctx, kidID)
		if err != nil {
			return err
		}

		s.logger.DebugContext(ctx, "retrieved comment", "comment", comment)
		err = s.database.Queries().InsertHNComment(ctx, queries.InsertHNCommentParams{
			CommentedAt: comment.Time.Time(),
			Value:       comment.Text,
			ID:          comment.ID,
			HnStoryID:   story.ID,
		})
		if err != nil {
			return err
		}
		ids <- comment.ID
	}
	s.logger.DebugContext(ctx, "completed scraping")
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
		return story, ErrExpectedStory
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
		return comment, ErrExpectedComment
	}
	return comment, nil
}
