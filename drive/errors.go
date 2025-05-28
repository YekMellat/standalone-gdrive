// Package drive implements a Google Drive client for standalone usage
//
// This file contains error definitions and handling utilities
package drive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/standalone-gdrive/fs"
	"google.golang.org/api/googleapi"
)

// Error definitions
var (
	InvalidCharacters      = errors.New("path contains invalid characters")
	DirectoryNotFound      = errors.New("directory not found")
	FileNotFound           = errors.New("file not found")
	AlreadyExists          = errors.New("item already exists")
	NotAFile               = errors.New("not a file")
	NotADirectory          = errors.New("not a directory")
	ItemIsDirectory        = errors.New("item is a directory")
	ErrorDriveFolderNotFound   = errors.New("drive folder not found")
	ErrorDriveFileNotFound     = errors.New("drive file not found")
	ErrorGoogleDriveTokenEmpty = errors.New("oauth token empty - please run authorize command")
	ErrorInvalidAuthCode       = errors.New("invalid auth code")
	ErrorAuthorizationFailed   = errors.New("authorization failed")
)

// shouldRetry returns a boolean as to whether this err deserves to be retried
func shouldRetry(ctx context.Context, err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	// Check for specific Google API errors
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		// Rate limiting errors
		if gerr.Code == 403 && strings.Contains(gerr.Message, "User Rate Limit Exceeded") {
			return true, err
		}
		if gerr.Code == 403 && strings.Contains(gerr.Message, "Rate Limit Exceeded") {
			return true, err
		}
		if gerr.Code == 403 && strings.Contains(gerr.Message, "rateLimitExceeded") {
			return true, err
		}
		
		// Quota exceeded errors
		if gerr.Code == 403 && strings.Contains(gerr.Message, "userRateLimitExceeded") {
			return false, fs.ErrorLimitExceeded
		}
		if gerr.Code == 403 && strings.Contains(gerr.Message, "Quota exceeded") {
			return false, fs.ErrorLimitExceeded
		}
		
		// Server error codes
		if gerr.Code == 500 || gerr.Code == 502 || gerr.Code == 503 || gerr.Code == 504 {
			return true, err
		}

		// Additional specific error handling
		if gerr.Code == 404 && strings.Contains(gerr.Message, "File not found") {
			return false, fs.ErrorObjectNotFound
		}
		if gerr.Code == 400 && strings.Contains(gerr.Message, "Bad Request") {
			return false, err
		}
		if gerr.Code == 401 {
			return false, err
		}
	}

	// Check context errors
	if errors.Is(err, context.Canceled) {
		return false, err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true, err
	}

	// Check for IO timeouts
	if errors.Is(err, io.EOF) {
		return true, err
	}
	
	// Generic HTTP errors
	if strings.Contains(err.Error(), "429 Too Many Requests") {
		return true, err
	}
	if strings.Contains(err.Error(), "http: can't write HTTP request on broken connection") {
		return true, err
	}
	if strings.Contains(err.Error(), "net/http: timeout awaiting response headers") {
		return true, err
	}
	if strings.Contains(err.Error(), "net/http: TLS handshake timeout") {
		return true, err
	}
	if strings.Contains(err.Error(), "connection reset by peer") {
		return true, err
	}
	
	// Default: don't retry
	return false, err
}

// parseRateLimit parses a rate limit header returning the duration to wait or 0 if not parsed
func parseRateLimit(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}

	// Try parsing as seconds first
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as a date format
	if date, err := http.ParseTime(retryAfter); err == nil {
		waitTime := date.Sub(time.Now())
		if waitTime > 0 {
			return waitTime
		}
	}

	// Failed to parse
	return 0
}

// translateError converts API errors into standard fs errors if possible
func translateError(err error, item string) error {
	if err == nil {
		return nil
	}
	
	// Check for Google API errors
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		switch gerr.Code {
		case 400:
			return fmt.Errorf("%s: bad request: %w", item, err)
		case 401:
			return fmt.Errorf("%s: unauthorized: %w", item, fs.ErrorPermissionDenied)
		case 403:
			// Check for specific error messages
			if strings.Contains(gerr.Message, "Rate Limit Exceeded") {
				return fmt.Errorf("%s: rate limit exceeded: %w", item, err)
			}
			if strings.Contains(gerr.Message, "User Rate Limit Exceeded") {
				return fmt.Errorf("%s: user rate limit exceeded: %w", item, err)
			}
			if strings.Contains(gerr.Message, "Quota exceeded") {
				return fmt.Errorf("%s: quota exceeded: %w", item, fs.ErrorLimitExceeded)
			}
			return fmt.Errorf("%s: access denied: %w", item, fs.ErrorPermissionDenied)
		case 404:
			return fs.ErrorObjectNotFound
		case 409:
			return fmt.Errorf("%s: conflict: %w", item, err)
		case 410:
			return fmt.Errorf("%s: item gone: %w", item, err)
		case 500:
			return fmt.Errorf("%s: internal server error: %w", item, err)
		}
	}

	// Keep the original error if no translation
	return err
}

// ProcessError extracts a more meaningful error from the googleapi error
func ProcessError(err error) error {
	if err == nil {
		return nil
	}
	
	// Check if it's a Google API specific error
	if apiErr, ok := err.(*googleapi.Error); ok {
		switch apiErr.Code {
		case 404:
			return ErrorDriveFileNotFound
		case 401:
			return ErrorAuthorizationFailed
		case 403:
			return fmt.Errorf("access denied: %v", apiErr.Message)
		default:
			return fmt.Errorf("API error: %v", apiErr.Message)
		}
	}
		// Check for other specific error types
	switch {
	case err.Error() == "oauth2: token expired and refresh token is not set":
		return ErrorGoogleDriveTokenEmpty
	case strings.Contains(err.Error(), "token has expired"):
		return fmt.Errorf("authorization token expired: %w", ErrorAuthorizationFailed)
	case strings.Contains(err.Error(), "invalid_grant"):
		return fmt.Errorf("invalid or expired refresh token: %w", ErrorAuthorizationFailed)
	}
	
	return err
}
