package webui

import "testing"

func TestSessionHumanThenBot(t *testing.T) {
	s := NewSession()
	st := s.State()
	if st.HandNo != 1 {
		t.Fatalf("hand %d", st.HandNo)
	}
	if !st.YourTurn && !st.Finished {
		// Bot may have acted first in some button setups; ensure state is coherent.
		if len(st.Players) != 2 {
			t.Fatal("want 2 players")
		}
	}
	if err := s.Reset(NewOpts{Seed: 1, Stack: 200, SmallBlind: 1, BigBlind: 2}); err != nil {
		t.Fatal(err)
	}
	st = s.State()
	for i := 0; i < 20 && !st.Finished; i++ {
		if !st.YourTurn {
			t.Fatalf("expected human turn or finished, got %+v", st)
		}
		var kind string
		for _, a := range st.Legal {
			if a.Kind == "check" || a.Kind == "call" {
				kind = a.Kind
				break
			}
		}
		if kind == "" {
			kind = "fold"
		}
		if err := s.Act(ActReq{Kind: kind}); err != nil {
			t.Fatal(err)
		}
		st = s.State()
	}
	if !st.Finished {
		t.Fatal("hand should finish")
	}
	if err := s.NextHand(); err != nil {
		t.Fatal(err)
	}
	if s.State().HandNo != 2 {
		t.Fatalf("want hand 2")
	}
}
