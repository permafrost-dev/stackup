package state

// Define the state struct
type AppState struct {
	Users   map[string]string // Map of usernames to user data
	Version string            // The current app version
}

// Initialize a new AppState
func NewAppState(version string) *AppState {
	return &AppState{
		Users:   make(map[string]string),
		Version: version,
	}
}

// Add a user to the state
func (s *AppState) AddUser(username, data string) {
	s.Users[username] = data
}

// Remove a user from the state
func (s *AppState) RemoveUser(username string) {
	delete(s.Users, username)
}

// Get user data from the state
func (s *AppState) GetUserData(username string) (string, bool) {
	data, exists := s.Users[username]
	return data, exists
}

// Update the app version
func (s *AppState) UpdateVersion(version string) {
	s.Version = version
}
