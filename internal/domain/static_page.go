package domain

import (
	"strings"
	"time"
)

type StaticPage struct {
	id              ImageID
	slug            string
	title           string
	content         string
	metaDescription string
	imageURL        string
	imageAlt        string
	updatedAt       time.Time
	publishedAt     *time.Time
	isPublished     bool
}

func (s *StaticPage) GetId() ImageID {
	return s.id
}

func (s *StaticPage) GetSlug() string {
	return s.slug
}

func (s *StaticPage) GetTitle() string {
	return s.title
}

func (s *StaticPage) GetContent() string {
	return s.content
}

func (s *StaticPage) GetMetaDescription() string {
	return s.metaDescription
}

func (s *StaticPage) GetImageURL() string {
	return s.imageURL
}

func (s *StaticPage) GetImageAlt() string {
	return s.imageAlt
}

func (s *StaticPage) GetUpdatedAt() time.Time {
	return s.updatedAt
}

func (s *StaticPage) GetPublishedAt() *time.Time {
	return s.publishedAt
}

func (s *StaticPage) IsPublished() bool {
	return s.isPublished
}

func (s *StaticPage) SetId(id ImageID) {
	s.id = id
}

func (s *StaticPage) SetSlug(slug string) {
	s.slug = slug
}

func (s *StaticPage) SetTitle(title string) {
	s.title = title
}

func (s *StaticPage) SetContent(content string) {
	s.content = content
}

func (s *StaticPage) SetMetaDescription(desc string) {
	s.metaDescription = desc
}

func (s *StaticPage) SetImageURL(url string) {
	s.imageURL = url
}

func (s *StaticPage) SetImageAlt(alt string) {
	s.imageAlt = alt
}

func (s *StaticPage) SetUpdatedAt(t time.Time) {
	s.updatedAt = t
}

func (s *StaticPage) SetPublishedAt(t *time.Time) {
	s.publishedAt = t
}

func (s *StaticPage) SetPublished(published bool) {
	s.isPublished = published
}

func NewStaticPage(
	id ImageID,
	slug, title, content, metaDescription, imageURL, imageAlt string,
) (*StaticPage, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrEmptyStaticPageSlug
	}
	if len(slug) > GetConfig().MaxStaticPageSlugLength {
		return nil, ErrStaticPageSlugTooLong
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyStaticPageTitle
	}
	if len(title) > GetConfig().MaxStaticPageTitleLength {
		return nil, ErrStaticPageTitleTooLong
	}

	if content == "" {
		return nil, ErrEmptyStaticPageContent
	}
	if len(content) > GetConfig().MaxStaticPageContentLength {
		return nil, ErrStaticPageContentTooLong
	}

	if len(metaDescription) > GetConfig().MaxStaticPageMetaDescriptionLength {
		return nil, ErrStaticPageMetaDescriptionTooLong
	}

	if len(imageURL) > GetConfig().MaxStaticPageImageURLLength {
		return nil, ErrStaticPageImageURLTooLong
	}

	if len(imageAlt) > GetConfig().MaxStaticPageImageAltLength {
		return nil, ErrStaticPageImageAltTooLong
	}

	return &StaticPage{
		id:              id,
		slug:            slug,
		title:           title,
		content:         content,
		metaDescription: metaDescription,
		imageURL:        imageURL,
		imageAlt:        imageAlt,
		updatedAt:       time.Now(),
		publishedAt:     nil,
		isPublished:     false,
	}, nil
}

func NewStaticPageFromDB(
	id ImageID,
	slug, title, content, metaDescription, imageURL, imageAlt string,
	updatedAt time.Time,
	publishedAt *time.Time,
	isPublished bool,
) *StaticPage {
	return &StaticPage{
		id:              id,
		slug:            slug,
		title:           title,
		content:         content,
		metaDescription: metaDescription,
		imageURL:        imageURL,
		imageAlt:        imageAlt,
		updatedAt:       updatedAt,
		publishedAt:     publishedAt,
		isPublished:     isPublished,
	}
}
