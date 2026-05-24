

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


## Data Source
#### TQQQ
- [Historical price - Yahoo Finance](https://finance.yahoo.com/quote/TQQQ/history?p=TQQQ)
- [Fear and Greed Index - CNN](https://edition.cnn.com/markets/fear-and-greed)

#### BTC
- [MVRV/Z-Score and Historical Price](https://charts.bitbo.io/mvrv-z-score/) 

