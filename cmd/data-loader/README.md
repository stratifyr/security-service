# Data Loader

## Setup
**Download the binary:** Use the below command to install the data-loader binary
```bash
go install github.com/stratifyr/security-service/cmd/data-loader@latest
```
**Configure envs:** Set credentials for the data provider, host of your security-service.
```bash
export MARKET_DATA_PROVIDER=DHAN_MARKET_API
export DHAN_CLIENT_ID=1111111111
export DHAN_TOTP_SECRET=ABCDEFGHIJKL123456
export DHAN_PIN=111111
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
data-loader load ltp --symbol=HDFCBANK
```
**load security-stats:** To load open, high, close and volume stats for the securities
```bash
data-loader load security-stats
```
```bash
data-loader load security-stats --symbol=HDFCBANK
```
**backfill security-stats:** To backfill the open, high, close and volume stats for the securities to accommodate corporate actions
```bash
data-loader backfill security-stats
```
```bash
data-loader backfill security-stats --symbol=HDFCBANK
```
