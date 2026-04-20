package contacts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type Repository interface {
	GetByUserID(ctx context.Context, userID string) (Contact, error)
	GetByExternalID(ctx context.Context, externalID string) (Contact, error)
	GetByID(ctx context.Context, id int) (Contact, error)
	Create(ctx context.Context, contact *Contact) error
	Update(ctx context.Context, id int, name, phone string) (Contact, error)
	GetByPhone(ctx context.Context, phone string) (Contact, error)
}

type service struct {
	repo Repository

	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) Resolve(ctx context.Context, userID, externalID *string) (Contact, error) {
	if userID != nil {
		contact, err := s.resolveByUserID(ctx, userID)
		if err != nil {
			return Contact{}, fmt.Errorf("resolve by userID: %w", err)
		}
		return contact, nil
	}

	if externalID != nil {
		contact, err := s.resolveByExternalID(ctx, externalID)
		if err != nil {
			return Contact{}, fmt.Errorf("resolve by externalID: %w", err)
		}
		return contact, nil
	}

	return Contact{}, fmt.Errorf("resolve contact: userID and externalID is nil")
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

	if !s.checkPhone(phone) {
		return Contact{}, ErrInvalidPhone
	}

	updatedContact, err := s.repo.Update(ctx, id, name, phone)
	if err != nil {
		return Contact{}, fmt.Errorf("update contact: %w", err)
	}
	s.logger.Info("contact updated", "id", updatedContact.ID)
	return updatedContact, nil
}

func (s *service) InitContact(ctx context.Context, externalID, name, phone string) (Contact, error) {
	c, err := s.repo.GetByPhone(ctx, phone)
	if err == nil {
		if c.ExternalID == nil {
			c.ExternalID = &externalID
		}
		return c, nil
	}

	if !errors.Is(err, ErrContactNotFound) {
		return Contact{}, err
	}

	contact, err := s.Resolve(ctx, nil, &externalID)
	if err != nil {
		return Contact{}, err
	}

	updatedContact, err := s.repo.Update(ctx, contact.ID, name, phone)
	if err != nil {
		return Contact{}, err
	}
	s.logger.Info("contact initialized", "id", updatedContact.ID, "name", updatedContact.Name, "phone", updatedContact.Phone)
	return updatedContact, err
}

func (s *service) checkPhone(phone string) bool {
	return tjPhoneRegex.MatchString(phone)
}
