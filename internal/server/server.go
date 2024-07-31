package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/LK4D4/trylock"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"github.com/astriaorg/seq-faucet/internal/chain"
	"github.com/astriaorg/seq-faucet/web"
)

type Server struct {
	chain.TxBuilder
	mutex trylock.Mutex
	cfg   *Config
	queue chan string
}

func NewServer(builder chain.TxBuilder, cfg *Config) *Server {
	return &Server{
		TxBuilder: builder,
		cfg:       cfg,
		queue:     make(chan string, cfg.queueCap),
	}
}

func (s *Server) setupRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(web.Dist()))
	limiter := NewLimiter(s.cfg.proxyCount, time.Duration(s.cfg.interval)*time.Minute)
	router.Handle("/api/claim", negroni.New(limiter, negroni.Wrap(s.handleClaim())))
	router.Handle("/api/info", s.handleInfo())

	return router
}

func (s *Server) Run() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			s.consumeQueue()
		}
	}()

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(s.setupRouter())
	log.Infof("Starting http server %d", s.cfg.httpPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(s.cfg.httpPort), n))
}

func (s *Server) consumeQueue() {
	if len(s.queue) == 0 {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	for len(s.queue) != 0 {
		address := <-s.queue
		log.Infof("DEBUG: cfg value before transfer from queue is %s", s.cfg.payoutNano)
		txHash, err := s.Transfer(context.Background(), address, s.cfg.payoutNano)
		log.Infof("DEBUG: cfg value after transfer from queue is %s", s.cfg.payoutNano)
		if err != nil {
			log.WithError(err).Error("Failed to handle transaction in the queue")
		} else {
			log.WithFields(log.Fields{
				"txHash":  txHash,
				"address": address,
			}).Info("Consume from queue successfully")
		}
	}
}

func (s *Server) handleClaim() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}

		// The error always be nil since it has already been handled in limiter
		address, _ := readAddress(r)
		// Try to lock mutex if the work queue is empty
		if len(s.queue) != 0 || !s.mutex.TryLock() {
			select {
			case s.queue <- address:
				log.WithFields(log.Fields{
					"address": address,
				}).Info("Added to queue successfully")
				resp := claimResponse{Message: fmt.Sprintf("Added %s to the queue", address)}
				renderJSON(w, resp, http.StatusOK)
			default:
				log.Warn("Max queue capacity reached")
				renderJSON(w, claimResponse{Message: "Faucet queue is too long, please try again later"}, http.StatusServiceUnavailable)
			}
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		log.Infof("DEBUG: cfg value before transfer from handleClaim is %s", s.cfg.payoutNano)
		txHash, err := s.Transfer(ctx, address, s.cfg.payoutNano)
		log.Infof("DEBUG: cfg value after transfer from handleClaim is %s", s.cfg.payoutNano)
		s.mutex.Unlock()
		if err != nil {
			log.WithError(err).Error("Failed to send transaction")
			renderJSON(w, claimResponse{Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{
			"txHash":  txHash,
			"address": address,
		}).Info("Funded directly successfully")
		resp := claimResponse{Message: fmt.Sprintf("Txhash: %s", txHash)}
		renderJSON(w, resp, http.StatusOK)
	}
}

func (s *Server) handleInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		renderJSON(w, infoResponse{
			Account: s.Sender(),
			Network: s.cfg.network,
			Payout:  s.cfg.payout.String(),
		}, http.StatusOK)
	}
}
