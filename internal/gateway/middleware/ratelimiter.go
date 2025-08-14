package middleware

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type CustomRateLimiter struct {
	ipLimiters    map[string]*rate.Limiter
	userLimiters  map[string]*rate.Limiter
	mutex         sync.RWMutex
	ipLimit       rate.Limit
	userLimit     rate.Limit
	cleanInterval time.Duration
}

func NewRateLimiter(ipLimit, userLimit rate.Limit, cleanInterval time.Duration) *CustomRateLimiter {
	rl := &CustomRateLimiter{
		ipLimiters:    make(map[string]*rate.Limiter),
		userLimiters:  make(map[string]*rate.Limiter),
		ipLimit:       ipLimit,
		userLimit:     userLimit,
		cleanInterval: time.Minute * 10,
	}

	return rl
}
