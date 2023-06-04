package stun

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/ArminGh02/golang-p2p-messenger/internal/peer"
	"github.com/ArminGh02/golang-p2p-messenger/internal/request"
	"github.com/ArminGh02/golang-p2p-messenger/internal/response"
	"github.com/ArminGh02/golang-p2p-messenger/internal/stun/repository"
)

type Stun struct {
	repo   repository.Repository[*peer.Peer]
	logger *logrus.Logger // TODO: use interface
}

func New(repo repository.Repository[*peer.Peer], logger *logrus.Logger) *Stun {
	return &Stun{
		repo:   repo,
		logger: logger,
	}
}

func (s *Stun) PeerHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.postPeer(w, r)
		case http.MethodGet:
			s.getPeer(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func (s *Stun) postPeer(w http.ResponseWriter, r *http.Request) {
	var (
		req  request.PostPeer
		resp response.PostPeer
		enc  = json.NewEncoder(w)
	)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = fmt.Sprintf("error decoding request: %v", err)
		enc.Encode(resp)
		return
	}

	usernameExists, err := s.repo.Exists(context.Background(), req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = fmt.Sprintf("error checking if username %q exists: %v", req.Username, err)
		enc.Encode(resp)
		return
	}

	if usernameExists {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = fmt.Sprintf("username %q already exists", req.Username)
		enc.Encode(resp)
		return
	}

	peer := &peer.Peer{
		UDPAddr:  req.UDPAddr, // TODO: why not use address in http.Request.RemoteAddr?
		TCPAddr:  req.TCPAddr,
		Username: req.Username,
	}
	if err := s.repo.Set(context.Background(), peer.Username, peer); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = fmt.Sprintf("error adding peer: %v", err)
		enc.Encode(resp)
		return
	}

	resp.OK = true
	w.WriteHeader(http.StatusOK)
	enc.Encode(resp)
}

func (s *Stun) getPeer(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Path[len("/peer/"):]

	if username == "" {
		s.listPeers(w, r)
		return
	}

	s.peerByUsername(w, username)
}

func (s *Stun) listPeers(w http.ResponseWriter, r *http.Request) {
	var (
		resp response.GetPeer
		enc  = json.NewEncoder(w)
	)

	peers, err := s.repo.Values(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = fmt.Sprintf("error listing peers: %v", err)
		enc.Encode(resp)
		return
	}

	resp.OK = true
	resp.Peers = peers
	w.WriteHeader(http.StatusOK)
	enc.Encode(resp)
}

func (s *Stun) peerByUsername(w http.ResponseWriter, username string) {
	var (
		resp response.GetPeer
		enc  = json.NewEncoder(w)
	)

	p, err := s.repo.Get(context.Background(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = fmt.Sprintf("error getting peer: %v", err)
		enc.Encode(resp)
		return
	}

	resp.OK = true
	resp.Peers = []*peer.Peer{p}
	w.WriteHeader(http.StatusOK)
	enc.Encode(resp)
}
