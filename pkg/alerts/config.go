package alerts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
)

// LoadConfig loads and validates the subscriptions configuration file
func LoadConfig(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig performs validation on the loaded configuration
func validateConfig(config *Config) error {
	if config.Version == "" {
		return fmt.Errorf("version field is required")
	}

	// Check for duplicate subscriber IDs
	ids := make(map[string]bool)
	for _, sub := range config.Subscribers {
		if sub.ID == "" {
			return fmt.Errorf("subscriber ID cannot be empty")
		}
		if ids[sub.ID] {
			return fmt.Errorf("duplicate subscriber ID: %s", sub.ID)
		}
		ids[sub.ID] = true

		// Validate email format
		if err := validateEmail(sub.Email); err != nil {
			return fmt.Errorf("invalid email for subscriber %s: %w", sub.ID, err)
		}

		// Validate preferences
		if err := validatePreferences(&sub.Preferences); err != nil {
			return fmt.Errorf("invalid preferences for subscriber %s: %w", sub.ID, err)
		}
	}

	return nil
}

// validateEmail checks if an email address is valid
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	return nil
}

// validatePreferences validates subscriber preferences
func validatePreferences(prefs *Preferences) error {
	// Validate states (if specified)
	validStates := map[string]bool{"VA": true, "NC": true}
	for _, state := range prefs.States {
		if !validStates[state] {
			return fmt.Errorf("invalid state: %s (must be VA or NC)", state)
		}
	}

	// Validate listing types (if specified)
	validListingTypes := map[string]bool{
		"Listed":     true,
		"Limited":    true,
		"Allocation": true,
		"Barrel":     true,
		"Christmas":  true,
	}
	for _, lt := range prefs.ListingTypes {
		if !validListingTypes[lt] {
			return fmt.Errorf("invalid listing_type: %s", lt)
		}
	}

	// Validate min_quantity
	if prefs.MinQuantity < 0 {
		return fmt.Errorf("min_quantity cannot be negative")
	}

	return nil
}

// GetEnabledSubscribers returns only enabled subscribers from the config
func GetEnabledSubscribers(config *Config) []Subscriber {
	var enabled []Subscriber
	for _, sub := range config.Subscribers {
		if sub.Enabled {
			enabled = append(enabled, sub)
		}
	}
	return enabled
}
