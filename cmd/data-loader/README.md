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
**load securities:** To load the securities from master list
```bash
data-loader load securities
```
**load metrics:** To load the metrics from master list
```bash
data-loader load metrics
```
**load market-holidays:** To load the market holidays from master list
```bash
data-loader load market-holidays
```
**load ltp:** To load last traded price for the securities
```bash
data-loader load ltp
```
```bash
data-loader load ltp --isin=INE883A01011
```
**load security-stats:** To load open, high, close and volume stats for the securities
```bash
data-loader load security-stats
```
```bash
data-loader load security-stats --isin=INE883A01011
```
```bash
data-loader load security-stats --start-date=2024-01-01 --end-date=2024-12-31
```
```bash
data-loader load security-stats --isin=INE883A01011 --start-date=2024-01-01 --end-date=2024-12-31
```
**load security-metrics:** To load configured metrics like SMA, EMA, RSI etc. for the securities
```bash
data-loader load security-metrics
```
```bash
data-loader load security-metrics --isin=INE883A01011
```
```bash
data-loader load security-metrics --start-date=2024-01-01 --end-date=2024-12-31
```
```bash
data-loader load security-metrics --isin=INE883A01011 --start-date=2024-01-01 --end-date=2024-12-31
```

