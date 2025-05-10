# Data Loader

## Setup
**Download the binary:** Use the below command to install the data-loader binary
```bash
go install github.com/stratifyr/security-service/cmd/data-loader@latest
```
**Configure envs:** Set credentials for the data provider, host of your security-service.
```bash
export MARKET_DATA_PROVIDER=DHAN_MARKET_API
export DHAN_API_KEY=aaa123aaa.bbb123bbb.ccc123ccc.ddd123dddd
export DHAN_CLIENT_ID=1234567890
export SECURITY_SERVICE_HOST=http://localhost:8000
```

## Commands
**load ltp:** To load last traded price for the securities
```bash
data-loader load ltp
```
```bash
data-loader load ltp --isin=INE883A01011
```
**load ohlc:** To load open, high, close and volume stats for the securities for this day
```bash
data-loader load ohlc
```
```bash
data-loader load ohlc --isin=INE883A01011
```
**load historical-ohlc:** To load open, high, close and volume stats for the securities for the given interval
```bash
data-loader load historical-ohlc --start-date=2024-01-01 --end-date=2024-12-31
```
```bash
data-loader load historical-ohlc --isin=INE883A01011 --start-date=2024-01-01 --end-date=2024-12-31
```
**load metrics:** To load metrics like SMA, EMA, RSI etc. for the securities for this day
```bash
data-loader load metrics
```
```bash
data-loader load metrics --isin=INE883A01011
```
**load historical-metrics:** To load metrics like SMA, EMA, RSI etc. for the securities for the given interval
```bash
data-loader load historical-metrics --start-date=2024-01-01 --end-date=2024-12-31
```
```bash
data-loader load historical-metrics --isin=INE883A01011 --start-date=2024-01-01 --end-date=2024-12-31
```

