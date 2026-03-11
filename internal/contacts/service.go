package contacts

import (
	"context"
	"errors"
)

type Repository interface {
	GetByUserID(ctx context.Context, userID string) (Contact, error)
	GetByExternalID(ctx context.Context, externalID string) (Contact, error)
	GetByID(ctx context.Context, id int) (Contact, error)
	Create(ctx context.Context, contact *Contact) error
	Update(ctx context.Context, id int, name, phone string) (Contact, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Resolve(ctx context.Context, userID, externalID *string) (Contact, error) {
	if userID != nil {
		return s.resolveByUserID(ctx, userID)
	}

	if externalID != nil {
		return s.resolveByExternalID(ctx, externalID)
	}

	return Contact{}, ErrUndefined
}

func (s *service) resolveByUserID(ctx context.Context, userID *string) (Contact, error) {
	contact, err := s.repo.GetByUserID(ctx, *userID)
	if err == nil {
		return contact, nil
	}

	if !errors.Is(err, ErrContactNotFound) {
		return Contact{}, err
	}

	newContact := Contact{UserID: userID}
	err = s.repo.Create(ctx, &newContact)
	return newContact, err
}

func (s *service) resolveByExternalID(ctx context.Context, externalID *string) (Contact, error) {
	contact, err := s.repo.GetByExternalID(ctx, *externalID)
	if err == nil {
		return contact, nil
	}

	if !errors.Is(err, ErrContactNotFound) {
		return Contact{}, err
	}

	newContact := Contact{ExternalID: externalID}
	err = s.repo.Create(ctx, &newContact)
	return newContact, err
}

func (s *service) Update(ctx context.Context, id int, name, phone string) (Contact, error) {
	if name == "" {
		return Contact{}, ErrInvalidName
	}

	if phone == "" { // TODO: phone check
		return Contact{}, ErrInvalidPhone
	}

	return s.repo.Update(ctx, id, name, phone)
}

func (s *service) InitContact(ctx context.Context, externalID, name, phone string) error {
	contact, err := s.Resolve(ctx, nil, &externalID)
	if err != nil {
		return err
	}

	_, err = s.repo.Update(ctx, contact.ID, name, phone)
	return err
}
