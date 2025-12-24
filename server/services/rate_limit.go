package services

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/constants"
	"openshortpath/server/models"
)

// GetClientIP extracts the client IP address from the request
// It checks X-Forwarded-For, X-Real-IP headers, and falls back to RemoteAddr
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (first IP in the chain)
	forwardedFor := c.GetHeader("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		// If RemoteAddr doesn't have a port, use it as-is
		return c.Request.RemoteAddr
	}
	return ip
}

// GetUserPlan retrieves the user's plan from the database
// Returns "hobbyist" as default if user not found or plan is empty
func GetUserPlan(db *gorm.DB, userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user ID is empty")
	}

	var user models.User
	result := db.Where("user_id = ?", userID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return constants.PlanHobbyist, nil // Default to hobbyist
		}
		return "", fmt.Errorf("failed to get user plan: %w", result.Error)
	}

	// Return plan if set, otherwise default to hobbyist
	if user.Plan == "" {
		return constants.PlanHobbyist, nil
	}
	return user.Plan, nil
}

// GetRateLimitForPlan returns the rate limit per hour based on the plan
// Returns 0 for unlimited plans
func GetRateLimitForPlan(plan string) int {
	switch plan {
	case constants.PlanVerifiedAccess, constants.PlanPro:
		return 0 // Unlimited
	case constants.PlanHobbyist, "":
		return 5 // 5 requests per hour
	default:
		// Unknown plan, default to hobbyist limit
		return 5
	}
}

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	Limit     int       // Maximum requests allowed per hour
	Remaining int       // Number of requests remaining in current window
	Reset     time.Time // Time when the rate limit resets
	Exceeded  bool      // Whether the limit has been exceeded
}

// CheckRateLimit checks if the rate limit has been exceeded and increments the counter
// Returns rate limit information including limit, remaining, and reset time
// limitPerHour of 0 means unlimited (always returns exceeded=false)
func CheckRateLimit(db *gorm.DB, identifier string, limitType string, limitPerHour int) (*RateLimitInfo, error) {
	// Unlimited plans don't need rate limiting
	if limitPerHour == 0 {
		return &RateLimitInfo{
			Limit:     0,
			Remaining: -1, // -1 indicates unlimited
			Reset:     time.Time{},
			Exceeded:  false,
		}, nil
	}

	// Get current hour window start (truncate to hour)
	now := time.Now().UTC()
	windowStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
	// Reset time is the start of the next hour
	resetTime := windowStart.Add(time.Hour)

	// Use a transaction to atomically check and increment
	var rateLimit models.RateLimit
	err := db.Transaction(func(tx *gorm.DB) error {
		// Try to find existing rate limit record for this window
		result := tx.Where("identifier = ? AND type = ? AND window_start = ?", identifier, limitType, windowStart).First(&rateLimit)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new rate limit record
			rateLimit = models.RateLimit{
				Identifier:   identifier,
				Type:         limitType,
				RequestCount: 1,
				WindowStart:  windowStart,
			}
			return tx.Create(&rateLimit).Error
		} else if result.Error != nil {
			return fmt.Errorf("failed to query rate limit: %w", result.Error)
		}

		// Increment request count
		rateLimit.RequestCount++
		return tx.Save(&rateLimit).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	// Calculate remaining requests
	remaining := limitPerHour - rateLimit.RequestCount
	if remaining < 0 {
		remaining = 0
	}

	// Check if limit exceeded
	exceeded := rateLimit.RequestCount > limitPerHour

	return &RateLimitInfo{
		Limit:     limitPerHour,
		Remaining: remaining,
		Reset:     resetTime,
		Exceeded:  exceeded,
	}, nil
}

// GetRateLimitInfo gets rate limit information without incrementing the counter
// This is useful for read-only operations like displaying current rate limit status
func GetRateLimitInfo(db *gorm.DB, identifier string, limitType string, limitPerHour int) (*RateLimitInfo, error) {
	// Unlimited plans don't need rate limiting
	if limitPerHour == 0 {
		return &RateLimitInfo{
			Limit:     0,
			Remaining: -1, // -1 indicates unlimited
			Reset:     time.Time{},
			Exceeded:  false,
		}, nil
	}

	// Get current hour window start (truncate to hour)
	now := time.Now().UTC()
	windowStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
	// Reset time is the start of the next hour
	resetTime := windowStart.Add(time.Hour)

	// Query existing rate limit record for this window
	var rateLimit models.RateLimit
	result := db.Where("identifier = ? AND type = ? AND window_start = ?", identifier, limitType, windowStart).First(&rateLimit)

	requestCount := 0
	if result.Error == nil {
		requestCount = rateLimit.RequestCount
	} else if result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query rate limit: %w", result.Error)
	}

	// Calculate remaining requests
	remaining := limitPerHour - requestCount
	if remaining < 0 {
		remaining = 0
	}

	// Check if limit exceeded
	exceeded := requestCount > limitPerHour

	return &RateLimitInfo{
		Limit:     limitPerHour,
		Remaining: remaining,
		Reset:     resetTime,
		Exceeded:  exceeded,
	}, nil
}

