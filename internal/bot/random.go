package bot

import (
	"math/rand"

	"github.com/showx/PokerShow/internal/game"
)

// Random picks among legal actions with simple weights; raises use min raise-to.
type Random struct {
	rng *rand.Rand
}

func NewRandom(seed int64) *Random {
	return &Random{rng: rand.New(rand.NewSource(seed))}
}

func (b *Random) Act(t *game.Table) game.Action {
	legal := t.LegalActions()
	if len(legal) == 0 {
		return game.Action{Kind: game.ActionCheck}
	}

	var pool []game.Action
	for _, a := range legal {
		switch a.Kind {
		case game.ActionFold:
			pool = append(pool, a)
		case game.ActionCheck:
			pool = append(pool, a, a, a)
		case game.ActionCall:
			pool = append(pool, a, a, a)
		case game.ActionRaise:
			pool = append(pool, a)
		case game.ActionAllIn:
			if b.rng.Intn(20) == 0 {
				pool = append(pool, a)
			}
		}
	}
	if len(pool) == 0 {
		pool = legal
	}
	pick := pool[b.rng.Intn(len(pool))]
	if pick.Kind == game.ActionRaise {
		p := t.Players[t.ToAct]
		cur := t.CurrentBet()
		minRaise := t.MinRaise()
		raiseTo := cur + minRaise
		if raiseTo-p.Bet > p.Stack {
			return game.Action{Kind: game.ActionAllIn}
		}
		if b.rng.Intn(2) == 0 && raiseTo+minRaise-p.Bet <= p.Stack {
			raiseTo += minRaise
		}
		pick.Amount = raiseTo
	}
	return pick
}
