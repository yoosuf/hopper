package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles user operations
type Service struct {
	repo Repository
}

// New creates a new users service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Profile represents a user profile
type Profile struct {
	ID           uuid.UUID
	Email        string
	FirstName    string
	LastName     string
	Role         string
	IsSuspended  bool
	CreatedAt    string
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName string
	LastName  string
}

// GetProfile retrieves a user's profile
func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &Profile{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Role:        string(user.Role),
		IsSuspended: user.IsSuspended,
		CreatedAt:   user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// UpdateProfile updates a user's profile
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*Profile, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetProfile(ctx, userID)
}

// Address represents a user address
type Address struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	StreetAddress string
	City         string
	State        string
	PostalCode   string
	CountryCode  string
	Latitude     *float64
	Longitude    *float64
	IsDefault    bool
}

// CreateAddressRequest represents a create address request
type CreateAddressRequest struct {
	StreetAddress string
	City         string
	State        string
	PostalCode   string
	CountryCode  string
	Latitude     *float64
	Longitude    *float64
	IsDefault    bool
}

// CreateAddress creates a new user address
func (s *Service) CreateAddress(ctx context.Context, userID uuid.UUID, req *CreateAddressRequest) (*Address, error) {
	address := &Address{
		ID:           uuid.New(),
		UserID:       userID,
		StreetAddress: req.StreetAddress,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		CountryCode:  req.CountryCode,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		IsDefault:    req.IsDefault,
	}

	if err := s.repo.CreateAddress(ctx, address); err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	return address, nil
}

// ListAddresses lists all addresses for a user
func (s *Service) ListAddresses(ctx context.Context, userID uuid.UUID) ([]*Address, error) {
	return s.repo.ListAddresses(ctx, userID)
}

// UpdateAddress updates a user address
func (s *Service) UpdateAddress(ctx context.Context, addressID uuid.UUID, userID uuid.UUID, req *CreateAddressRequest) (*Address, error) {
	address, err := s.repo.GetAddress(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	if address.UserID != userID {
		return nil, fmt.Errorf("address does not belong to user")
	}

	address.StreetAddress = req.StreetAddress
	address.City = req.City
	address.State = req.State
	address.PostalCode = req.PostalCode
	address.CountryCode = req.CountryCode
	address.Latitude = req.Latitude
	address.Longitude = req.Longitude
	address.IsDefault = req.IsDefault

	if err := s.repo.UpdateAddress(ctx, address); err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	return address, nil
}

// DeleteAddress deletes a user address
func (s *Service) DeleteAddress(ctx context.Context, addressID uuid.UUID, userID uuid.UUID) error {
	address, err := s.repo.GetAddress(ctx, addressID)
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}

	if address.UserID != userID {
		return fmt.Errorf("address does not belong to user")
	}

	return s.repo.DeleteAddress(ctx, addressID)
}
