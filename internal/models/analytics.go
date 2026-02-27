package models

type DailyCount struct {
	Date  string
	Count int
}

type AnalyticsAggregate struct {
	ShortCode   string
	TotalClicks int
	CLicksDay   int
	ClicksWeek  int
	ClickMonth  int
	Breakdown   []DailyCount
	Referrers   map[string]int
	Countries   map[string]int
	Devices     map[string]int
}
