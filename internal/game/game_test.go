package game

import (
	"testing"
)

func TestHeadsUpFoldEndsHand(t *testing.T) {
	table, err := NewTable(
		[]string{"A", "B"},
		[]int{1000, 1000},
		[]bool{false, false},
		Config{SmallBlind: 5, BigBlind: 10, Seed: 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := table.StartHand(); err != nil {
		t.Fatal(err)
	}
	// First actor folds.
	if err := table.Apply(Action{Kind: ActionFold}); err != nil {
		t.Fatal(err)
	}
	if table.Street != StreetDone {
		t.Fatalf("want done, got %v", table.Street)
	}
	total := table.Players[0].Stack + table.Players[1].Stack
	if total != 2000 {
		t.Fatalf("chips not conserved: %d", total)
	}
}

func TestAutoCheckDownConservesChips(t *testing.T) {
	table, err := NewTable(
		[]string{"A", "B"},
		[]int{500, 500},
		[]bool{false, false},
		Config{SmallBlind: 5, BigBlind: 10, Seed: 99},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := table.StartHand(); err != nil {
		t.Fatal(err)
	}
	steps := 0
	for table.Street != StreetDone && steps < 200 {
		steps++
		if !table.NeedAction() {
			t.Fatalf("stuck at street=%v", table.Street)
		}
		legal := table.LegalActions()
		var act Action
		found := false
		for _, a := range legal {
			if a.Kind == ActionCheck || a.Kind == ActionCall {
				act = a
				found = true
				break
			}
		}
		if !found {
			act = legal[0]
		}
		if err := table.Apply(act); err != nil {
			t.Fatalf("apply %v: %v", act, err)
		}
	}
	if table.Street != StreetDone {
		t.Fatal("hand did not finish")
	}
	total := table.Players[0].Stack + table.Players[1].Stack
	if total != 1000 {
		t.Fatalf("chips not conserved: %d", total)
	}
}

func TestSeedReproducibleHoles(t *testing.T) {
	mk := func() []string {
		table, err := NewTable(
			[]string{"A", "B"},
			[]int{100, 100},
			[]bool{false, false},
			Config{SmallBlind: 1, BigBlind: 2, Seed: 12345},
		)
		if err != nil {
			t.Fatal(err)
		}
		if err := table.StartHand(); err != nil {
			t.Fatal(err)
		}
		return []string{table.Players[0].Hole[0].String(), table.Players[0].Hole[1].String()}
	}
	a, b := mk(), mk()
	if a[0] != b[0] || a[1] != b[1] {
		t.Fatalf("seed not reproducible: %v vs %v", a, b)
	}
}
