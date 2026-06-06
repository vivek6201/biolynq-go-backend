package analytics

import (
	"time"

	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/models"
	"gorm.io/gorm"
)

type AnalyticsRepository struct {
	db *gorm.DB
}

type IAnalyticsRepository interface {
	GetTotalViews(profileID uuid.UUID) (int64, error)
	GetTotalClicks(profileID uuid.UUID) (int64, error)
	GetUniqueVisitors(profileID uuid.UUID) (int64, error)
	GetViewsOverTime(profileID uuid.UUID, since time.Time) ([]TimeSeriesItem, error)
	GetClicksOverTime(profileID uuid.UUID, since time.Time) ([]TimeSeriesItem, error)
	GetLinkStats(profileID uuid.UUID) ([]LinkStatsItem, error)
	GetDeviceBreakdown(profileID uuid.UUID) ([]GroupedStat, error)
	GetBrowserBreakdown(profileID uuid.UUID) ([]GroupedStat, error)
	GetOSBreakdown(profileID uuid.UUID) ([]GroupedStat, error)
	GetReferrerBreakdown(profileID uuid.UUID) ([]GroupedStat, error)
	GetCountryBreakdown(profileID uuid.UUID) ([]GroupedStat, error)
	GetCityBreakdown(profileID uuid.UUID) ([]GroupedStat, error)

	GetLinkByID(linkID uuid.UUID) (*models.Link, error)
	RecordEventTransaction(event *models.AnalyticEvent, metadata *models.VisitorMetadata) error
}

func NewAnalyticsRepository(db *gorm.DB) IAnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetTotalViews(profileID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.AnalyticEvent{}).
		Where("profile_id = ? AND event_type = ?", profileID, models.EventTypeProfileView).
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) GetTotalClicks(profileID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.AnalyticEvent{}).
		Where("profile_id = ? AND event_type = ?", profileID, models.EventTypeLinkClick).
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) GetUniqueVisitors(profileID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.AnalyticEvent{}).
		Where("profile_id = ?", profileID).
		Select("COUNT(DISTINCT visitor_metadata_id)").
		Scan(&count).Error
	return count, err
}

func (r *AnalyticsRepository) GetViewsOverTime(profileID uuid.UUID, since time.Time) ([]TimeSeriesItem, error) {
	var items []TimeSeriesItem
	err := r.db.Model(&models.AnalyticEvent{}).
		Select("TO_CHAR(clicked_at, 'YYYY-MM-DD') as date, count(*) as count").
		Where("profile_id = ? AND event_type = ? AND clicked_at >= ?", profileID, models.EventTypeProfileView, since).
		Group("TO_CHAR(clicked_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetClicksOverTime(profileID uuid.UUID, since time.Time) ([]TimeSeriesItem, error) {
	var items []TimeSeriesItem
	err := r.db.Model(&models.AnalyticEvent{}).
		Select("TO_CHAR(clicked_at, 'YYYY-MM-DD') as date, count(*) as count").
		Where("profile_id = ? AND event_type = ? AND clicked_at >= ?", profileID, models.EventTypeLinkClick, since).
		Group("TO_CHAR(clicked_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetLinkStats(profileID uuid.UUID) ([]LinkStatsItem, error) {
	var items []LinkStatsItem
	err := r.db.Model(&models.Link{}).
		Select("links.id, links.title, links.url, links.is_social, links.is_active, links.position, COUNT(analytic_events.id) as total_clicks").
		Joins("LEFT JOIN analytic_events ON analytic_events.link_id = links.id AND analytic_events.event_type = ?", models.EventTypeLinkClick).
		Where("links.profile_id = ?", profileID).
		Group("links.id, links.title, links.url, links.is_social, links.is_active, links.position").
		Order("links.position ASC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetDeviceBreakdown(profileID uuid.UUID) ([]GroupedStat, error) {
	var items []GroupedStat
	err := r.db.Model(&models.AnalyticEvent{}).
		Select("visitor_metadata.device as name, count(analytic_events.id) as count").
		Joins("JOIN visitor_metadata ON visitor_metadata.id = analytic_events.visitor_metadata_id").
		Where("analytic_events.profile_id = ?", profileID).
		Group("visitor_metadata.device").
		Order("count DESC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetBrowserBreakdown(profileID uuid.UUID) ([]GroupedStat, error) {
	var items []GroupedStat
	err := r.db.Model(&models.AnalyticEvent{}).
		Select("visitor_metadata.browser as name, count(analytic_events.id) as count").
		Joins("JOIN visitor_metadata ON visitor_metadata.id = analytic_events.visitor_metadata_id").
		Where("analytic_events.profile_id = ?", profileID).
		Group("visitor_metadata.browser").
		Order("count DESC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetOSBreakdown(profileID uuid.UUID) ([]GroupedStat, error) {
	var items []GroupedStat
	err := r.db.Model(&models.AnalyticEvent{}).
		Select("visitor_metadata.os as name, count(analytic_events.id) as count").
		Joins("JOIN visitor_metadata ON visitor_metadata.id = analytic_events.visitor_metadata_id").
		Where("analytic_events.profile_id = ?", profileID).
		Group("visitor_metadata.os").
		Order("count DESC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetReferrerBreakdown(profileID uuid.UUID) ([]GroupedStat, error) {
	var items []GroupedStat
	err := r.db.Model(&models.AnalyticEvent{}).
		Select("COALESCE(NULLIF(referrer, ''), 'Direct') as name, count(*) as count").
		Where("profile_id = ?", profileID).
		Group("COALESCE(NULLIF(referrer, ''), 'Direct')").
		Order("count DESC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetCountryBreakdown(profileID uuid.UUID) ([]GroupedStat, error) {
	var items []GroupedStat
	err := r.db.Model(&models.AnalyticEvent{}).
		Select("visitor_metadata.country as name, count(analytic_events.id) as count").
		Joins("JOIN visitor_metadata ON visitor_metadata.id = analytic_events.visitor_metadata_id").
		Where("analytic_events.profile_id = ?", profileID).
		Group("visitor_metadata.country").
		Order("count DESC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetCityBreakdown(profileID uuid.UUID) ([]GroupedStat, error) {
	var items []GroupedStat
	err := r.db.Model(&models.AnalyticEvent{}).
		Select("visitor_metadata.city as name, count(analytic_events.id) as count").
		Joins("JOIN visitor_metadata ON visitor_metadata.id = analytic_events.visitor_metadata_id").
		Where("analytic_events.profile_id = ?", profileID).
		Group("visitor_metadata.city").
		Order("count DESC").
		Scan(&items).Error
	return items, err
}

func (r *AnalyticsRepository) GetLinkByID(linkID uuid.UUID) (*models.Link, error) {
	var link models.Link
	err := r.db.First(&link, "id = ?", linkID).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *AnalyticsRepository) RecordEventTransaction(event *models.AnalyticEvent, metadata *models.VisitorMetadata) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existingMetadata models.VisitorMetadata
		err := tx.Where(models.VisitorMetadata{
			IP:      metadata.IP,
			Browser: metadata.Browser,
			OS:      metadata.OS,
			Device:  metadata.Device,
		}).Attrs(models.VisitorMetadata{
			ID:        uuid.New(),
			Country:   metadata.Country,
			City:      metadata.City,
			UserAgent: metadata.UserAgent,
		}).FirstOrCreate(&existingMetadata).Error

		if err != nil {
			return err
		}

		event.VisitorMetadataID = existingMetadata.ID
		return tx.Create(event).Error
	})
}
