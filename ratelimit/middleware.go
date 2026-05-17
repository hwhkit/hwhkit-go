// Package ratelimit provides a Redis token-bucket rate limiter middleware.
package ratelimit

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/hwhkit/hwhkit-go/core/apierror"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Limit     int
	Window    time.Duration
	Burst     int
	KeyPrefix string
}

type KeyExtractor func(*http.Request) string

func ByIP() KeyExtractor {
	return func(r *http.Request) string { return r.RemoteAddr }
}

func ByHeader(name string) KeyExtractor {
	return func(r *http.Request) string {
		return r.Header.Get(name)
	}
}

func ByContextValue(key any) KeyExtractor {
	return func(r *http.Request) string {
		v, _ := r.Context().Value(key).(string)
		return v
	}
}

type Limiter struct {
	rdb    *redis.Client
	cfg    Config
	script *redis.Script
}

const tokenBucketLua = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local burst = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1]) or burst
local ts = tonumber(data[2]) or now

local elapsed = math.max(0, now - ts)
local refill = elapsed * (limit / window)
tokens = math.min(burst, tokens + refill)

if tokens < 1 then
  redis.call("HMSET", key, "tokens", tokens, "ts", now)
  redis.call("PEXPIRE", key, window * 2)
  local wait = math.ceil((1 - tokens) / (limit / window))
  return {0, wait}
end

tokens = tokens - 1
redis.call("HMSET", key, "tokens", tokens, "ts", now)
redis.call("PEXPIRE", key, window * 2)
return {1, 0}
`

func New(rdb *redis.Client, cfg Config) *Limiter {
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.Limit
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "ratelimit:"
	}
	return &Limiter{rdb: rdb, cfg: cfg, script: redis.NewScript(tokenBucketLua)}
}

func (l *Limiter) Allow(ctx context.Context, key string) (bool, time.Duration, error) {
	if key == "" {
		return true, 0, nil
	}
	now := time.Now().UnixMilli()
	res, err := l.script.Run(ctx, l.rdb,
		[]string{l.cfg.KeyPrefix + key},
		l.cfg.Limit, l.cfg.Window.Milliseconds(), l.cfg.Burst, now,
	).Slice()
	if err != nil {
		return false, 0, err
	}
	if len(res) < 2 {
		return false, 0, errors.New("malformed script reply")
	}
	allowed, _ := res[0].(int64)
	waitMs, _ := res[1].(int64)
	return allowed == 1, time.Duration(waitMs) * time.Millisecond, nil
}

func (l *Limiter) Middleware(extractor KeyExtractor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractor(r)
			ok, retry, err := l.Allow(r.Context(), key)
			if err != nil {
				apierror.Internal("rate limiter unavailable").WriteJSON(w)
				return
			}
			if !ok {
				if retry > 0 {
					w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())+1))
				}
				apierror.TooManyRequests("rate limit exceeded").WriteJSON(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
