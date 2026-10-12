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
	"github.com/showx/PokerShow/internal/webui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "gui":
			runGUI(os.Args[2:])
			return
		case "cli":
			runCLI(os.Args[2:])
			return
		case "help", "-h", "--help":
			printUsage()
			return
		}
	}
	// Default: CLI for backward compatibility with existing flags.
	runCLI(os.Args[1:])
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `PokerShow — 本地德州调试沙盒

  pokershow gui [-addr 127.0.0.1:8788]
  pokershow cli [-seed 42] [-stack 1000] [-sb 5] [-bb 10] [-hands 1] [-auto]
  pokershow          # 等同 cli

`)
}

func runGUI(args []string) {
	fs := flag.NewFlagSet("gui", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8788", "listen address")
	_ = fs.Parse(args)
	srv := webui.NewServer(*addr)
	if err := srv.ListenAndServe(); err != nil {
		fail(err)
	}
}

func runCLI(args []string) {
	fs := flag.NewFlagSet("cli", flag.ExitOnError)
	seed := fs.Int64("seed", 42, "RNG seed for reproducible hands")
	stack := fs.Int("stack", 1000, "starting stack")
	sb := fs.Int("sb", 5, "small blind")
	bb := fs.Int("bb", 10, "big blind")
	hands := fs.Int("hands", 1, "number of hands to play")
	auto := fs.Bool("auto", false, "bot vs bot (no human input)")
	_ = fs.Parse(args)

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
	fmt.Println("图形界面: go run ./cmd/pokershow gui")

	for h := 0; h < *hands; h++ {
		if err := table.StartHand(); err != nil {
			fail(err)
		}
		printState(table)
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
			printState(table)
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

func printState(t *game.Table) {
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
		if p.Human && len(p.Hole) > 0 {
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
