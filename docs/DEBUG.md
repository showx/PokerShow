# PokerShow 调试落地手册

目标：任何一手牌都能 **复现 → 单步看状态 → 定位是规则 bug 还是策略问题**。

## 三层怎么查

| 层 | 目录 | 先问什么 | 怎么验 |
| --- | --- | --- | --- |
| 牌型 | `internal/hand` | 这 7 张牌的胜负对不对？ | `go test ./internal/hand -v` |
| 规则 | `internal/game` | 谁该行动、底池、街是否推进？ | `go test ./internal/game -v` + 看 `table.Log` |
| 对局入口 | `cmd/pokershow` | 人机/机器人能否把一手打完？ | `go run ./cmd/pokershow gui` 或 `cli -seed N` |

不要一上来调 Bot。牌型错了，后面全是噪音。

## 可复现种子

同一 `seed` + 同一手序号 → 同一副牌序。

```bash
go run ./cmd/pokershow -seed 42 -hands 1
go run ./cmd/pokershow -seed 42 -hands 1   # 手牌应完全一致
go run ./cmd/pokershow -auto -seed 7 -hands 5
```

CLI 每手结束会打印完整 `Log`（盲注、行动、公共牌、摊牌）。把日志贴进 issue 即可复现。

## 最短调试闭环

1. **写失败用例**（已知牌面）放进 `internal/hand/rank_test.go`
2. `go test ./...` 红灯
3. 修评估器或引擎，直到绿灯
4. 用同一 `-seed` 开 CLI，肉眼看公共牌/行动是否合理
5. 筹码守恒：一局结束所有人 `stack` 之和应等于开局之和（测试已覆盖）

## 常见坑

- **加注额度**：`r` 是 *raise-to*（本街总下注），不是再加多少
- **Heads-up 盲注**：按钮是小盲，另一边大盲
- **全下后**：无人可行动时应自动发完公共牌到摊牌
- **只改 Bot**：先确认 `LegalActions` 在该状态下是否正确

## 下一阶段（有需要再加）

- `deck.ResetWithFixed`：指定翻牌/手牌做局面复盘
- JSON 手牌记录（HH）导出，方便 diff
- 赔率/胜率蒙特卡洛（仍挂在 `hand` + 固定 board 上测）
