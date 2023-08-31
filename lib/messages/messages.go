package messages

import (
	"fmt"

	"github.com/stackup-app/stackup/lib/types"
)

func TaskNotFound(name string) string {
	return fmt.Sprintf("Task %s not found.", name)
}

func NotExplicitlyAllowed(at types.AccessType, str string) string {
	return fmt.Sprintf("Access to %s '%s' has not been explicitly allowed.", at.String(), str)
}

func AccessBlocked(at types.AccessType, str string) string {
	return fmt.Sprintf("Access to %s '%s' has is blocked.", at.String(), str)
}

func HttpRequestFailed(urlStr string, code int) string {
	return fmt.Sprintf("HTTP request failed: error %d (url: %s)", code, urlStr)
}

func ExitDueToChecksumMismatch() string {
	return "Exiting due to checksum mismatch."
}

func RemoteIncludeStatus(status string, name string) string {
	return "remote include (" + status + "): " + name
}

func RemoteIncludeChecksumMismatch(name string) string {
	return "checksum verification failed: " + name
}

func RemoteIncludeCannotLoad(name string) string {
	return "unable to load remote include: " + name
}
