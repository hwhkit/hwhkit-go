// Package idempotency implements an Idempotency-Key middleware that replays prior responses.
package idempotency

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/hwhkit/hwhkit-go/core/apierror"
	"github.com/redis/go-redis/v9"
)

const HeaderKey = "Idempotency-Key"

type Config struct {
	TTL       time.Duration
	KeyPrefix string
	LockTTL   time.Duration
}

type Store struct {
	rdb *redis.Client
	cfg Config
}

func New(rdb *redis.Client, cfg Config) *Store {
	if cfg.TTL <= 0 {
		cfg.TTL = 10 * time.Minute
	}
	if cfg.LockTTL <= 0 {
		cfg.LockTTL = 30 * time.Second
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "idempotency:"
	}
	return &Store{rdb: rdb, cfg: cfg}
}

type cachedResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

func (s *Store) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get(HeaderKey)
			if key == "" || !mutating(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			cacheKey := s.cfg.KeyPrefix + r.Method + ":" + r.URL.Path + ":" + key
			lockKey := cacheKey + ":lock"

			if cached, ok, err := s.get(ctx, cacheKey); err == nil && ok {
				replay(w, cached)
				return
			}

			got, err := s.rdb.SetNX(ctx, lockKey, "1", s.cfg.LockTTL).Result()
			if err != nil {
				apierror.Internal("idempotency lock unavailable").WriteJSON(w)
				return
			}
			if !got {
				apierror.Conflict("idempotency: a request with this key is in progress").WriteJSON(w)
				return
			}
			defer s.rdb.Del(context.Background(), lockKey)

			rec := &recorder{ResponseWriter: w, buf: &bytes.Buffer{}, headers: make(map[string]string)}
			next.ServeHTTP(rec, r)

			if rec.status >= 200 && rec.status < 300 {
				_ = s.put(ctx, cacheKey, &cachedResponse{
					Status:  rec.status,
					Headers: rec.headers,
					Body:    rec.buf.Bytes(),
				})
			}
		})
	}
}

func (s *Store) get(ctx context.Context, key string) (*cachedResponse, bool, error) {
	b, err := s.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out cachedResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func (s *Store) put(ctx context.Context, key string, v *cachedResponse) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, b, s.cfg.TTL).Err()
}

func mutating(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

func replay(w http.ResponseWriter, c *cachedResponse) {
	for k, v := range c.Headers {
		w.Header().Set(k, v)
	}
	w.Header().Set("Idempotency-Replay", "true")
	w.WriteHeader(c.Status)
	_, _ = w.Write(c.Body)
}

type recorder struct {
	http.ResponseWriter
	buf     *bytes.Buffer
	status  int
	headers map[string]string
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	for k, vs := range r.ResponseWriter.Header() {
		if len(vs) > 0 {
			r.headers[k] = vs[0]
		}
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	r.buf.Write(b)
	return r.ResponseWriter.Write(b)
}
