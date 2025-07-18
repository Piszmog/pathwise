package utils

import "strings"

// CleanJobTitle removes leading/trailing quotes and spaces from job titles
// Handles cases like:
// "Job title -> Job title
// "Job title" -> Job title
// " Job title -> Job title
func CleanJobTitle(title string) string {
	// Trim leading and trailing whitespace
	cleaned := strings.TrimSpace(title)

	// Remove leading quote if present
	cleaned = strings.TrimPrefix(cleaned, "\"")

	// Remove trailing quote if present
	cleaned = strings.TrimSuffix(cleaned, "\"")

	// Trim any remaining whitespace
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}
