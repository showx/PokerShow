package game

import (
	"fmt"

	"github.com/showx/PokerShow/internal/card"
	"github.com/showx/PokerShow/internal/hand"
)

type Street int

const (
	StreetPreflop Street = iota
	StreetFlop
	StreetTurn
	StreetRiver
	StreetShowdown
	StreetDone
)

func (s Street) String() string {
	names := []string{"翻前", "翻牌", "转牌", "河牌", "摊牌", "结束"}
	if int(s) < 0 || int(s) >= len(names) {
		return "?"
	}
	return names[s]
}

type ActionKind int

const (
	ActionFold ActionKind = iota
	ActionCheck
	ActionCall
	ActionRaise
	ActionAllIn
)

func (a ActionKind) String() string {
	return []string{"弃牌", "过牌", "跟注", "加注", "全下"}[a]
}

type Action struct {
	Kind   ActionKind
	Amount int // for Raise: total chips to put in this street (raise-to); ignored otherwise
}

type Player struct {
	Name   string
	Stack  int
	Hole   []card.Card
	Bet    int  // chips committed this street
	Total  int  // chips committed this hand
	Folded bool
	AllIn  bool
	Human  bool
}

type Config struct {
	SmallBlind int
	BigBlind   int
	Seed       int64
}

type Table struct {
	cfg     Config
	deck    *card.Deck
	Players []*Player
	Button  int
	Street  Street
	Board   []card.Card
	Pot     int
	ToAct   int
	// betting state
	currentBet int // highest bet this street
	minRaise   int
	lastAggressor int
	acted      []bool
	HandNo     int
	Log        []string
}

func NewTable(names []string, stacks []int, humans []bool, cfg Config) (*Table, error) {
	n := len(names)
	if n < 2 || n > 9 {
		return nil, fmt.Errorf("need 2-9 players")
	}
	if len(stacks) != n || len(humans) != n {
		return nil, fmt.Errorf("stacks/humans length mismatch")
	}
	if cfg.SmallBlind <= 0 || cfg.BigBlind < cfg.SmallBlind {
		return nil, fmt.Errorf("invalid blinds")
	}
	ps := make([]*Player, n)
	for i := range names {
		ps[i] = &Player{Name: names[i], Stack: stacks[i], Human: humans[i]}
	}
	return &Table{
		cfg:    cfg,
		deck:   card.NewDeck(cfg.Seed),
		Players: ps,
		Button: -1,
	}, nil
}

func (t *Table) CfgBigBlind() int { return t.cfg.BigBlind }
func (t *Table) CfgSmallBlind() int { return t.cfg.SmallBlind }
func (t *Table) CurrentBet() int { return t.currentBet }
func (t *Table) MinRaise() int { return t.minRaise }

func (t *Table) note(format string, args ...any) {
	t.Log = append(t.Log, fmt.Sprintf(format, args...))
}

func (t *Table) alive() []*Player {
	var out []*Player
	for _, p := range t.Players {
		if !p.Folded {
			out = append(out, p)
		}
	}
	return out
}

func (t *Table) canAct(p *Player) bool {
	return !p.Folded && !p.AllIn && p.Stack > 0
}

func (t *Table) StartHand() error {
	t.HandNo++
	t.Board = nil
	t.Pot = 0
	t.Street = StreetPreflop
	t.currentBet = 0
	t.minRaise = t.cfg.BigBlind
	t.Log = nil
	t.acted = make([]bool, len(t.Players))

	for _, p := range t.Players {
		p.Hole = nil
		p.Bet = 0
		p.Total = 0
		p.Folded = false
		p.AllIn = false
	}

	t.Button = (t.Button + 1) % len(t.Players)
	t.deck = card.NewDeck(t.cfg.Seed + int64(t.HandNo)*9973)

	for _, p := range t.Players {
		hole, err := t.deck.DrawN(2)
		if err != nil {
			return err
		}
		p.Hole = hole
	}

	sb := (t.Button + 1) % len(t.Players)
	bb := (t.Button + 2) % len(t.Players)
	if len(t.Players) == 2 {
		sb = t.Button
		bb = (t.Button + 1) % len(t.Players)
	}

	t.postBlind(sb, t.cfg.SmallBlind, "小盲")
	t.postBlind(bb, t.cfg.BigBlind, "大盲")
	t.currentBet = t.cfg.BigBlind
	t.lastAggressor = bb

	// UTG / heads-up: first to act is left of BB
	t.ToAct = (bb + 1) % len(t.Players)
	t.skipToNextActor()
	t.note("手牌 #%d 开始 button=%s", t.HandNo, t.Players[t.Button].Name)
	return nil
}

