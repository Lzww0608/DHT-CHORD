package node

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

const (
	M      = 16
	MBytes = M / 8
)

type ChordNode struct {
	Id      []byte
	Address string
}

type ChordServer struct {
	self            *ChordNode
	predecessor     *ChordNode
	finger          []*ChordNode
	fix_finger_next uint
	mux             *sync.Mutex
	logger          *log.Entry
	storage         map[string]string
}

func (s *ChordServer) sucessor() *ChordNode {
	if len(s.finger) >= 1 {
		return s.finger[0]
	} else {
		return nil
	}
}

func NewChordNode(addr string) *ChordNode {
	return &ChordNode{
		Id:      generateChordHash(addr),
		Address: addr,
	}
}
