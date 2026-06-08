

# Introduction
This is for getting the best algorithm for BTC and TQQQ investment using historical price and some matrics.


# BackTest Result
2026-05-22 

## TQQQ BackTest
Price follows the tqqq.csv file `AdjClose` column.

| Index | Initial Amount (USD) | Buy Condition                                                                                                                                                        | Buy Amount (USD)                               | Sell Condition                                 | Sell Amount | MDD | Final Account (USD) | ROI |
|-------|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------|------------------------------------------------|-------------|-----|---------------------|-----|
| 1     | Year * $12,000       | Start date                                                                                                                                                           | 100%                                           | Never                                          | 0%          |     |                     |     |
| 2     | $0                   | Every January 1st                                                                                                                                                    | $12,000                                        | Never                                          | 0%          |     |                     |     |
| 3     | $0                   | Every 25th monthly                                                                                                                                                   | $1,000                                         | Never                                          | 0%          |     |                     |     |
| 4     | $0                   | CNN F&G Index <= 24 (Extreme Fear) once in a month, and deposit cash every month 25th day, 1000$ until next buy date                                                 | Save 1,000$. Buy it using all accumulated cash | Never                                          | 0%          |     |                     |     |
| 5     | $0                   | CNN F&G index <= 24 (Extreme Fear) 50% once in a month, index <= 15 80% cash,  inex <= 10 100% cash and deposit cash every month 25th day, 1000$ until next buy date | Save 1,000$. Buy it using all accumulated cash | Never                                          | 0%          |     |                     |     |
| 6     | $0                   | CNN F&G index <= 24 (Extreme Fear) once in a month, and deposit cash every month 25th day, 1000$ until next buy date                                                 | Save 1,000$. Buy it using all accumulated cash | CNN Fear and Greed Index >= 76 (Extreme Greed) | 100%        |     |                     |     |

## Tax calculation

| Stock | Capital Gains Tax | Tax-free Amount | Gross Expense Ratio (per year) |
|-------|-------------------|-----------------|--------------------------------|
| TQQQ  | 22%               | 1,700 $         | 0.97%                          |
| QLD   | 22%               | 1,700 $         | 0.95%                          |
| KQLD  | 9.9%              | 1,333 $         | 0.3372%                        |
| QQQ   | 22%               | 1,700 $         | 0.18%                          |

What's KQLD?
미래에셋 TIGER 미국나스닥100레버리지증권상장지수투자신탁(주식혼합-파생형)(합성)[K55301DQ3577]

"KQLD is a leveraged ETF that tracks 2x the daily price movement, just like QLD. 
Its expense ratio is 0.3372%, which is 0.6128% cheaper than QLD."

> You don't have to care about Gross Expense Ratio TQQQ, QLD, QQQ Since it already reflected in the price. 
> But for KQLD, you have to consider it because it's synthetic ETF and the price doesn't reflect the expense ratio. 
> So I applied discounted Gross Expense Ratio on the KQLD

## BTC BackTest

No Tax for BTC in Korea.

