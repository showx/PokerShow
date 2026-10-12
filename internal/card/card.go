package card

import (
	"fmt"
	"strconv"
	"strings"
)

// Rank: 2..14 (A=14). Suit: c d h s.
type Card uint8

const (
	Clubs Card = iota
	Diamonds
	Hearts
	Spades
)

func New(rank int, suit Card) Card {
	if rank < 2 || rank > 14 || suit > Spades {
		panic("invalid card")
	}
	return Card(rank<<2 | int(suit))
}

func (c Card) Rank() int { return int(c >> 2) }
func (c Card) Suit() Card { return c & 3 }

func (c Card) String() string {
	r := c.Rank()
	var rs string
	switch r {
	case 14:
		rs = "A"
	case 13:
		rs = "K"
	case 12:
		rs = "Q"
	case 11:
		rs = "J"
	case 10:
		rs = "T"
	default:
		rs = strconv.Itoa(r)
	}
	suits := []string{"c", "d", "h", "s"}
	return rs + suits[c.Suit()]
}

// Parse accepts "Ah", "Td", "2c". Case-insensitive.
func Parse(s string) (Card, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0, fmt.Errorf("bad card %q", s)
	}
	rankPart := s[:len(s)-1]
	suitPart := strings.ToLower(s[len(s)-1:])
	var rank int
	switch strings.ToUpper(rankPart) {
	case "A":
		rank = 14
	case "K":
		rank = 13
	case "Q":
		rank = 12
	case "J":
		rank = 11
	case "T", "10":
		rank = 10
	default:
		n, err := strconv.Atoi(rankPart)
		if err != nil || n < 2 || n > 9 {
			return 0, fmt.Errorf("bad rank in %q", s)
		}
		rank = n
	}
	var suit Card
	switch suitPart {
	case "c":
		suit = Clubs
	case "d":
		suit = Diamonds
	case "h":
		suit = Hearts
	case "s":
		suit = Spades
	default:
		return 0, fmt.Errorf("bad suit in %q", s)
	}
	return New(rank, suit), nil
}

func ParseMany(ss ...string) ([]Card, error) {
	out := make([]Card, 0, len(ss))
	for _, s := range ss {
		c, err := Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func Format(cards []Card) string {
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = c.String()
	}
	return strings.Join(parts, " ")
}
