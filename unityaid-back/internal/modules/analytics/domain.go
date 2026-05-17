package analytics

type Filters struct {
	From string
	To   string
}

type Metric struct {
	Code  string  `json:"code"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type ChartPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type TopVolunteer struct {
	UserID     string  `json:"userId"`
	UserName   string  `json:"userName"`
	Email      string  `json:"email"`
	TotalHours float64 `json:"totalHours"`
	Points     int     `json:"points"`
	Level      int     `json:"level"`
}

type OverviewReport struct {
	Metrics       []Metric       `json:"metrics"`
	Applications  []ChartPoint   `json:"applications"`
	Attendance    []ChartPoint   `json:"attendance"`
	TopVolunteers []TopVolunteer `json:"topVolunteers"`
}

type VolunteersReport struct {
	Metrics       []Metric       `json:"metrics"`
	Status        []ChartPoint   `json:"status"`
	HoursByLevel  []ChartPoint   `json:"hoursByLevel"`
	TopVolunteers []TopVolunteer `json:"topVolunteers"`
}

type EventsReport struct {
	Metrics      []Metric     `json:"metrics"`
	ByStatus     []ChartPoint `json:"byStatus"`
	Applications []ChartPoint `json:"applications"`
	Attendance   []ChartPoint `json:"attendance"`
}

type TasksReport struct {
	Metrics    []Metric     `json:"metrics"`
	ByStatus   []ChartPoint `json:"byStatus"`
	ByPriority []ChartPoint `json:"byPriority"`
	Completed  []ChartPoint `json:"completed"`
}

type GamificationTransaction struct {
	ID              string `json:"id"`
	UserID          string `json:"userId"`
	UserName        string `json:"userName"`
	Email           string `json:"email"`
	AchievementName string `json:"achievementName"`
	SourceType      string `json:"sourceType"`
	SourceID        string `json:"sourceId"`
	Points          int    `json:"points"`
	Reason          string `json:"reason"`
	CreatedAt       string `json:"createdAt"`
}

type GamificationReport struct {
	Metrics      []Metric                  `json:"metrics"`
	ByReason     []ChartPoint              `json:"byReason"`
	PointsByDay  []ChartPoint              `json:"pointsByDay"`
	Transactions []GamificationTransaction `json:"transactions"`
}

type AuditEntry struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	UserName   string `json:"userName"`
	Email      string `json:"email"`
	Method     string `json:"method"`
	Action     string `json:"action"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	Path       string `json:"path"`
	StatusCode int    `json:"statusCode"`
	CreatedAt  string `json:"createdAt"`
}

type AuditReport struct {
	Metrics  []Metric     `json:"metrics"`
	ByAction []ChartPoint `json:"byAction"`
	ByEntity []ChartPoint `json:"byEntity"`
	Activity []ChartPoint `json:"activity"`
	Entries  []AuditEntry `json:"entries"`
}

type ReportRow struct {
	Label   string  `json:"label"`
	Group   string  `json:"group"`
	Value   float64 `json:"value"`
	Details string  `json:"details"`
}

type ManagementReport struct {
	Code    string      `json:"code"`
	Title   string      `json:"title"`
	Metrics []Metric    `json:"metrics"`
	Rows    []ReportRow `json:"rows"`
	Risks   []ReportRow `json:"risks"`
}
