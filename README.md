

# Introduction
This is for getting best algorithm for BTC and TQQQ investment using historical price and some matrics.


# BackTest Result
@Today 2026-05-17 

## TQQQ BackTest
Price follows the tqqq.csv file `AdjClose` column.

| Index | Initial Amount (USD) | Buy Condition                                                 | Buy Amount (USD)              | Sell Condition                                 | Sell Amount | MDD | Final Account (USD) | ROI |
|-------|----------------------|---------------------------------------------------------------|-------------------------------|------------------------------------------------|-------------|-----|---------------------|-----|
| 1     | Year * $12,000       | Start date                                                    | 100%                          | Never                                          | 0%          |     |                     |     |
| 2     | $0                   | Every January 1st                                             | $12,000                       | Never                                          | 0%          |     |                     |     |
| 3     | $0                   | Every 25th monthly                                            | $1,000                        | Never                                          | 0%          |     |                     |     |
| 4     | $0                   | Every 1st monthly                                             | $1,000                        | Never                                          | 0%          |     |                     |     |
| 5     | $0                   | CNN Fear and Greed Index <= 24 (Extreme Fear) once in a month | First $12,000 (Previous 100%) | Never                                          | 0%          |     |                     |     |
| 6     | $0                   | CNN Fear and Greed Index <= 24 (Extreme Fear) once in a month | First $12,000 (Previous 100%) | CNN Fear and Greed Index >= 76 (Extreme Greed) | 100%        |     |                     |     |
| 7     | $0                   | CNN Fear and Greed Index <= 24 (Extreme Fear) once in a month | First $12,000 (Previous 100%) | CNN Fear and Greed Index >= 60 (Greed)         | 100%        |     |                     |     |


## BTC BackTest




## Data Source
#### TQQQ
- [Historical price - Yahoo Finance](https://finance.yahoo.com/quote/TQQQ/history?p=TQQQ)
- [Fear and Greed Index - CNN](https://edition.cnn.com/markets/fear-and-greed)

#### BTC
- [Historical price - Coingecko](https://www.coingecko.com/en/coins/bitcoin/historical_data)
- MVRV/Z-Score 

