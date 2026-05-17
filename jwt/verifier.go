// Package jwt provides JWKS and HMAC token verifiers with a chi-compatible Bearer middleware.
package jwt

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

var (
	ErrMissingToken    = errors.New("missing bearer token")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrExpired         = errors.New("token expired")
	ErrInvalidAudience = errors.New("invalid audience")
	ErrInvalidIssuer   = errors.New("invalid issuer")
)

type Claims struct {
	jwt.RegisteredClaims
	Raw map[string]any `json:"-"`
}

func (c Claims) Get(key string) (any, bool) {
	v, ok := c.Raw[key]
	return v, ok
}

func TypedClaims[T any](c Claims) (T, error) {
	var out T
	if len(c.Raw) == 0 {
		return out, errors.New("no claims body")
	}
	b, err := json.Marshal(c.Raw)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}

type Verifier interface {
	Verify(ctx context.Context, token string) (Claims, error)
}

type Options struct {
	Audience  string
	Issuer    string
	ClockSkew time.Duration
}

type hmacVerifier struct {
	secret []byte
	opts   Options
}

func NewHMACVerifier(secret []byte, opts Options) Verifier {
	return &hmacVerifier{secret: secret, opts: opts}
}

func (v *hmacVerifier) Verify(ctx context.Context, token string) (Claims, error) {
	return parseAndValidate(token, v.opts, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return v.secret, nil
	})
}

type jwksVerifier struct {
	cache *jwk.Cache
	url   string
	opts  Options
}

func NewJWKSVerifier(ctx context.Context, jwksURL string, opts Options) (Verifier, error) {
	cache := jwk.NewCache(ctx)
	if err := cache.Register(jwksURL, jwk.WithMinRefreshInterval(5*time.Minute)); err != nil {
		return nil, err
	}
	if _, err := cache.Refresh(ctx, jwksURL); err != nil {
		return nil, fmt.Errorf("initial JWKS refresh: %w", err)
	}
	return &jwksVerifier{cache: cache, url: jwksURL, opts: opts}, nil
}

func (v *jwksVerifier) Verify(ctx context.Context, token string) (Claims, error) {
	return parseAndValidate(token, v.opts, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		set, err := v.cache.Get(ctx, v.url)
		if err != nil {
			return nil, err
		}
		key, ok := set.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}
		var raw rsa.PublicKey
		if err := key.Raw(&raw); err != nil {
			return nil, err
		}
		return &raw, nil
	})
}

func parseAndValidate(tokenStr string, opts Options, keyFunc jwt.Keyfunc) (Claims, error) {
	tok, err := jwt.Parse(tokenStr, keyFunc, jwt.WithLeeway(opts.ClockSkew))
	if err != nil {
		return Claims{}, ErrInvalidSignature
	}
	if !tok.Valid {
		return Claims{}, ErrInvalidSignature
	}
	mc, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, errors.New("unexpected claim type")
	}

	exp, _ := mc["exp"].(float64)
	if exp > 0 && time.Unix(int64(exp), 0).Before(time.Now().Add(-opts.ClockSkew)) {
		return Claims{}, ErrExpired
	}
	if opts.Audience != "" {
		aud, _ := mc["aud"].(string)
		if aud != opts.Audience {
			return Claims{}, ErrInvalidAudience
		}
	}
	if opts.Issuer != "" {
		iss, _ := mc["iss"].(string)
		if iss != opts.Issuer {
			return Claims{}, ErrInvalidIssuer
		}
	}

	claims := Claims{Raw: map[string]any(mc)}
	if sub, _ := mc["sub"].(string); sub != "" {
		claims.Subject = sub
	}
	if iss, _ := mc["iss"].(string); iss != "" {
		claims.Issuer = iss
	}
	return claims, nil
}

