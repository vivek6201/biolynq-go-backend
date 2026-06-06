package analytics

import (
	"github.com/google/uuid"
)

type TimeSeriesItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type OverviewResponse struct {
	Views          int64            `json:"views"`
	Clicks         int64            `json:"clicks"`
	CTR            float64          `json:"ctr"`
	UniqueVisitors int64            `json:"unique_visitors"`
	ClicksOverTime []TimeSeriesItem `json:"clicks_over_time"`
	ViewsOverTime  []TimeSeriesItem `json:"views_over_time"`
}

type LinkStatsItem struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	IsSocial    bool      `json:"is_social"`
	IsActive    bool      `json:"is_active"`
	Position    int       `json:"position"`
	TotalClicks int64     `json:"total_clicks"`
	CTR         float64   `json:"ctr"`
}

type GroupedStat struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type DemographicsResponse struct {
	Devices   []GroupedStat `json:"devices"`
	Browsers  []GroupedStat `json:"browsers"`
	OS        []GroupedStat `json:"os"`
	Referrers []GroupedStat `json:"referrers"`
	Countries []GroupedStat `json:"countries"`
	Cities    []GroupedStat `json:"cities"`
}
