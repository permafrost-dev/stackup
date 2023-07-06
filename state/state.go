package state

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stackup-app/stackup/containers"
)

// Define the state struct
type AppState struct {
	StartedAt         time.Time // The time the app was started
	Version           string    // The current app version
	CronEngine        *cron.Cron
	RunningContainers *containers.PodmanContainers
}

// Initialize a new AppState
func NewAppState(cron *cron.Cron, version string) *AppState {
	// cronengine := cron.New(cron.WithParser(
	// 	cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	// ))

	result := &AppState{
		StartedAt:  time.Now(),
		Version:    version,
		CronEngine: cron,
	}

	return result
}

func (s *AppState) Init() {
	// s.CronEngine.AddFunc("* * * * *", s.RefreshRunningContainers)
}

func (s *AppState) RefreshRunningContainers() {
	fmt.Println("refreshing running containers...")
	s.SetRunningContainers()
}

// Add a user to the state
func (s *AppState) SetRunningContainers() {
	s.RunningContainers = containers.GetActivePodmanContainers()
}

// Update the app version
func (s *AppState) UpdateVersion(version string) {
	s.Version = version
}
