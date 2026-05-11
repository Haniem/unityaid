package audit

type Entry struct {
	UserID     string
	Method     string
	Action     string
	EntityType string
	EntityID   string
	Path       string
	StatusCode int
	IPAddress  string
	UserAgent  string
}
