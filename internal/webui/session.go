package webui

import (
	"fmt"
	"sync"

	"github.com/showx/PokerShow/internal/bot"
	"github.com/showx/PokerShow/internal/card"
	"github.com/showx/PokerShow/internal/game"
)

type NewOpts struct {
	Seed       int64 `json:"seed"`
	Stack      int   `json:"stack"`
	SmallBlind int   `json:"smallBlind"`
	BigBlind   int   `json:"bigBlind"`
}

type Session struct {
	mu    sync.Mutex
	opts  NewOpts
	table *game.Table
	ai    *bot.Random
	err   string
}

func NewSession() *Session {
	s := &Session{}
	_ = s.Reset(NewOpts{Seed: 42, Stack: 1000, SmallBlind: 5, BigBlind: 10})
	return s
}

func (s *Session) Reset(opts NewOpts) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if opts.Stack <= 0 {
		opts.Stack = 1000
	}
	if opts.SmallBlind <= 0 {
		opts.SmallBlind = 5
	}
	if opts.BigBlind < opts.SmallBlind {
		opts.BigBlind = opts.SmallBlind * 2
	}
	table, err := game.NewTable(
		[]string{"你", "Bot"},
		[]int{opts.Stack, opts.Stack},
		[]bool{true, false},
		game.Config{SmallBlind: opts.SmallBlind, BigBlind: opts.BigBlind, Seed: opts.Seed},
	)
	if err != nil {
		return err
	}
	if err := table.StartHand(); err != nil {
		return err
	}
	s.opts = opts
	s.table = table
	s.ai = bot.NewRandom(opts.Seed + 7)
	s.err = ""
	s.advanceBotsLocked()
	return nil
}

func (s *Session) NextHand() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.table == nil {
		return fmt.Errorf("no table")
	}
	if s.table.Street != game.StreetDone {
		return fmt.Errorf("current hand not finished")
	}
	if err := s.table.StartHand(); err != nil {
		return err
	}
	s.err = ""
	s.advanceBotsLocked()
	return nil
}

type ActReq struct {
	Kind   string `json:"kind"` // fold|check|call|raise|allin
	Amount int    `json:"amount"`
}

func (s *Session) Act(req ActReq) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.table == nil {
		return fmt.Errorf("no table")
	}
	if !s.table.NeedAction() {
		return fmt.Errorf("no action needed")
	}
	p := s.table.Players[s.table.ToAct]
	if !p.Human {
		return fmt.Errorf("not your turn")
	}
	act, err := parseAct(req, s.table)
	if err != nil {
		return err
	}
	if err := s.table.Apply(act); err != nil {
		s.err = err.Error()
		return err
	}
	s.err = ""
	s.advanceBotsLocked()
	return nil
}

func parseAct(req ActReq, t *game.Table) (game.Action, error) {
	switch req.Kind {
	case "fold":
		return game.Action{Kind: game.ActionFold}, nil
	case "check":
		return game.Action{Kind: game.ActionCheck}, nil
	case "call":
		return game.Action{Kind: game.ActionCall}, nil
	case "allin", "all-in":
		return game.Action{Kind: game.ActionAllIn}, nil
	case "raise":
		amt := req.Amount
		if amt <= 0 {
			amt = t.CurrentBet() + t.MinRaise()
		}
		return game.Action{Kind: game.ActionRaise, Amount: amt}, nil
	default:
		return game.Action{}, fmt.Errorf("unknown kind %q", req.Kind)
	}
}

func (s *Session) advanceBotsLocked() {
	for i := 0; i < 64; i++ {
		if s.table.Street == game.StreetDone || !s.table.NeedAction() {
			return
		}
		p := s.table.Players[s.table.ToAct]
		if p.Human {
			return
		}
		act := s.ai.Act(s.table)
		if err := s.table.Apply(act); err != nil {
			s.err = err.Error()
			return
		}
	}
}

