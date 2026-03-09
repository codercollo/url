package service

import (
	"context"
	"url/internal/models"
	"url/internal/repository"
)

//AnalyticsService handles analytics queries for short codes

// Interface to interact with click data
type AnalyticsService struct {
	clicks repository.ClickStore
}

// NewAnalyticsService creates a new AnalyticsService
func NewAnalyticsService(clicks repository.ClickStore) *AnalyticsService {
	return &AnalyticsService{
		clicks: clicks,
	}
}

// GetForCode fetches aggregated click stats for a specific short code
func (s *AnalyticsService) GetForCode(c context.Context, code string) (*models.AnalyticsAggregate, error) {
	return s.clicks.GetAnalytics(c, code)
}

// GetAll fetches analytics for multiple short codes
// If codes is nil, fetches all active short codes (admin dashboard)
func (s *AnalyticsService) GetAll(c context.Context, codes []string) ([]*models.AnalyticsAggregate, error) {
	if codes == nil {
		var err error
		//Fetch all codes from DB
		codes, err = s.clicks.GetAllShortCodes(c)
		if err != nil {
			return nil, err
		}
	}
	results := make([]*models.AnalyticsAggregate, 0, len(codes))

	for _, code := range codes {
		//Fetch analytics per code
		agg, err := s.clicks.GetAnalytics(c, code)
		if err != nil {
			return nil, err
		}
		if agg != nil {
			//Add valid aggregates
			results = append(results, agg)
		}
	}

	return results, nil

}

// GetAllShortCodes returns all active short codes from the store
func (s *AnalyticsService) GetAllShortCodes(ctx context.Context) ([]string, error) {
	return s.clicks.GetAllShortCodes(ctx)
}
