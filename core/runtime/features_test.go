package runtime

import "testing"

func TestFeaturesRoundTrip(t *testing.T) {
	f := New().Enable("postgres").EnableIf("redis", false).EnableIf("nats", true)
	if !f.Contains("postgres") {
		t.Fatal("postgres should be enabled")
	}
	if f.Contains("redis") {
		t.Fatal("redis should NOT be enabled")
	}
	if !f.Contains("nats") {
		t.Fatal("nats should be enabled")
	}
}

func TestKnownFeatures(t *testing.T) {
	want := map[string]bool{
		"postgres": true, "redis": true, "mongodb": true,
		"nats": true, "qdrant": true, "neo4j": true, "s3": true, "oss": true,
	}
	got := KnownFeatures()
	for _, k := range got {
		if !want[k] {
			t.Errorf("unexpected feature: %s", k)
		}
		delete(want, k)
	}
	if len(want) > 0 {
		t.Errorf("missing features: %v", want)
	}
}