type StateDTO struct {
	Seed       int64        `json:"seed"`
	StackStart int          `json:"stackStart"`
	SmallBlind int          `json:"smallBlind"`
	BigBlind   int          `json:"bigBlind"`
	HandNo     int          `json:"handNo"`
	Street     string       `json:"street"`
	Board      []string     `json:"board"`
	Pot        int          `json:"pot"`
	CurrentBet int          `json:"currentBet"`
	MinRaise   int          `json:"minRaise"`
	ToAct      int          `json:"toAct"`
	Players    []PlayerDTO  `json:"players"`
	Legal      []LegalDTO   `json:"legal"`
	Finished   bool         `json:"finished"`
	YourTurn   bool         `json:"yourTurn"`
	Log        []string     `json:"log"`
	Error      string       `json:"error,omitempty"`
}

type PlayerDTO struct {
	Name   string   `json:"name"`
	Stack  int      `json:"stack"`
	Bet    int      `json:"bet"`
	Folded bool     `json:"folded"`
	AllIn  bool     `json:"allIn"`
	Human  bool     `json:"human"`
	Hole   []string `json:"hole"`
	ToAct  bool     `json:"toAct"`
}

type LegalDTO struct {
	Kind        string `json:"kind"`
	Label       string `json:"label"`
	MinRaiseTo  int    `json:"minRaiseTo,omitempty"`
	CallAmount  int    `json:"callAmount,omitempty"`
}

func (s *Session) State() StateDTO {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.table == nil {
		return StateDTO{Error: "no table"}
	}
	snap := s.table.Snapshot(false)
	players := make([]PlayerDTO, len(snap.Players))
	for i, p := range snap.Players {
		players[i] = PlayerDTO{
			Name:   p.Name,
			Stack:  p.Stack,
			Bet:    p.Bet,
			Folded: p.Folded,
			AllIn:  p.AllIn,
			Human:  p.Human,
			Hole:   cardsToStr(p.Hole),
			ToAct:  !snap.Finished && i == snap.ToAct,
		}
	}
	yourTurn := false
	if !snap.Finished && s.table.NeedAction() {
		yourTurn = s.table.Players[s.table.ToAct].Human
	}

	human := s.table.Players[0]
	toCall := s.table.CurrentBet() - human.Bet
	if toCall < 0 {
		toCall = 0
	}

	legal := make([]LegalDTO, 0, len(snap.Legal))
	for _, a := range snap.Legal {
		dto := LegalDTO{Kind: actionKindKey(a.Kind), Label: a.Kind.String()}
		switch a.Kind {
		case game.ActionCall:
			dto.CallAmount = toCall
			dto.Label = fmt.Sprintf("跟注 %d", toCall)
		case game.ActionRaise:
			dto.MinRaiseTo = s.table.CurrentBet() + s.table.MinRaise()
			dto.Label = fmt.Sprintf("加注到 ≥%d", dto.MinRaiseTo)
		}
		legal = append(legal, dto)
	}

	return StateDTO{
		Seed:       s.opts.Seed,
		StackStart: s.opts.Stack,
		SmallBlind: s.opts.SmallBlind,
		BigBlind:   s.opts.BigBlind,
		HandNo:     snap.HandNo,
		Street:     snap.Street.String(),
		Board:      cardsToStr(snap.Board),
		Pot:        snap.Pot,
		CurrentBet: s.table.CurrentBet(),
		MinRaise:   s.table.MinRaise(),
		ToAct:      snap.ToAct,
		Players:    players,
		Legal:      legal,
		Finished:   snap.Finished,
		YourTurn:   yourTurn,
		Log:        append([]string{}, s.table.Log...),
		Error:      s.err,
	}
}

func cardsToStr(cs []card.Card) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.String()
	}
	return out
}

func actionKindKey(k game.ActionKind) string {
	switch k {
	case game.ActionFold:
		return "fold"
	case game.ActionCheck:
		return "check"
	case game.ActionCall:
		return "call"
	case game.ActionRaise:
		return "raise"
	case game.ActionAllIn:
		return "allin"
	default:
		return "unknown"
	}
}