| Index | Initial Amount (USD) | Buy Condition                                                                                         | Buy Amount (USD)                               | Sell Condition                      | Sell Amount | MDD | Final Account (USD) | ROI |
|-------|----------------------|-------------------------------------------------------------------------------------------------------|------------------------------------------------|-------------------------------------|-------------|-----|---------------------|-----|
| 1     | Year * $12,000       | Start date                                                                                            | 100%                                           | Never                               | 0%          |     |                     |     ||
| 2     | $0                   | Every January 1st                                                                                     | $12,000                                        | Never                               | 0%          |     |                     |     ||
| 3     | $0                   | Every 25th monthly                                                                                    | $1,000                                         | Never                               | 0%          |     |                     |     |
| 4     | $0                   | MVRV/Z-Score <= 0 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | Never                               | 0%          |     |                     |     ||
| 5     | $0                   | MVRV/Z-Score <= 0 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 7 once in a month   | 100%        |     |                     |     ||
| 6     | $0                   | MVRV/Z-Score <= 0 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 6 once in a month   | 100%        |     |                     |     ||
| 7     | $0                   | MVRV/Z-Score <= 0 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 3.5 once in a month | 100%        |     |                     |     ||
| 8     | $0                   | MVRV/Z-Score <= 0 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 3.5 once in a month | 100%        |     |                     |     ||
| 9     | $0                   | MVRV/Z-Score <= 0.5 once in a month, and deposit cash every month 25th day, 1000$ until next buy date | Save 1,000$. Buy it using all accumulated cash | Never                               | 0%          |     |                     |     ||
| 10    | $0                   | MVRV/Z-Score <= 0.5 once in a month, and deposit cash every month 25th day, 1000$ until next buy date | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 7 once in a month   | 100%        |     |                     |     ||
| 11    | $0                   | MVRV/Z-Score <= 0.5 once in a month, and deposit cash every month 25th day, 1000$ until next buy date | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 6 once in a month   | 100%        |     |                     |     ||
| 12    | $0                   | MVRV/Z-Score <= 0.5 once in a month, and deposit cash every month 25th day, 1000$ until next buy date | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 3.5 once in a month | 100%        |     |                     |     ||
| 13    | $0                   | MVRV/Z-Score <= 0.5 once in a month, and deposit cash every month 25th day, 1000$ until next buy date | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 3.5 once in a month | 100%        |     |                     |     ||
| 14    | $0                   | MVRV/Z-Score <= 1 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | Never                               | 0%          |     |                     |     ||
| 15    | $0                   | MVRV/Z-Score <= 1 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 7 once in a month   | 100%        |     |                     |     ||
| 16    | $0                   | MVRV/Z-Score <= 1 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 6 once in a month   | 100%        |     |                     |     ||
| 17    | $0                   | MVRV/Z-Score <= 1 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 3.5 once in a month | 100%        |     |                     |     ||
| 18    | $0                   | MVRV/Z-Score <= 1 once in a month, and deposit cash every month 25th day, 1000$ until next buy date   | Save 1,000$. Buy it using all accumulated cash | MVRV/Z-Score >= 3.5 once in a month | 100%        |     |                     |     ||

## USD-KRW BackTest
Can buy Kimchi discount and sell it Kimchi premium using USD-KRW exchange rate.

Kimchi premium is difference rate between USDT-KRW - USD-KRW.
e.g. USDT-KRW 1500 won and USD-KRW 1485 won, then Kimchi premium is 15 won (1%)

