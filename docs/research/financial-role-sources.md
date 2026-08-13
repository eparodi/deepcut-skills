# Financial Role — Research Sources

Gathered 2026-08-13 to seed the `financial-analyst` skill (template
version lives in `skills-test`; the reference implementation is
`deepcut-binance-bot`). Quality
bar: **academic, institutional, or regulator-grade only.** No signal
sellers, no "99% win rate" gurus, no Telegram channels, no sites that sell
you a strategy. Anything below marked `[Unverified]` could not be fetched
by the agent (bot protection) — verify before citing.

## Verified academic papers (the canon)

| Paper | Where | Why it matters here |
|---|---|---|
| Kelly (1956), "A New Interpretation of Information Rate", *Bell System Technical Journal* 35(4):917–926 | Classic, public | Mathematical foundation of optimal position sizing (growth-optimal betting). We use **fractional** Kelly thinking, not full Kelly — full Kelly is known to produce ruin-level drawdowns. |
| Markowitz (1952), "Portfolio Selection", *Journal of Finance* 7(1):77–91 | Classic, public | Diversification/portfolio theory — underpins the exposure cap (30%) and why independent per-symbol positions are not the same as account-level safety. |
| Liu & Tsyvinski (2021), "Risks and Returns of Cryptocurrency", *Review of Financial Studies* 34(6):2689–2727. NBER w24877 **verified 2026-08-13** via nber.org | [NBER](https://www.nber.org/papers/w24877) | Peer-reviewed evidence on crypto return factors: time-series momentum + investor attention. Directly relevant to what the personas should (and shouldn't) look at; a sanity check against TA folklore. |
| Makarov & Schoar (2020), "Trading and Arbitrage in Cryptocurrency Markets", *Journal of Financial Economics* 135(2):293–319 `[Unverified link]` | SSRN/JFE | The serious academic reference on crypto market microstructure, exchange fragmentation, and arbitrage — good grounding for fee/slippage realism. |
| López de Prado (2018), "The 10 Reasons Most Machine Learning Funds Fail", *Journal of Portfolio Management* 44(4):72–81 `[Unverified link]` | SSRN/JPM | The canonical warning list for ML-in-trading: backtest overfitting, non-stationarity, insufficient sample size. **This is the anti-scam lens for our own bot** — every new indicator/rule must survive these critiques. |
| Thorp (2006), "The Kelly Criterion in Blackjack, Sports Betting, and the Stock Market", *Handbook of Asset and Liability Management* `[Unverified link]` | Elsevier | How an actual quant legend applies Kelly in practice, including partial-Kelly sizing — the closest published analogue to our `position_fraction` rule. |

## Verified institutional / education sources

- **MIT OpenCourseWare 15.401 Finance Theory I** (Prof. Andrew Lo) —
  **verified 2026-08-13**. Free lecture videos + notes: valuation,
  diversification, portfolio selection, efficient markets, options intro.
  https://ocw.mit.edu/courses/15-401-finance-theory-i-fall-2008/
- **SEC Investor.gov** — **verified 2026-08-13**. Official US regulator
  education hub: "Crypto Assets" section, red-flags-of-fraud list,
  fee and compounding tools. https://www.investor.gov/
- **CFA Institute Research Foundation** — site verified; free practitioner
  monographs (search "Trend Following with Managed Futures", Greyserman &
  Kaminski, 2014 — the most-cited serious trend-following monograph).
  https://rpc.cfainstitute.org/research-foundation
- **QuantStart** — **verified 2026-08-13**. Algorithmic trading education
  (QuarkGluon Ltd): position sizing, backtesting, risk management articles.
  Serious tone, no signal-selling. https://www.quantstart.com/
- **CME Group Education** `[Unverified — bot-blocked]` — futures
  fundamentals and professional risk-management material.
  https://www.cmegroup.com/education/
- **Binance official API docs** `[Unverified — JS-rendered]` — the
  authoritative exchange mechanics our bot already consumes:
  `LOT_SIZE`, `MIN_NOTIONAL`, `MARKET_LOT_SIZE` filters, fee schedules,
  error codes. https://developers.binance.com/docs/binance-spot-api-docs

## Books (practitioner standard, not guru material)

- Ernie Chan — *Algorithmic Trading* (2013) and *Quantitative Trading*
  (2008/2021). How real small-scale quant trading is done: stationarity,
  cointegration, execution, walk-forward validation.
- Robert Pardo — *The Evaluation and Optimization of Trading Strategies*
  (2008). The definitive text on **walk-forward analysis and out-of-sample
  testing** — the methodology the financial role must demand before any
  strategy change ships.
- Ralph Vince — *The Mathematics of Money Management* (1992). Fixed
  fractional sizing and **risk of ruin** math — the theory behind
  `position_fraction` and why 10% fixed fractional ≠ 10% of account risk.
- Van Tharp — *The Definitive Guide to Position Sizing* (2008). Position
  sizing as the primary risk lever; expectancy framing (win rate ×
  payoff ratio).
- Perry Kaufman — *Trading Systems and Methods* (6th ed., 2019). The
  encyclopedia of indicators/methods with a quantitative, skeptical eye.
- John Hull — *Options, Futures, and Other Derivatives*. Risk metrics
  (Greeks, VaR) and market mechanics.
- Philippe Jorion — *Value at Risk*. The institutional VaR/drawdown
  reference.
- Marcos López de Prado — *Advances in Financial Machine Learning* (2018).
  The financial-ML canon **with a caveat**: his own work argues most ML
  strategies fail validation; treat as a source of rigor, not of alpha.

## Concepts the financial role owns (mapped to the current spec)

| Concept | Where it already lives | Source anchor |
|---|---|---|
| Fractional position sizing | `position_fraction` 10% of free USDT (US5) | Vince, Thorp (partial Kelly) |
| Account-level exposure cap | `exposure_cap_fraction` 30% (US4) | Markowitz diversification |
| Daily loss breaker | `daily_loss_fraction` 5% → dry-run (US4/4b) | Risk-of-ruin math; institutional kill-switch practice |
| Fee realism | fees included in PnL; simulated 0.1% in dry-run (US4b, US5) | Makarov & Schoar microstructure |
| Win-rate weighting | pseudocount smoothing, 20-trade window, 10% floor (US3) | Bayesian smoothing; small-sample statistics (López de Prado's sample-size warnings) |
| Cooldown after loss | skip next cycle after a loss (US4) | Tilt/recency management (Chan, Pardo) |
| NO shorting, spot-only | Non-goals | Simplest robust posture; matches exchange reality |

## The anti-scam filter (what the role must REJECT)

Sources/claims that fail this bar and must never enter the spec:

- Any strategy sold by the person advertising it (signal subscriptions,
  paid groups, "mentorships").
- Claims without out-of-sample or walk-forward evidence; "99% win rate";
  backtests with no fees/slippage; martingale / doubling-down sizing.
- Reddit/Twitter/YouTube "alpha" with no testable rules.
- Any rule that increases size after a loss (anti-martingale is the only
  defensible direction for a small account).

## Open questions for the user (spec phase)

1. Should the financial role own the *definitions* of risk constants
   (10% slice, 30% cap, 5% breaker) or only *review* changes to them?
   (The spec currently treats the risk layer as deliberately static.)
2. Do you want a written "risk policy" doc (sizing rationale, maximum
   drawdown tolerance, when the bot gets shut off) as the role's first
   deliverable?
3. Is VaR or max-drawdown-based reporting wanted for the dashboard
   (currently: equity history only)?
