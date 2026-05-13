package service

import (
	"net/http"
	"sync"
)

type ErrorType int

const (
	ErrNoError ErrorType = iota
	ErrQuotaExceeded
	ErrRateLimited
	ErrInvalidAuth
	ErrBadRequest
	ErrServerError
)

// ClassifyError maps HTTP status + body to an ErrorType.
func ClassifyError(statusCode int) ErrorType {
	switch {
	case statusCode == http.StatusTooManyRequests:
		return ErrRateLimited
	case statusCode == http.StatusPaymentRequired || statusCode == http.StatusRequestEntityTooLarge:
		return ErrQuotaExceeded
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return ErrInvalidAuth
	case statusCode == http.StatusBadRequest:
		return ErrBadRequest
	case statusCode >= 500:
		return ErrServerError
	default:
		return ErrNoError
	}
}

func ShouldRetry(errType ErrorType) bool {
	return errType == ErrRateLimited || errType == ErrServerError
}

func ShouldDisableChannel(errType ErrorType, consecutiveFailures int) bool {
	if errType == ErrServerError && consecutiveFailures >= 5 {
		return true
	}
	if errType == ErrInvalidAuth && consecutiveFailures >= 3 {
		return true
	}
	return false
}

// ChannelFailureTracker tracks consecutive failures per channel.
type ChannelFailureTracker struct {
	mu     sync.Mutex
	counts map[int]int // channelId -> consecutive failures
}

func NewChannelFailureTracker() *ChannelFailureTracker {
	return &ChannelFailureTracker{
		counts: make(map[int]int),
	}
}

func (t *ChannelFailureTracker) RecordFailure(channelId int, errType ErrorType) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if errType == ErrInvalidAuth || errType == ErrServerError {
		t.counts[channelId]++
	} else {
		// Reset on non-critical errors
		delete(t.counts, channelId)
		return false
	}

	failures := t.counts[channelId]
	return ShouldDisableChannel(errType, failures)
}

func (t *ChannelFailureTracker) RecordSuccess(channelId int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.counts, channelId)
}

// ChannelFailureTrack is the global tracker instance.
var ChannelFailureTrack = NewChannelFailureTracker()
