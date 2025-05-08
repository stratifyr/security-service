# Data Loader

## Setup
**Download the binary:** Use the below command to install the data-loader binary
```bash
go install github.com/stratifyr/security-service/cmd/data-loader@latest
```
**Configure envs:** Set credentials for the data provider, host of your security-service, and market holidays based on the time of run.
```bash
export MARKET_DATA_PROVIDER=DHAN_MARKET_API
export DHAN_API_KEY=aaa123aaa.bbb123bbb.ccc123ccc.ddd123dddd
export DHAN_CLIENT_ID=1234567890
export SECURITY_SERVICE_HOST=http://localhost:8000
export MARKET_HOLIDAYS=2025-01-26,2025-03-17,2025-04-14,2025-04-18,2025-05-01,2025-08-15,2025-10-02,2025-10-23,2025-10-31,2025-12-25
```

## Commands
**load ltp:** To load last traded price for all the securities
```bash
data-loader load ltp
```
**load ohlc:** To load open, high, close and volume stats for all the securities for this day
```bash
data-loader load ohlc
```
**load historical-ohlc:** To load open, high, close and volume stats for all the securities for the given interval
```bash
data-loader load historical-ohlc --start-date=2025-01-01 --end-date=2025-05-01
```

