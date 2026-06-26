package services

import (
	"fmt"
	"regexp"
	"strings"
)

// generateSlug creates a URL-friendly slug from a name
// Converts to lowercase and replaces spaces with hyphens
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

// validateEmailFormat validates email format using a simple regex
func validateEmailFormat(email string) error {
	// Simple email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// validatePhone validates phone format
// Accepts formats: +1234567890, 123-456-7890, (123) 456-7890, etc.
func validatePhone(phone string) error {
	// Remove common phone number formatting characters
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '+' {
			return r
		}
		return -1
	}, phone)

	// Check if it has a reasonable length (7-15 digits)
	if len(cleaned) < 7 || len(cleaned) > 15 {
		return fmt.Errorf("invalid phone format: must be 7-15 digits")
	}

	return nil
}