// CleanupOldRateLimits removes rate limit records older than the specified duration
// This should be called periodically to prevent database bloat
func CleanupOldRateLimits(db *gorm.DB, olderThan time.Duration) error {
	cutoff := time.Now().UTC().Add(-olderThan)
	result := db.Where("window_start < ?", cutoff).Delete(&models.RateLimit{})
	return result.Error
}

// GetMonthlyLinkLimitForPlan returns the monthly link limit based on the plan
// Returns 0 for unlimited plans (though currently no plan has unlimited monthly links)
func GetMonthlyLinkLimitForPlan(plan string) int {
	switch plan {
	case constants.PlanPro:
		return 10000 // 10,000 links per month
	case constants.PlanHobbyist, constants.PlanVerifiedAccess, "":
		return 1000 // 1,000 links per month
	default:
		// Unknown plan, default to hobbyist limit
		return 1000
	}
}

// MonthlyLinkLimitInfo contains monthly link limit information
type MonthlyLinkLimitInfo struct {
	Limit     int       // Maximum links allowed per month
	Remaining int       // Number of links remaining in current month
	Reset     time.Time // Time when the monthly limit resets (start of next month)
	Exceeded  bool      // Whether the limit has been exceeded
}

// CheckMonthlyLinkLimit checks if the monthly link limit has been exceeded and increments the counter
// Returns monthly link limit information including limit, remaining, and reset time
// limitPerMonth of 0 means unlimited (always returns exceeded=false)
func CheckMonthlyLinkLimit(db *gorm.DB, identifier string, limitType string, limitPerMonth int) (*MonthlyLinkLimitInfo, error) {
	// Unlimited plans don't need monthly limiting
	if limitPerMonth == 0 {
		return &MonthlyLinkLimitInfo{
			Limit:     0,
			Remaining: -1, // -1 indicates unlimited
			Reset:     time.Time{},
			Exceeded:  false,
		}, nil
	}

	// Get current month start (first day of current month at 00:00:00 UTC)
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Reset time is the start of the next month
	var resetTime time.Time
	if now.Month() == 12 {
		resetTime = time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		resetTime = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	}

	// Use a transaction to atomically check and increment
	var monthlyLimit models.MonthlyLinkLimit
	err := db.Transaction(func(tx *gorm.DB) error {
		// Try to find existing monthly limit record for this month
		result := tx.Where("identifier = ? AND type = ? AND month_start = ?", identifier, limitType, monthStart).First(&monthlyLimit)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new monthly limit record
			monthlyLimit = models.MonthlyLinkLimit{
				Identifier: identifier,
				Type:       limitType,
				LinkCount:  1,
				MonthStart: monthStart,
			}
			return tx.Create(&monthlyLimit).Error
		} else if result.Error != nil {
			return fmt.Errorf("failed to query monthly link limit: %w", result.Error)
		}

		// Increment link count
		monthlyLimit.LinkCount++
		return tx.Save(&monthlyLimit).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to check monthly link limit: %w", err)
	}

	// Calculate remaining links
	remaining := limitPerMonth - monthlyLimit.LinkCount
	if remaining < 0 {
		remaining = 0
	}

	// Check if limit exceeded
	exceeded := monthlyLimit.LinkCount > limitPerMonth

	return &MonthlyLinkLimitInfo{
		Limit:     limitPerMonth,
		Remaining: remaining,
		Reset:     resetTime,
		Exceeded:  exceeded,
	}, nil
}