func (t *Table) postBlind(i, amount int, label string) {
	p := t.Players[i]
	if amount > p.Stack {
		amount = p.Stack
	}
	p.Stack -= amount
	p.Bet += amount
	p.Total += amount
	t.Pot += amount
	if p.Stack == 0 {
		p.AllIn = true
	}
	t.note("%s 下%s %d", p.Name, label, amount)
}

func (t *Table) skipToNextActor() {
	n := len(t.Players)
	for i := 0; i < n; i++ {
		if t.canAct(t.Players[t.ToAct]) {
			return
		}
		t.ToAct = (t.ToAct + 1) % n
	}
}

func (t *Table) NeedAction() bool {
	if t.Street == StreetShowdown || t.Street == StreetDone {
		return false
	}
	alive := t.alive()
	if len(alive) <= 1 {
		return false
	}
	actors := 0
	for _, p := range alive {
		if t.canAct(p) {
			actors++
		}
	}
	if actors == 0 {
		return false
	}
	// betting complete when all who can act have matched currentBet and acted since last aggression
	for i, p := range t.Players {
		if !t.canAct(p) {
			continue
		}
		if p.Bet < t.currentBet || !t.acted[i] {
			return true
		}
	}
	return false
}

func (t *Table) LegalActions() []Action {
	if !t.NeedAction() {
		return nil
	}
	p := t.Players[t.ToAct]
	toCall := t.currentBet - p.Bet
	var out []Action
	out = append(out, Action{Kind: ActionFold})
	if toCall == 0 {
		out = append(out, Action{Kind: ActionCheck})
	} else if toCall < p.Stack {
		out = append(out, Action{Kind: ActionCall})
	}
	if p.Stack > toCall {
		out = append(out, Action{Kind: ActionRaise})
	}
	if p.Stack > 0 {
		out = append(out, Action{Kind: ActionAllIn})
	}
	return out
}

func (t *Table) Apply(a Action) error {
	if !t.NeedAction() {
		return fmt.Errorf("no action needed")
	}
	p := t.Players[t.ToAct]
	toCall := t.currentBet - p.Bet

	switch a.Kind {
	case ActionFold:
		p.Folded = true
		t.note("%s 弃牌", p.Name)
	case ActionCheck:
		if toCall != 0 {
			return fmt.Errorf("cannot check, toCall=%d", toCall)
		}
		t.note("%s 过牌", p.Name)
	case ActionCall:
		if toCall <= 0 {
			return fmt.Errorf("nothing to call")
		}
		pay := toCall
		if pay > p.Stack {
			pay = p.Stack
		}
		p.Stack -= pay
		p.Bet += pay
		p.Total += pay
		t.Pot += pay
		if p.Stack == 0 {
			p.AllIn = true
		}
		t.note("%s 跟注 %d", p.Name, pay)
	case ActionRaise:
		// a.Amount is raise-to (total bet this street)
		raiseTo := a.Amount
		if raiseTo < t.currentBet+t.minRaise && raiseTo < p.Bet+p.Stack {
			return fmt.Errorf("raise too small: need >= %d", t.currentBet+t.minRaise)
		}
		need := raiseTo - p.Bet
		if need <= toCall {
			return fmt.Errorf("raise must exceed call")
		}
		if need > p.Stack {
			return fmt.Errorf("not enough chips")
		}
		prev := t.currentBet
		p.Stack -= need
		p.Bet += need
		p.Total += need
		t.Pot += need
		raisedBy := p.Bet - prev
		if raisedBy > t.minRaise {
			t.minRaise = raisedBy
		}
		t.currentBet = p.Bet
		t.lastAggressor = t.ToAct
		for i := range t.acted {
			t.acted[i] = false
		}
		if p.Stack == 0 {
			p.AllIn = true
		}
		t.note("%s 加注到 %d", p.Name, p.Bet)
	case ActionAllIn:
		need := p.Stack
		p.Stack = 0
		p.Bet += need
		p.Total += need
		t.Pot += need
		p.AllIn = true
		if p.Bet > t.currentBet {
			raisedBy := p.Bet - t.currentBet
			if raisedBy > t.minRaise {
				t.minRaise = raisedBy
			}
			t.currentBet = p.Bet
			t.lastAggressor = t.ToAct
			for i := range t.acted {
				t.acted[i] = false
			}
		}
		t.note("%s 全下 %d (本街下注 %d)", p.Name, need, p.Bet)
	default:
		return fmt.Errorf("unknown action")
	}

	t.acted[t.ToAct] = true
	t.ToAct = (t.ToAct + 1) % len(t.Players)
	t.skipToNextActor()

	if !t.NeedAction() {
		return t.advanceStreet()
	}
	return nil
}

