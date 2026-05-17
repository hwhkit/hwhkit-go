package appctx

import "testing"

type tokenA struct{ V int }
type tokenB struct{ V string }

func TestInsertGetTyped(t *testing.T) {
	c := New()
	Insert(c, &tokenA{V: 7})
	Insert(c, &tokenB{V: "x"})

	a, ok := Get[tokenA](c)
	if !ok || a.V != 7 {
		t.Fatalf("Get[tokenA]: got=%+v ok=%v", a, ok)
	}
	b, ok := Get[tokenB](c)
	if !ok || b.V != "x" {
		t.Fatalf("Get[tokenB]: got=%+v ok=%v", b, ok)
	}
	if _, ok := Get[struct{ Z int }](c); ok {
		t.Fatalf("expected miss for unrelated type")
	}
}

func TestInsertGetNamed(t *testing.T) {
	c := New()
	c.InsertNamed("k", 42)
	v, ok := c.GetNamed("k")
	if !ok || v.(int) != 42 {
		t.Fatalf("named lookup: got=%v ok=%v", v, ok)
	}
}

func TestMustGetPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	c := New()
	_ = MustGet[tokenA](c)
}
