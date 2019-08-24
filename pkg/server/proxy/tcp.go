package proxy

import (
	"context"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/srvc/ery"
	"go.uber.org/zap"
)

func NewTCPServer(src, dest *ery.Addr) *TCPServer {
	return &TCPServer{
		src:  src,
		dest: dest,
		log:  zap.L().Named("proxy").Named("tcp"),
	}
}

type TCPServer struct {
	src, dest *ery.Addr
	log       *zap.Logger
}

func (s *TCPServer) Serve(ctx context.Context) error {
	addr, err := net.ResolveTCPAddr("tcp", s.src.String())
	if err != nil {
		return err
	}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		lis.Close()
	}()

	for {
		srcConn, err := lis.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok {
				if ne.Temporary() {
					s.log.Warn("failed to accept tcp connection", zap.Error(err))
					continue
				}
			}
			if isErrorClosedConn(err) {
				select {
				case <-ctx.Done():
					return nil // already shutdowned
				default:
					// no-op
				}
			}
			s.log.Warn("failed to accept tcp connection", zap.Error(err))
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleConn(ctx, srcConn)
		}()
	}
}

func (s *TCPServer) handleConn(ctx context.Context, srcConn *net.TCPConn) {
	s.log.Debug("start handling the connection")
	defer s.log.Debug("finish handling the connection")

	cp := func(src, dest *net.TCPConn, wg *sync.WaitGroup) {
		defer wg.Done()
		_, err := io.Copy(dest, src)
		if err != nil && !isErrorClosedConn(err) {
			// TODO: handle errors
			s.log.Warn("failed to copy packets", zap.Error(err))
			return
		}
	}

	destAddr, err := net.ResolveTCPAddr("tcp", s.dest.String())
	if err != nil {
		// TODO: handle error
		return
	}
	destConn, err := net.DialTCP("tcp", nil, destAddr)
	if err != nil {
		// TODO: handle error
		return
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		srcConn.Close()
		destConn.Close()
	}()

	wg.Add(2)
	go cp(srcConn, destConn, &wg)
	go cp(destConn, srcConn, &wg)
}

func isErrorClosedConn(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}