func (t *Table) advanceStreet() error {
	// collect bets already in pot; reset street bets
	for _, p := range t.Players {
		p.Bet = 0
	}
	t.currentBet = 0
	t.minRaise = t.cfg.BigBlind
	t.acted = make([]bool, len(t.Players))

	alive := t.alive()
	if len(alive) == 1 {
		return t.awardFoldWin(alive[0])
	}

	switch t.Street {
	case StreetPreflop:
		t.Street = StreetFlop
		flop, err := t.deck.DrawN(3)
		if err != nil {
			return err
		}
		t.Board = flop
		t.note("翻牌 %s", card.Format(t.Board))
	case StreetFlop:
		t.Street = StreetTurn
		c, err := t.deck.Draw()
		if err != nil {
			return err
		}
		t.Board = append(t.Board, c)
		t.note("转牌 %s", card.Format(t.Board))
	case StreetTurn:
		t.Street = StreetRiver
		c, err := t.deck.Draw()
		if err != nil {
			return err
		}
		t.Board = append(t.Board, c)
		t.note("河牌 %s", card.Format(t.Board))
	case StreetRiver:
		t.Street = StreetShowdown
		return t.showdown()
	default:
		return fmt.Errorf("bad street %v", t.Street)
	}

	// first to act postflop: left of button
	t.ToAct = (t.Button + 1) % len(t.Players)
	t.skipToNextActor()
	if !t.NeedAction() {
		// everyone all-in → deal remaining
		return t.advanceStreet()
	}
	return nil
}

func (t *Table) awardFoldWin(winner *Player) error {
	winner.Stack += t.Pot
	t.note("%s 收池 %d（其余弃牌）", winner.Name, t.Pot)
	t.Pot = 0
	t.Street = StreetDone
	return nil
}

type ShowdownResult struct {
	Player *Player
	Rank   hand.Rank
	Won    int
}

func (t *Table) showdown() error {
	type contender struct {
		p    *Player
		rank hand.Rank
	}
	var cs []contender
	for _, p := range t.alive() {
		cards := append(append([]card.Card{}, p.Hole...), t.Board...)
		r, err := hand.BestOf7(cards)
		if err != nil {
			return err
		}
		t.note("%s 摊牌 %s → %s", p.Name, card.Format(p.Hole), r)
		cs = append(cs, contender{p, r})
	}
	best := cs[0].rank
	for _, c := range cs[1:] {
		if c.rank.Cmp(best) > 0 {
			best = c.rank
		}
	}
	var winners []*Player
	for _, c := range cs {
		if c.rank.Cmp(best) == 0 {
			winners = append(winners, c.p)
		}
	}
	share := t.Pot / len(winners)
	rem := t.Pot % len(winners)
	for i, w := range winners {
		got := share
		if i < rem {
			got++
		}
		w.Stack += got
		t.note("%s 赢 %d", w.Name, got)
	}
	t.Pot = 0
	t.Street = StreetDone
	return nil
}

// Snapshot for CLI / bots.
type Snapshot struct {
	HandNo   int
	Street   Street
	Board    []card.Card
	Pot      int
	ToAct    int
	Players  []PlayerView
	Legal    []Action
	Finished bool
}

type PlayerView struct {
	Name   string
	Stack  int
	Bet    int
	Folded bool
	AllIn  bool
	Hole   []card.Card // only filled for human / showdown
	Human  bool
}

func (t *Table) Snapshot(revealAll bool) Snapshot {
	views := make([]PlayerView, len(t.Players))
	for i, p := range t.Players {
		v := PlayerView{
			Name: p.Name, Stack: p.Stack, Bet: p.Bet,
			Folded: p.Folded, AllIn: p.AllIn, Human: p.Human,
		}
		if revealAll || p.Human || t.Street == StreetShowdown || t.Street == StreetDone {
			v.Hole = append([]card.Card{}, p.Hole...)
		}
		views[i] = v
	}
	return Snapshot{
		HandNo:   t.HandNo,
		Street:   t.Street,
		Board:    append([]card.Card{}, t.Board...),
		Pot:      t.Pot,
		ToAct:    t.ToAct,
		Players:  views,
		Legal:    t.LegalActions(),
		Finished: t.Street == StreetDone,
	}
}
