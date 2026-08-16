# Data Loader

## Setup
**Download the binary:** Use the below command to install the data-loader binary
```bash
go install github.com/stratifyr/security-service/cmd/data-loader@latest
```
**Configure envs:** Set host of your security-service.
```bash
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