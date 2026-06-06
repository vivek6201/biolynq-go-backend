package analytics

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mssola/user_agent"
	"github.com/vivek6201/biolynq/internal/models"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/worker"
)

type AnalyticsService struct {
	repo        IAnalyticsRepository
	userService users.IUserService
	worker      worker.TaskDistributor
}

type IAnalyticsService interface {
	GetOverview(userID uuid.UUID) (*OverviewResponse, error)
	GetLinkStats(userID uuid.UUID) ([]LinkStatsItem, error)
	GetDemographics(userID uuid.UUID) (*DemographicsResponse, error)

	// Direct public queries
	GetLinkByID(linkID uuid.UUID) (*models.Link, error)
	RecordEvent(eventType models.EventType, profileID uuid.UUID, linkID *uuid.UUID, ip, userAgent, referrer string, clickedAt time.Time) error
	TrackClickAsync(ctx context.Context, link *models.Link, ip, userAgent, referrer string)
}

func NewAnalyticsService(repo IAnalyticsRepository, userService users.IUserService, worker worker.TaskDistributor) IAnalyticsService {
	return &AnalyticsService{
		repo:        repo,
		userService: userService,
		worker:      worker,
	}
}

func (s *AnalyticsService) GetOverview(userID uuid.UUID) (*OverviewResponse, error) {
	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	views, err := s.repo.GetTotalViews(profile.ID)
	if err != nil {
		return nil, err
	}

	clicks, err := s.repo.GetTotalClicks(profile.ID)
	if err != nil {
		return nil, err
	}

	uniques, err := s.repo.GetUniqueVisitors(profile.ID)
	if err != nil {
		return nil, err
	}

	var ctr float64
	if views > 0 {
		ctr = (float64(clicks) / float64(views)) * 100
	}

	// Fetch daily time series for the last 30 days
	since := time.Now().AddDate(0, 0, -30)
	rawViewsOverTime, err := s.repo.GetViewsOverTime(profile.ID, since)
	if err != nil {
		return nil, err
	}

	rawClicksOverTime, err := s.repo.GetClicksOverTime(profile.ID, since)
	if err != nil {
		return nil, err
	}

	// Populate missing dates with zero counts
	viewsOverTime := fillMissingDates(rawViewsOverTime, 30)
	clicksOverTime := fillMissingDates(rawClicksOverTime, 30)

	return &OverviewResponse{
		Views:          views,
		Clicks:         clicks,
		CTR:            ctr,
		UniqueVisitors: uniques,
		ViewsOverTime:  viewsOverTime,
		ClicksOverTime: clicksOverTime,
	}, nil
}

func (s *AnalyticsService) GetLinkStats(userID uuid.UUID) ([]LinkStatsItem, error) {
	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	totalViews, err := s.repo.GetTotalViews(profile.ID)
	if err != nil {
		return nil, err
	}

	rawStats, err := s.repo.GetLinkStats(profile.ID)
	if err != nil {
		return nil, err
	}

	// Calculate individual Link CTRs dynamically
	for i := range rawStats {
		if totalViews > 0 {
			rawStats[i].CTR = (float64(rawStats[i].TotalClicks) / float64(totalViews)) * 100
		} else {
			rawStats[i].CTR = 0
		}
	}

	return rawStats, nil
}

func (s *AnalyticsService) GetDemographics(userID uuid.UUID) (*DemographicsResponse, error) {
	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	devices, err := s.repo.GetDeviceBreakdown(profile.ID)
	if err != nil {
		return nil, err
	}

	browsers, err := s.repo.GetBrowserBreakdown(profile.ID)
	if err != nil {
		return nil, err
	}

	os, err := s.repo.GetOSBreakdown(profile.ID)
	if err != nil {
		return nil, err
	}

	referrers, err := s.repo.GetReferrerBreakdown(profile.ID)
	if err != nil {
		return nil, err
	}

	countries, err := s.repo.GetCountryBreakdown(profile.ID)
	if err != nil {
		return nil, err
	}

	cities, err := s.repo.GetCityBreakdown(profile.ID)
	if err != nil {
		return nil, err
	}

	return &DemographicsResponse{
		Devices:   devices,
		Browsers:  browsers,
		OS:        os,
		Referrers: referrers,
		Countries: countries,
		Cities:    cities,
	}, nil
}

func (s *AnalyticsService) GetLinkByID(linkID uuid.UUID) (*models.Link, error) {
	return s.repo.GetLinkByID(linkID)
}

func (s *AnalyticsService) RecordEvent(eventType models.EventType, profileID uuid.UUID, linkID *uuid.UUID, ip, userAgentStr, referrer string, clickedAt time.Time) error {
	// 1. Parse User-Agent details
	ua := user_agent.New(userAgentStr)
	browserName, _ := ua.Browser()
	if browserName == "" {
		browserName = "Unknown"
	}
	osName := ua.OS()
	if osName == "" {
		osName = "Unknown"
	}

	deviceName := "Desktop"
	if ua.Bot() {
		deviceName = "Bot"
	} else if ua.Mobile() {
		deviceName = "Mobile"
	}

	// 2. Perform a simple GeoIP stub
	country := "Unknown"
	city := "Unknown"
	clientIP := strings.TrimSpace(ip)
	if clientIP == "127.0.0.1" || clientIP == "::1" || strings.ToLower(clientIP) == "localhost" {
		country = "Local"
		city = "Local"
	}

	// 3. Prepare entities
	metadata := &models.VisitorMetadata{
		IP:        clientIP,
		Browser:   browserName,
		OS:        osName,
		Device:    deviceName,
		Country:   country,
		City:      city,
		UserAgent: userAgentStr,
	}

	if clickedAt.IsZero() {
		clickedAt = time.Now()
	}

	event := &models.AnalyticEvent{
		ID:        uuid.New(),
		ProfileID: profileID,
		EventType: eventType,
		LinkID:    linkID,
		Referrer:  referrer,
		ClickedAt: clickedAt,
	}

	// 4. Save inside a single transaction
	return s.repo.RecordEventTransaction(event, metadata)
}

func (s *AnalyticsService) TrackClickAsync(ctx context.Context, link *models.Link, ip, userAgent, referrer string) {
	payload := &worker.RecordEventPayload{
		EventType: models.EventTypeLinkClick,
		ProfileID: link.ProfileID,
		LinkID:    &link.ID,
		IP:        ip,
		UserAgent: userAgent,
		Referrer:  referrer,
		ClickedAt: time.Now(),
	}

	_ = s.worker.DistributeTaskRecordEvent(ctx, payload)
}

// fillMissingDates populates any missing days in the last N days with zero values
func fillMissingDates(rawItems []TimeSeriesItem, days int) []TimeSeriesItem {
	itemMap := make(map[string]int64)
	for _, item := range rawItems {
		itemMap[item.Date] = item.Count
	}

	result := make([]TimeSeriesItem, days)
	now := time.Now()
	for i := 0; i < days; i++ {
		dateStr := now.AddDate(0, 0, -1*(days-1-i)).Format("2006-01-02")
		
		count := int64(0)
		if val, exists := itemMap[dateStr]; exists {
			count = val
		}
		result[i] = TimeSeriesItem{
			Date:  dateStr,
			Count: count,
		}
	}
	return result
}
