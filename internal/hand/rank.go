package hand

import (
	"fmt"
	"sort"

	"github.com/showx/PokerShow/internal/card"
)

// Category from high to low.
type Category int

const (
	HighCard Category = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
)

func (c Category) String() string {
	names := []string{
		"高牌", "一对", "两对", "三条", "顺子", "同花", "葫芦", "四条", "同花顺",
	}
	if int(c) < 0 || int(c) >= len(names) {
		return "?"
	}
	return names[c]
}

// Rank is comparable: higher Category wins; then kickers lexicographically.
type Rank struct {
	Category Category
	Kickers  [5]int // high to low, padded with 0
}

func (r Rank) Cmp(o Rank) int {
	if r.Category != o.Category {
		if r.Category > o.Category {
			return 1
		}
		return -1
	}
	for i := 0; i < 5; i++ {
		if r.Kickers[i] != o.Kickers[i] {
			if r.Kickers[i] > o.Kickers[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

func (r Rank) String() string {
	return fmt.Sprintf("%s %v", r.Category, r.Kickers)
}

// BestOf7 evaluates 5–7 cards and returns the best 5-card rank.
func BestOf7(cards []card.Card) (Rank, error) {
	n := len(cards)
	if n < 5 || n > 7 {
		return Rank{}, fmt.Errorf("need 5-7 cards, got %d", n)
	}
	if n == 5 {
		return evaluate5(cards), nil
	}
	best := Rank{}
	first := true
	var choose func(start int, cur []card.Card)
	choose = func(start int, cur []card.Card) {
		if len(cur) == 5 {
			r := evaluate5(cur)
			if first || r.Cmp(best) > 0 {
				best = r
				first = false
			}
			return
		}
		need := 5 - len(cur)
		for i := start; i <= n-need; i++ {
			choose(i+1, append(cur, cards[i]))
		}
	}
	choose(0, nil)
	return best, nil
}

func evaluate5(cards []card.Card) Rank {
	ranks := make([]int, 5)
	suits := make([]card.Card, 5)
	for i, c := range cards {
		ranks[i] = c.Rank()
		suits[i] = c.Suit()
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ranks)))

	flush := suits[0] == suits[1] && suits[1] == suits[2] && suits[2] == suits[3] && suits[3] == suits[4]
	straight, high := isStraight(ranks)

	if flush && straight {
		return Rank{Category: StraightFlush, Kickers: [5]int{high}}
	}

	counts := map[int]int{}
	for _, r := range ranks {
		counts[r]++
	}
	type group struct {
		rank  int
		count int
	}
	var groups []group
	for r, c := range counts {
		groups = append(groups, group{r, c})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].count != groups[j].count {
			return groups[i].count > groups[j].count
		}
		return groups[i].rank > groups[j].rank
	})

	var kickers [5]int
	set := func(vals ...int) {
		for i, v := range vals {
			if i < 5 {
				kickers[i] = v
			}
		}
	}

	if groups[0].count == 4 {
		set(groups[0].rank, groups[1].rank)
		return Rank{Category: FourOfAKind, Kickers: kickers}
	}
	if groups[0].count == 3 && groups[1].count == 2 {
		set(groups[0].rank, groups[1].rank)
		return Rank{Category: FullHouse, Kickers: kickers}
	}
	if flush {
		set(ranks...)
		return Rank{Category: Flush, Kickers: kickers}
	}
	if straight {
		set(high)
		return Rank{Category: Straight, Kickers: kickers}
	}
	if groups[0].count == 3 {
		set(groups[0].rank, groups[1].rank, groups[2].rank)
		return Rank{Category: ThreeOfAKind, Kickers: kickers}
	}
	if groups[0].count == 2 && groups[1].count == 2 {
		hi, lo := groups[0].rank, groups[1].rank
		if lo > hi {
			hi, lo = lo, hi
		}
		set(hi, lo, groups[2].rank)
		return Rank{Category: TwoPair, Kickers: kickers}
	}
	if groups[0].count == 2 {
		set(groups[0].rank, groups[1].rank, groups[2].rank, groups[3].rank)
		return Rank{Category: OnePair, Kickers: kickers}
	}
	set(ranks...)
	return Rank{Category: HighCard, Kickers: kickers}
}

// isStraight returns whether ranks (desc, unique-ish) form a straight, and the high card (5 for wheel).
func isStraight(ranks []int) (bool, int) {
	uniq := make([]int, 0, 5)
	seen := map[int]bool{}
	for _, r := range ranks {
		if !seen[r] {
			seen[r] = true
			uniq = append(uniq, r)
		}
	}
	if len(uniq) != 5 {
		return false, 0
	}
	// A-5 wheel: A,5,4,3,2
	if uniq[0] == 14 && uniq[1] == 5 && uniq[2] == 4 && uniq[3] == 3 && uniq[4] == 2 {
		return true, 5
	}
	for i := 0; i < 4; i++ {
		if uniq[i]-uniq[i+1] != 1 {
			return false, 0
		}
	}
	return true, uniq[0]
}