| Index | Initial Amount (KRW) | Buy Condition             | Buy Amount (KRW)                    | Sell Condition           | Sell Amount |
|-------|----------------------|---------------------------|-------------------------------------|--------------------------|-------------|
| 1     | 10,000,000           | Never                     | Never                               | Never                    | Never       |
| 2     | 0                    | Every 25th monthly        | 10,000,000 / months period          | Never                    | Never       |
| 3     | 0                    | Kimchi Premium <= 0%      | first 10,000,000 after current cash | Kimchi Premium >= 2%     | 100%        |
| 4     | 0                    | Kimchi Premium <= -1%     | first 10,000,000 after current cash | Kimchi Premium >= 2%     | 100%        |
| 5     | 0                    | Kimchi Premium <= -2%     | first 10,000,000 after current cash | Kimchi Premium >= 2%     | 100%        |
| 6     | 0                    | Kimchi Premium <= 0%      | first 10,000,000 after current cash | Kimchi Premium >= 3%     | 100%        |
| 7     | 0                    | Kimchi Premium <= -1%     | first 10,000,000 after current cash | Kimchi Premium >= 3%     | 100%        |
| 8     | 0                    | Kimchi Premium <= -2%     | first 10,000,000 after current cash | Kimchi Premium >= 3%     | 100%        |
| 9     | 0                    | Kimchi Premium <= 10 KRW  | first 10,000,000 after current cash | Kimchi Premium >= 21 KRW | 100%        |
| 10    | 0                    | Kimchi Premium <= 10 KRW  | first 10,000,000 after current cash | Kimchi Premium >= 30 KRW | 100%        |
| 11    | 0                    | Kimchi Premium <= 10 KRW  | first 10,000,000 after current cash | Kimchi Premium >= 40 KRW | 100%        |
| 12    | 0                    | Kimchi Premium <= 5 KRW   | first 10,000,000 after current cash | Kimchi Premium >= 21 KRW | 100%        |
| 13    | 0                    | Kimchi Premium <= 4 KRW   | first 10,000,000 after current cash | Kimchi Premium >= 21 KRW | 100%        |
| 14    | 0                    | Kimchi Premium <= 3 KRW   | first 10,000,000 after current cash | Kimchi Premium >= 21 KRW | 100%        |
| 15    | 0                    | Kimchi Premium <= 2 KRW   | first 10,000,000 after current cash | Kimchi Premium >= 21 KRW | 100%        |
| 16    | 0                    | Kimchi Premium <= 1 KRW   | first 10,000,000 after current cash | Kimchi Premium >= 21 KRW | 100%        |
| 17    | 0                    | Kimchi Premium <= 0 KRW   | first 10,000,000 after current cash | Kimchi Premium >= 21 KRW | 100%        |
| 18    | 0                    | Kimchi Premium <= 5 KRW   | first 10,000,000 after current cash | Kimchi Premium >= 30 KRW | 100%        |
| 19    | 0                    | Kimchi Premium <= 5 KRW   | first 10,000,000 after current cash | Kimchi Premium >= 40 KRW | 100%        |
| 20    | 0                    | Kimchi Premium <= -10 KRW | first 10,000,000 after current cash | Kimchi Premium >= 10 KRW | 100%        |
| 21    | 0                    | Kimchi Premium <= -10 KRW | first 10,000,000 after current cash | Kimchi Premium >= 21 KRW | 100%        |
| 22    | 0                    | Kimchi Premium <= -10 KRW | first 10,000,000 after current cash | Kimchi Premium >= 30 KRW | 100%        |
| 23    | 0                    | USD-KRW <= 1400 won       | first 10,000,000 after current cash | USD-KRW >= 1500 won      | 100%        |
| 24    | 0                    | USD-KRW <= 1450 won       | first 10,000,000 after current cash | USD-KRW >= 1500 won      | 100%        |
| 25    | 0                    | USD-KRW <= 1420 won       | first 10,000,000 after current cash | USD-KRW >= 1475 won      | 100%        |

In Upbit, trading fee is 0.05% for both maker and taker, and the price is based on the order book's current price. So I applied 0.05% fee on the buy and sell amount.


#### Final value calculation
Final value is calculated like this:
- Last day asset is KRW → final value is last sell USDT amount * USDT price of KRW
- Last day asset is USDT → final value is evaluated last day USDT amount * USDT price of KRW


## Metrics
#### Avg hold days
buy - sell period in days.
e.g. 
- 2025-01-01 buy, 2025-01-15 sell -> 14 days
- 2025-02-01 buy, 2025-02-28 sell -> 27 days 

Avg hold days = (14 + 27) / 2 = 20.5 days

#### Trade count
How many buy and sell count according to the conditions.
- 2025-01-01 buy, 2025-01-15 sell -> 2 trades
- 2025-02-01 buy, 2025-02-28 sell -> 2 trades

Trade count: 4 (2 + 2)


## Data Source
#### TQQQ
- [Historical price - Yahoo Finance](https://finance.yahoo.com/quote/TQQQ/history?p=TQQQ)
- [Fear and Greed Index - CNN](https://edition.cnn.com/markets/fear-and-greed)

#### BTC
- [MVRV/Z-Score and Historical Price](https://charts.bitbo.io/mvrv-z-score/) 

#### USD
- [Historical Price - Yahoo Finance](https://finance.yahoo.com/quote/KRW%3DX/history)
- [USDT Price - Upbit Developer](https://docs.upbit.com/docs/upbit-orderbook-currentWh)
