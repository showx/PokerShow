package hand

import (
	"testing"

	"github.com/showx/PokerShow/internal/card"
)

func must(t *testing.T, ss ...string) []card.Card {
	t.Helper()
	cs, err := card.ParseMany(ss...)
	if err != nil {
		t.Fatal(err)
	}
	return cs
}

func TestStraightFlushBeatsFour(t *testing.T) {
	sf, err := BestOf7(must(t, "Ah", "Kh", "Qh", "Jh", "Th", "2c", "3d"))
	if err != nil {
		t.Fatal(err)
	}
	quad, err := BestOf7(must(t, "9c", "9d", "9h", "9s", "8c", "2d", "3h"))
	if err != nil {
		t.Fatal(err)
	}
	if sf.Category != StraightFlush {
		t.Fatalf("want SF got %v", sf)
	}
	if sf.Cmp(quad) <= 0 {
		t.Fatalf("SF should beat quads: %v vs %v", sf, quad)
	}
}

func TestWheelStraight(t *testing.T) {
	r, err := BestOf7(must(t, "Ah", "2c", "3d", "4s", "5h", "9c", "Kd"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Category != Straight {
		t.Fatalf("want straight got %v", r)
	}
	if r.Kickers[0] != 5 {
		t.Fatalf("wheel high should be 5, got %v", r)
	}
}

func TestPairKickers(t *testing.T) {
	a, err := BestOf7(must(t, "Ah", "Ad", "Kc", "Qd", "2s", "3c", "4h"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := BestOf7(must(t, "Ah", "Ad", "Kc", "Jd", "2s", "3c", "4h"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Cmp(b) <= 0 {
		t.Fatalf("AQ kicker should beat AJ: %v vs %v", a, b)
	}
}

func TestFullHouse(t *testing.T) {
	r, err := BestOf7(must(t, "Ah", "Ad", "Ac", "Kd", "Ks", "2c", "3h"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Category != FullHouse {
		t.Fatalf("want full house got %v", r)
	}
}
