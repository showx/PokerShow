package card

import (
	"fmt"
	"math/rand"
)

type Deck struct {
	cards []Card
	i     int
	rng   *rand.Rand
}

func NewDeck(seed int64) *Deck {
	d := &Deck{
		cards: make([]Card, 0, 52),
		rng:   rand.New(rand.NewSource(seed)),
	}
	for r := 2; r <= 14; r++ {
		for s := Clubs; s <= Spades; s++ {
			d.cards = append(d.cards, New(r, s))
		}
	}
	d.Shuffle()
	return d
}

func (d *Deck) Shuffle() {
	d.rng.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
	d.i = 0
}

func (d *Deck) Remaining() int { return len(d.cards) - d.i }

func (d *Deck) Draw() (Card, error) {
	if d.i >= len(d.cards) {
		return 0, fmt.Errorf("deck empty")
	}
	c := d.cards[d.i]
	d.i++
	return c, nil
}

func (d *Deck) DrawN(n int) ([]Card, error) {
	out := make([]Card, 0, n)
	for i := 0; i < n; i++ {
		c, err := d.Draw()
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// ResetWithFixed puts known cards at the front (for replay/debug), then fills the rest.
func (d *Deck) ResetWithFixed(front []Card) error {
	seen := map[Card]bool{}
	for _, c := range front {
		if seen[c] {
			return fmt.Errorf("duplicate card %s", c)
		}
		seen[c] = true
	}
	rest := make([]Card, 0, 52-len(front))
	for r := 2; r <= 14; r++ {
		for s := Clubs; s <= Spades; s++ {
			c := New(r, s)
			if !seen[c] {
				rest = append(rest, c)
			}
		}
	}
	d.rng.Shuffle(len(rest), func(i, j int) {
		rest[i], rest[j] = rest[j], rest[i]
	})
	d.cards = append(append([]Card{}, front...), rest...)
	d.i = 0
	return nil
}
