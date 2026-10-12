# PokerShow — 本地德州扑克调试沙盒

把「发牌 → 下注 → 摊牌 → 结算」整条链路摊开，方便改规则、改对手、对照一手牌。  
目标不是最强 Bot，而是 **能跑、能复现、能单步查**。

> 空仓库起步：先落地可测的规则引擎 + CLI，再往策略 / 赔率 / UI 长。

## 怎么调试落地

按层推进，不要跳级：

```text
① 牌型评估  internal/hand     ← 单元测试钉死胜负
        ↓
② 牌局引擎  internal/game     ← 盲注、行动、街、底池、筹码守恒
        ↓
③ 对手接口  internal/bot      ← 先 Random，再换策略
        ↓
④ CLI 入口  cmd/pokershow     ← -seed 复现整手牌
```

详细步骤见 [docs/DEBUG.md](docs/DEBUG.md)。

## 快速开始

```bash
go test ./...
go run ./cmd/pokershow -seed 42
```

人机命令：`f` 弃牌 · `c` 过牌/跟注 · `r [raise-to]` 加注 · `a` 全下 · `q` 退出

机器人互打（无交互）：

```bash
go run ./cmd/pokershow -auto -seed 7 -hands 3
```

## 目录

| 路径 | 职责 |
| --- | --- |
| `internal/card` | 牌面、牌堆、固定种子洗牌 |
| `internal/hand` | 5–7 张最佳牌型比较 |
| `internal/game` | 德州桌：街、行动、摊牌 |
| `internal/bot` | 可替换对手（当前 Random） |
| `cmd/pokershow` | 终端对局 / 复现入口 |
| `docs/DEBUG.md` | 复现、查层、常见坑 |

## 验收标准（当前骨架）

- [x] 同花顺 / 轮子顺 / 踢脚等牌型单测
- [x] 弃牌收池、过牌到摊牌、筹码守恒
- [x] 同一 `-seed` 开局手牌可复现
- [x] CLI 能打完一手并打印行动日志

## 刻意没做（避免一上来铺开）

多边边池精确拆分、限注/彩池注规则变体、网络牌桌、GUI、强化学习。  
这些都挂在现有 `Table` / `hand.Rank` 接口上往后加；先保证本机能把一手牌查清楚。
