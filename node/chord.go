package node

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	pb "DHT/protos"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Chord identifiers use an M-bit hash ring.
const (
	M      = 16
	MBytes = M / 8
)

type ChordNode struct {
	Id      []byte
	Address string
}

// ChordServer encapsulates the local node state and chord routing metadata.
type ChordServer struct {
	self            *ChordNode
	predecessor     *ChordNode
	finger          []*ChordNode
	fix_finger_next uint
	mu              *sync.Mutex
	logger          *log.Entry
	storage         map[string]string
}

// sucessor returns the first entry in the finger table, if present.
func (s *ChordServer) successor() *ChordNode {
	if len(s.finger) >= 1 {
		return s.finger[0]
	} else {
		return nil
	}
}

// NewChordNode builds a node instance for the provided network address.
func NewChordNode(addr string) *ChordNode {
	return &ChordNode{
		Id:      generateChordHash(addr),
		Address: addr,
	}
}

func NewChordServer(addr string) *ChordServer {
	self := &ChordNode{
		generateChordHash(addr),
		addr,
	}

	finger := make([]*ChordNode, M)
	finger[0] = self
	for i := 1; i < M; i++ {
		finger[i] = nil
	}

	return &ChordServer{
		self:            self,
		predecessor:     nil,
		finger:          finger,
		fix_finger_next: 0,
		mu:              &sync.Mutex{},
		logger: log.WithFields(log.Fields{
			"id": fmt.Sprintf("%X", self.Id),
		}),
		storage: make(map[string]string),
	}
}

func (s *ChordServer) Serve(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			break
		default:
			{
				func() {
					ctx_, cancel := context.WithTimeout(ctx, time.Second)
					defer cancel()
					if err := s.Stabilize(ctx_); err != nil {
						s.logger.Warningf("Stabilize routine error %v", err)
					}
				}()
				time.Sleep(time.Millisecond * 5)
			}

			{
				func() {
					ctx_, cancel := context.WithTimeout(ctx, time.Second)
					defer cancel()
					if err := s.CheckPredecessor(ctx_); err != nil {
						s.logger.Warningf("Stabilize routine error %v", err)
					}
				}()
				time.Sleep(time.Millisecond * 5)
			}

			{
				func() {
					ctx_, cancel := context.WithTimeout(ctx, time.Second)
					defer cancel()
					if err := s.FixFingers(ctx_); err != nil {
						s.logger.Warningf("Stabilize routine error %v", err)
					}
				}()
				time.Sleep(time.Millisecond * 5)
			}
		}
	}
}

func (s *ChordServer) FindSuccessor(ctx context.Context, in *pb.FindSuccessorRequest) (*pb.Node, error) {
	s.mu.Lock()
	if inRange(in.Id, s.self.Id, s.successor().Id) {
		s.mu.Unlock()
		return &pb.Node{Id: s.successor().Id, Addr: s.successor().Address}, nil
	} else {

		n, err := s.ClosestPrecedingNode(ctx, in.Id)
		if err != nil {
			return nil, err
		}
		node, err := n.FindSuccessor(ctx, in.Id)
		if err != nil {
			return nil, err
		}

		s.mu.Unlock()
		return &pb.Node{Id: node.Id, Addr: node.Address}, nil
	}
}

func (s *ChordServer) Join(ctx context.Context, node *ChordNode) error {
	s.mu.Lock()
	s.predecessor = nil
	s.mu.Unlock()

	node, err := node.FindSuccessor(ctx, s.self.Id)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.finger[0] = node
	s.mu.Unlock()
	return nil
}

func (s *ChordServer) Stabilize(ctx context.Context) error {
	x, err := s.successor().FindPredecessor(ctx, s.self)
	if err != nil {
		return err
	}
	if inRangeExclude(x.Id, s.self.Id, s.successor().Id) {
		s.finger[0] = x
	}
	err = s.successor().Notify(ctx, s.self)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChordServer) Notify(ctx context.Context, in *pb.Node) (*pb.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.predecessor == nil ||
		inRangeExclude(in.Id, s.predecessor.Id, s.self.Id) {
		s.predecessor = &ChordNode{in.Id, in.Addr}
	}
	return &pb.Result{Result: "success"}, nil
}

func (s *ChordServer) FindPredecessor(ctx context.Context, in *pb.Node) (*pb.Node, error) {
	_, err := s.Notify(ctx, in)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.predecessor == nil {
		return nil, errors.New("no predecessor")
	}
	return &pb.Node{Id: s.predecessor.Id, Addr: s.predecessor.Address}, nil
}

func (s *ChordServer) Ping(ctx context.Context, in *pb.Node) (*pb.Void, error) {
	return &pb.Void{}, nil
}

func (s *ChordServer) FixFingers(ctx context.Context) error {
	s.mu.Lock()
	s.fix_finger_next = s.fix_finger_next + 1
	if s.fix_finger_next >= M {
		s.fix_finger_next = 0
	}
	next := s.fix_finger_next
	s.mu.Unlock()
	x, err := s.FindSuccessor(ctx, &pb.FindSuccessorRequest{Id: byteAddPowerOf2(s.self.Id, s.fix_finger_next)})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.finger[next] = &ChordNode{x.Id, x.Addr}
	defer s.mu.Unlock()
	return nil
}

func (s *ChordServer) CheckPredecessor(ctx context.Context) error {
	_, err := s.Notify(ctx, &pb.Node{Id: s.self.Id, Addr: s.self.Address})
	if err != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.predecessor = nil
	}
	return nil
}

func (s *ChordServer) ClosestPrecedingNode(ctx context.Context, id []byte) (*ChordNode, error) {
	for i := M - 1; i >= 0; i-- {
		if s.finger[i] == nil {
			continue
		}
		if inRangeExclude(s.finger[i].Id, s.self.Id, id) {
			return s.finger[i], nil
		}
	}
	return nil, nil
}

func (n *ChordNode) FindSuccessor(ctx context.Context, id []byte) (*ChordNode, error) {
	conn, err := grpc.Dial(n.Address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	defer conn.Close()
	c := pb.NewChordClient(conn)
	resp, err := c.FindSuccessor(ctx, &pb.FindSuccessorRequest{Id: id})
	if err != nil {
		return nil, err
	}

	return &ChordNode{resp.Id, resp.Addr}, nil
}

func (n *ChordNode) FindPredecessor(ctx context.Context, self *ChordNode) (*ChordNode, error) {
	conn, err := grpc.Dial(n.Address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewChordClient(conn)
	r, err := c.FindPredecessor(ctx, &pb.Node{Id: self.Id, Addr: self.Address})
	if err != nil {
		return nil, err
	}
	return &ChordNode{r.Id, r.Addr}, nil
}

func (n *ChordNode) Notify(ctx context.Context, self *ChordNode) error {
	conn, err := grpc.Dial(n.Address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := pb.NewChordClient(conn)
	_, err = c.Notify(ctx, &pb.Node{Id: self.Id, Addr: self.Address})
	if err != nil {
		return err
	}
	return nil
}
