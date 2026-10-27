package health

import "time"

// Status represents the application health status response.
type Status struct {
	Status      string    `json:"status"`
	AppName     string    `json:"appName"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
}

// GetStatus returns the current health status payload.
func GetStatus(appName, environment string) Status {
	return Status{
		Status:      "healthy",
		AppName:     appName,
		Environment: environment,
		Timestamp:   time.Now().UTC(),
	}
}
