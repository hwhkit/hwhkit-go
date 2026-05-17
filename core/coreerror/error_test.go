package coreerror

import (
	"errors"
	"testing"
)

func TestIntegrationErrorUnwrap(t *testing.T) {
	root := errors.New("dial: connection refused")
	e := Integration("postgres", KindConnectFailed, root)

	if !errors.Is(e, root) {
		t.Fatalf("errors.Is should walk wrapped root cause")
	}
	var ie *IntegrationError
	if !errors.As(e, &ie) {
		t.Fatalf("errors.As should find IntegrationError")
	}
	if ie.Key != "postgres" || ie.Kind != KindConnectFailed {
		t.Fatalf("unexpected fields: %+v", ie)
	}
}

func TestFeatureMismatch(t *testing.T) {
	e := FeatureMismatch("redis")
	if e.Kind != KindFeatureMismatch {
		t.Fatalf("expected KindFeatureMismatch, got %s", e.Kind)
	}
	if e.Key != "redis" {
		t.Fatalf("expected key=redis, got %s", e.Key)
	}
}
