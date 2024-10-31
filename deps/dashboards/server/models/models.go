package models

type Dashboard struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Project     string `json:"project"`
}

type DashboardList struct {
	Dashboards []Dashboard `json:"dashboards"`
	Next       int         `json:"next"`
}
