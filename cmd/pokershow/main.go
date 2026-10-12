package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/showx/PokerShow/internal/bot"
	"github.com/showx/PokerShow/internal/card"
	"github.com/showx/PokerShow/internal/game"
)

func main() {
	seed := flag.Int64("seed", 42, "RNG seed for reproducible hands")
	stack := flag.Int("stack", 1000, "starting stack")
	sb := flag.Int("sb", 5, "small blind")
	bb := flag.Int("bb", 10, "big blind")
	hands := flag.Int("hands", 1, "number of hands to play")
	auto := flag.Bool("auto", false, "bot vs bot (no human input)")
	flag.Parse()

	names := []string{"你", "Bot"}
	stacks := []int{*stack, *stack}
	humans := []bool{!*auto, false}
	if *auto {
		names = []string{"BotA", "BotB"}
	}

	table, err := game.NewTable(names, stacks, humans, game.Config{
		SmallBlind: *sb,
		BigBlind:   *bb,
		Seed:       *seed,
	})
	if err != nil {
		fail(err)
	}
	ai := bot.NewRandom(*seed + 7)
	in := bufio.NewReader(os.Stdin)

	fmt.Printf("PokerShow 调试局 seed=%d blinds=%d/%d stack=%d\n", *seed, *sb, *bb, *stack)
	fmt.Println("命令: f弃牌 c过牌/跟注 r加注 a全下 q退出  h帮助")

	for h := 0; h < *hands; h++ {
		if err := table.StartHand(); err != nil {
			fail(err)
		}
		printState(table, true)
		for !table.Snapshot(false).Finished {
			if !table.NeedAction() {
				break
			}
			p := table.Players[table.ToAct]
			var act game.Action
			if p.Human {
				act, err = readHuman(in, table)
				if err != nil {
					if err.Error() == "quit" {
						fmt.Println("退出")
						return
					}
					fmt.Println("输入无效:", err)
					continue
				}
			} else {
				act = ai.Act(table)
				fmt.Printf("→ %s: %s", p.Name, act.Kind)
				if act.Kind == game.ActionRaise {
					fmt.Printf(" 到 %d", act.Amount)
				}
				fmt.Println()
			}
			if err := table.Apply(act); err != nil {
				fmt.Println("动作失败:", err)
				continue
			}
			printState(table, true)
		}
		fmt.Println("--- 本手日志 ---")
		for _, line := range table.Log {
			fmt.Println(line)
		}
		fmt.Println("--- 筹码 ---")
		for _, p := range table.Players {
			fmt.Printf("%s: %d\n", p.Name, p.Stack)
		}
	}
}

func printState(t *game.Table, showHumanHole bool) {
	s := t.Snapshot(false)
	fmt.Printf("\n[手牌 #%d | %s | 底池 %d | 当前注 %d]\n", s.HandNo, s.Street, s.Pot, t.CurrentBet())
	if len(s.Board) > 0 {
		fmt.Println("公共牌:", card.Format(s.Board))
	}
	for i, p := range s.Players {
		mark := " "
		if i == s.ToAct && !s.Finished {
			mark = ">"
		}
		hole := ""
		if showHumanHole && p.Human && len(p.Hole) > 0 {
			hole = " 手牌=" + card.Format(p.Hole)
		}
		if s.Finished && len(p.Hole) > 0 {
			hole = " 手牌=" + card.Format(p.Hole)
		}
		flags := ""
		if p.Folded {
			flags += " 弃"
		}
		if p.AllIn {
			flags += " 全下"
		}
		fmt.Printf("%s %s stack=%d bet=%d%s%s\n", mark, p.Name, p.Stack, p.Bet, flags, hole)
	}
}

func readHuman(in *bufio.Reader, t *game.Table) (game.Action, error) {
	legal := t.LegalActions()
	fmt.Print("你的行动> ")
	line, err := in.ReadString('\n')
	if err != nil {
		return game.Action{}, err
	}
	line = strings.TrimSpace(strings.ToLower(line))
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return game.Action{}, fmt.Errorf("empty")
	}
	switch parts[0] {
	case "q", "quit":
		return game.Action{}, fmt.Errorf("quit")
	case "h", "help", "?":
		fmt.Println("f fold | c check/call | r [raise-to] | a all-in")
		fmt.Println("合法:", formatLegal(legal, t))
		return game.Action{}, fmt.Errorf("help")
	case "f", "fold":
		return game.Action{Kind: game.ActionFold}, nil
	case "c", "check", "call":
		for _, a := range legal {
			if a.Kind == game.ActionCheck || a.Kind == game.ActionCall {
				return a, nil
			}
		}
		return game.Action{}, fmt.Errorf("不能过牌/跟注")
	case "a", "allin", "all-in":
		return game.Action{Kind: game.ActionAllIn}, nil
	case "r", "raise":
		raiseTo := t.CurrentBet() + t.MinRaise()
		if len(parts) > 1 {
			n, err := strconv.Atoi(parts[1])
			if err != nil {
				return game.Action{}, err
			}
			raiseTo = n
		}
		return game.Action{Kind: game.ActionRaise, Amount: raiseTo}, nil
	default:
		return game.Action{}, fmt.Errorf("未知命令 %q", parts[0])
	}
}

func formatLegal(legal []game.Action, t *game.Table) string {
	var parts []string
	for _, a := range legal {
		if a.Kind == game.ActionRaise {
			parts = append(parts, fmt.Sprintf("raise>=%d", t.CurrentBet()+t.MinRaise()))
		} else {
			parts = append(parts, a.Kind.String())
		}
	}
	return strings.Join(parts, ", ")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
