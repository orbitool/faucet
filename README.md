# Simple Ethereum faucet

A simple backend for distributing test Ethereum written in go.

> ⚠️ use at your own risk.

### Features:

- Configure amount and frequency ETH can be claimed.
- Optional captcha protection to mitigate bots.
- Supports any EVM network.

# Quick start

### 1. Start the faucet:

```bash
go run cli/main.go --rpc <your-rpc-url> --key <your-private-key> --delay 600
--port 8080
```

Starts the faucet with claiming delay of
10 mins (600 seconds).

### 2. Claim some ETH

```bash
curl http://localhost:8080/claim?address=<your-address>
```

Faucet will respond with receipt containing the transaction hash.

# Configuration

| Flag          | Env                | Required | Default | Description                                                            |
| ------------- | ------------------ | -------- | ------- | ---------------------------------------------------------------------- |
| --rpc         | FAUCET_RPC         | true     | N/A     | The url of your networks rpc endpoint                                  |
| --key         | FAUCET_KEY         | true     | N/A     | Private key of the faucets wallet                                      |
| --amount      | FAUCET_AMOUNT      | false    | 0.01    | Amount of eth that can be claimed, value in ETH                        |
| --port        | PORT               | false    | 8080    | Port this service will be served on                                    |
| --delay       | FAUCET_DELAY       | false    | 43200   | Amount of time in seconds a user must wait before making another claim |
| --captcha     | FAUCET_CAPTCHA     | false    | false   | Require a captcha to be solved for each claim request                  |
| --chain       | FAUCET_CHAIN       | false    | N/A     | Chain id. If not provided will be fetched from the rpc provider        |
| --cors        | FAUCET_CORS        | false    | "\*"    | Http request allowed origin                                            |
| --proxy-count | FAUCET_PROXY_COUNT | false    | 0       | The number of proxies in front of this service                         |

# Using Captcha

When captcha is enabled users must solve a challenge before claiming.

### 1. Create a new Captcha

```
curl http://localhost:8080/captcha/create
```

Will create a new captcha and return the **capctha_id**.

### 2. View the captcha image

`http://localhost:8080/captcha/challenge/<captcha_id>.png`

### 3. Solve and Claim

Post the solution in the header along with the claim request

```
curl https://localhost:8080/claim?address=<your-address> \
  -H "X-Captcha: <captcha_id> \
  -H "X-Captcha-Solution: <captcha_solution>"
```
