# seq-faucet

[![Build](https://img.shields.io/github/actions/workflow/status/astriaorg/seq-faucet/build.yml?branch=main)](https://github.com/astriaorg/seq-faucet/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/astriaorg/seq-faucet)](https://github.com/astriaorg/seq-faucet/releases)
[![Report](https://goreportcard.com/badge/github.com/astriaorg/seq-faucet)](https://goreportcard.com/report/github.com/astriaorg/seq-faucet)
[![Go](https://img.shields.io/github/go-mod/go-version/astriaorg/seq-faucet)](https://go.dev/)
[![License](https://img.shields.io/github/license/astriaorg/seq-faucet)](https://github.com/astriaorg/seq-faucet/blob/main/LICENSE)

The faucet is a web application with the goal of distributing small amounts of Ether in private and test networks.

## Features

* Allow to configure the funding account via private key
* Asynchronous processing Txs to achieve parallel execution of user requests
* Rate limiting by ETH address and IP address as a precaution against spam
* Prevent X-Forwarded-For spoofing by specifying the count of reverse proxies

## Get started

### Prerequisites

* Go (1.16 or later)
* Node.js
* Yarn

### Installation

1. Clone the repository and navigate to the app’s directory
```bash
git clone https://github.com/astriaorg/seq-faucet.git
cd seq-faucet
```

2. Bundle Frontend web with Vite
```bash
go generate
```

3. Build Go project
```bash
go build -o seq-faucet
```

## Usage

**Use private key to fund users**

```bash
./seq-faucet -wallet.provider tcp://sequencer.localdev.me:80 -wallet.privkey privkey
```

### Configuration

You can configure the funder by using environment variables instead of command-line flags as follows:
```bash
export WEB3_PROVIDER=tcp://sequencer.localdev.me:80
export PRIVATE_KEY=0x...
```

Then run the faucet application without the wallet command-line flags:
```bash
./seq-faucet -httpport 8080
```

**Optional Flags**

The following are the available command-line flags(excluding above wallet flags):

| Flag            | Description                                      | Default Value  |
|-----------------|--------------------------------------------------|----------------|
| -httpport       | Listener port to serve HTTP connection           | 8080           |
| -proxycount     | Count of reverse proxies in front of the server  | 0              |
| -queuecap       | Maximum transactions waiting to be sent          | 100            |
| -faucet.amount  | Number of Ethers to transfer per user request    | 1              |
| -faucet.minutes | Number of minutes to wait between funding rounds | 1440           |
| -faucet.name    | Network name to display on the frontend          | testnet        |

### Docker deployment

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=tcp://sequencer.localdev.me:80 -e PRIVATE_KEY=0x... astriaorg/seq-faucet:1.1.0
```

#### Build the Docker image

```bash
docker buildx build -t ghcr.io/astriaorg/seq-faucet:0.0.1-local .
```

## License

Distributed under the MIT License. See LICENSE for more information.
