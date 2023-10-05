package cmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/astriaorg/seq-faucet/internal/chain"
	"github.com/astriaorg/seq-faucet/internal/server"
)

var (
	appVersion = "v1.1.0"

	httpPortFlag = flag.Int("httpport", 8080, "Listener port to serve HTTP connection")
	proxyCntFlag = flag.Int("proxycount", 0, "Count of reverse proxies in front of the server")
	queueCapFlag = flag.Int("queuecap", 100, "Maximum transactions waiting to be sent")
	versionFlag  = flag.Bool("version", false, "Print version number")

	payoutFlag   = flag.Int("faucet.amount", 1, "Number of Sequencer tokens to transfer per user request")
	intervalFlag = flag.Int("faucet.minutes", 1440, "Number of minutes to wait between funding rounds")
	netnameFlag  = flag.String("faucet.name", "Astria Sequencer Network", "Network name to display on the frontend")

	privKeyFlag  = flag.String("wallet.privkey", os.Getenv("PRIVATE_KEY"), "Private key hex to fund user requests with")
	providerFlag = flag.String("wallet.provider", os.Getenv("WEB3_PROVIDER"), "Endpoint for Ethereum JSON-RPC connection")
)

func init() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}
}

func Execute() {
	privateKey, err := getPrivateKeyFromFlag()
	if err != nil {
		panic(fmt.Errorf("failed to read private key: %w", err))
	}

	txBuilder, err := chain.NewTxBuilder(*providerFlag, privateKey)
	if err != nil {
		panic(fmt.Errorf("cannot connect to web3 provider: %w", err))
	}
	config := server.NewConfig(*netnameFlag, *httpPortFlag, *intervalFlag, *payoutFlag, *proxyCntFlag, *queueCapFlag)
	go server.NewServer(txBuilder, config).Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func getPrivateKeyFromFlag() (*ed25519.PrivateKey, error) {
	if *privKeyFlag == "" {
		return nil, errors.New("no private key provided")
	}

	hexkey := *privKeyFlag
	if chain.Has0xPrefix(hexkey) {
		hexkey = hexkey[2:]
	}

	privateKeyBytes, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, err
	}

	if len(privateKeyBytes) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid private key length, expected 32: %d", len(privateKeyBytes))
	}

	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	return &privateKey, nil
}
