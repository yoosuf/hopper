package restaurants

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles restaurant operations
type Service struct {
	repo Repository
}

// New creates a new restaurants service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Restaurant represents a restaurant
type Restaurant struct {
	ID            uuid.UUID
	OwnerID       uuid.UUID
	Name          string
	Description   string
	CuisineType   string
	StreetAddress string
	City          string
	State         string
	PostalCode    string
	CountryCode   string
	Latitude      *float64
	Longitude     *float64
	Phone         string
	Email         string
	RegionID      uuid.UUID
	CurrencyCode  string
	Timezone      string
	IsActive      bool
	IsApproved    bool
	ApprovedAt    *string
}

// RestaurantHour represents restaurant operating hours
type RestaurantHour struct {
	ID           uuid.UUID
	RestaurantID uuid.UUID
	DayOfWeek    int
	OpenTime     string
	CloseTime    string
	IsClosed     bool
}

// CreateRestaurantRequest represents a create restaurant request
type CreateRestaurantRequest struct {
	Name          string
	Description   string
	CuisineType   string
	StreetAddress string
	City          string
	State         string
	PostalCode    string
	CountryCode   string
	Latitude      *float64
	Longitude     *float64
	Phone         string
	Email         string
	RegionID      uuid.UUID
	Timezone      string
}

// CreateRestaurant creates a new restaurant
func (s *Service) CreateRestaurant(ctx context.Context, ownerID uuid.UUID, req *CreateRestaurantRequest) (*Restaurant, error) {
	restaurant := &Restaurant{
		ID:            uuid.New(),
		OwnerID:       ownerID,
		Name:          req.Name,
		Description:   req.Description,
		CuisineType:   req.CuisineType,
		StreetAddress: req.StreetAddress,
		City:          req.City,
		State:         req.State,
		PostalCode:    req.PostalCode,
		CountryCode:   req.CountryCode,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		Phone:         req.Phone,
		Email:         req.Email,
		RegionID:      req.RegionID,
		CurrencyCode:  "USD", // Default, should be from region config
		Timezone:      req.Timezone,
		IsActive:      true,
		IsApproved:    false,
	}

	if err := s.repo.Create(ctx, restaurant); err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %w", err)
	}

	return restaurant, nil
}

// GetRestaurant retrieves a restaurant by ID
func (s *Service) GetRestaurant(ctx context.Context, id uuid.UUID) (*Restaurant, error) {
	return s.repo.GetByID(ctx, id)
}

// ListRestaurants lists restaurants for a region
func (s *Service) ListRestaurants(ctx context.Context, regionID uuid.UUID) ([]*Restaurant, error) {
	return s.repo.ListByRegion(ctx, regionID)
}

// ListMyRestaurants lists restaurants owned by a user
func (s *Service) ListMyRestaurants(ctx context.Context, ownerID uuid.UUID) ([]*Restaurant, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// UpdateRestaurant updates a restaurant
func (s *Service) UpdateRestaurant(ctx context.Context, id uuid.UUID, ownerID uuid.UUID, req *CreateRestaurantRequest) (*Restaurant, error) {
	restaurant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get restaurant: %w", err)
	}

	if restaurant.OwnerID != ownerID {
		return nil, fmt.Errorf("restaurant does not belong to owner")
	}

	restaurant.Name = req.Name
	restaurant.Description = req.Description
	restaurant.CuisineType = req.CuisineType
	restaurant.StreetAddress = req.StreetAddress
	restaurant.City = req.City
	restaurant.State = req.State
	restaurant.PostalCode = req.PostalCode
	restaurant.CountryCode = req.CountryCode
	restaurant.Latitude = req.Latitude
	restaurant.Longitude = req.Longitude
	restaurant.Phone = req.Phone
	restaurant.Email = req.Email
	restaurant.Timezone = req.Timezone

	if err := s.repo.Update(ctx, restaurant); err != nil {
		return nil, fmt.Errorf("failed to update restaurant: %w", err)
	}

	return restaurant, nil
}

// SetRestaurantHours sets restaurant operating hours
func (s *Service) SetRestaurantHours(ctx context.Context, restaurantID uuid.UUID, hours []*RestaurantHour) error {
	// Delete existing hours
	if err := s.repo.DeleteHours(ctx, restaurantID); err != nil {
		return fmt.Errorf("failed to delete existing hours: %w", err)
	}

	// Create new hours
	for _, hour := range hours {
		hour.RestaurantID = restaurantID
		if err := s.repo.CreateHour(ctx, hour); err != nil {
			return fmt.Errorf("failed to create hour: %w", err)
		}
	}

	return nil
}

// GetRestaurantHours gets restaurant operating hours
func (s *Service) GetRestaurantHours(ctx context.Context, restaurantID uuid.UUID) ([]*RestaurantHour, error) {
	return s.repo.ListHours(ctx, restaurantID)
}

// SearchRequest represents a restaurant search request with filters
type SearchRequest struct {
	RegionID      *uuid.UUID
	CuisineType   *string
	SearchQuery   *string
	MinRating     *float64
	MaxPrice      *int
	MinPrice      *int
	IsOpenNow     *bool
	Latitude      *float64
	Longitude     *float64
	MaxDistanceKm *float64
	SortBy        string
	SortOrder     string
	Limit         int
	Offset        int
}

// SearchResponse represents a restaurant search response
type SearchResponse struct {
	Restaurants []*Restaurant
	Total       int
}

// SearchRestaurants performs advanced search and filtering of restaurants
func (s *Service) SearchRestaurants(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// Set default values
	if req.SortBy == "" {
		req.SortBy = "name"
	}
	if req.SortOrder == "" {
		req.SortOrder = "asc"
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}

	restaurants, total, err := s.repo.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search restaurants: %w", err)
	}

	return &SearchResponse{
		Restaurants: restaurants,
		Total:       total,
	}, nil
}

// GetRestaurantRatingStats retrieves rating statistics for a restaurant
func (s *Service) GetRestaurantRatingStats(ctx context.Context, restaurantID uuid.UUID) (map[string]interface{}, error) {
	return s.repo.GetRatingStats(ctx, restaurantID)
}
