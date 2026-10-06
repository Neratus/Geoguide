package mocks

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type MockStaticPageRepo struct {
	sync.RWMutex
	pages     map[domain.ImageID]domain.StaticPage
	slugIndex map[string]domain.ImageID
	files     map[domain.ImageID][]byte
}

func NewMockStaticPageRepo() *MockStaticPageRepo {
	return &MockStaticPageRepo{
		pages:     make(map[domain.ImageID]domain.StaticPage),
		slugIndex: make(map[string]domain.ImageID),
		files:     make(map[domain.ImageID][]byte),
	}
}

func (r *MockStaticPageRepo) Save(ctx context.Context, page *domain.StaticPage, file io.Reader) error {
	r.Lock()
	defer r.Unlock()
	if page.GetId() == uuid.Nil {
		newID := uuid.New()
		page.SetId(newID)
		r.pages[newID] = *page
		r.slugIndex[page.GetSlug()] = newID
		if file != nil {
			data, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			r.files[newID] = data
		}
		return nil
	}
	if _, ok := r.pages[page.GetId()]; !ok {
		return errors.New("static page not found")
	}
	oldPage := r.pages[page.GetId()]
	if oldPage.GetSlug() != page.GetSlug() {
		delete(r.slugIndex, oldPage.GetSlug())
		r.slugIndex[page.GetSlug()] = page.GetId()
	}
	r.pages[page.GetId()] = *page
	if file != nil {
		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		r.files[page.GetId()] = data
	}
	return nil
}

func (r *MockStaticPageRepo) FindByID(ctx context.Context, id domain.ImageID) (*domain.StaticPage, error) {
	r.RLock()
	defer r.RUnlock()
	page, ok := r.pages[id]
	if !ok {
		return nil, errors.New("static page not found")
	}
	cpy := page
	return &cpy, nil
}

func (r *MockStaticPageRepo) FindBySlug(ctx context.Context, slug string) (*domain.StaticPage, error) {
	r.RLock()
	defer r.RUnlock()
	id, ok := r.slugIndex[slug]
	if !ok {
		return nil, errors.New("static page not found")
	}
	page, ok := r.pages[id]
	if !ok {
		return nil, errors.New("static page not found")
	}
	cpy := page
	return &cpy, nil
}

func (r *MockStaticPageRepo) Update(ctx context.Context, page *domain.StaticPage, file io.Reader) error {
	return r.Save(ctx, page, file)
}

func (r *MockStaticPageRepo) Delete(ctx context.Context, id domain.ImageID) error {
	r.Lock()
	defer r.Unlock()
	page, ok := r.pages[id]
	if !ok {
		return errors.New("static page not found")
	}
	delete(r.pages, id)
	delete(r.slugIndex, page.GetSlug())
	delete(r.files, id)
	return nil
}

func (r *MockStaticPageRepo) GetFile(ctx context.Context, id domain.ImageID) (io.ReadCloser, error) {
	r.RLock()
	defer r.RUnlock()
	data, ok := r.files[id]
	if !ok {
		return nil, errors.New("file not found")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (r *MockStaticPageRepo) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	return nil
}

func (r *MockStaticPageRepo) DeleteFile(ctx context.Context, objectName string) error {
	return nil
}

func (r *MockStaticPageRepo) GetFileURL(ctx context.Context, objectName string) (string, error) {
	r.RLock()
	defer r.RUnlock()
	return objectName, nil
}
