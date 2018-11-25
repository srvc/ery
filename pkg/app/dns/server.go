package dns

import (
	"context"
	"fmt"
	"net"
	"strings"

	godns "github.com/miekg/dns"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/domain"
)

var (
	defaultTTL     uint32 = 60
	defaultNetwork        = "udp"
)

// Server is an interface of DNS server.
type Server interface {
	Serve(context.Context) error
}

// Config is a configuration object concerning in the DNS server.
type Config struct {
	Port domain.Port
}

func (c *Config) addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// NewServer creates a DNS server instance.
func NewServer(mappingRepo domain.MappingRepository, cfg *Config) Server {
	return &server{
		Config:      cfg,
		mappingRepo: mappingRepo,
		log:         zap.L().Named("dns"),
	}
}

type server struct {
	*Config
	mappingRepo domain.MappingRepository
	server      *godns.Server
	log         *zap.Logger
}

func (s *server) Serve(ctx context.Context) error {
	s.server = &godns.Server{
		Handler: godns.HandlerFunc(s.handle),
		Addr:    s.addr(),
		Net:     defaultNetwork,
	}

	var err error
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("starting DNS server...", zap.String("addr", s.addr()))
		errCh <- errors.WithStack(s.server.ListenAndServe())
	}()

	select {
	case err = <-errCh:
		err = errors.WithStack(err)
	case <-ctx.Done():
		s.log.Info("shutdowning DNS server...", zap.Error(ctx.Err()))
		s.server.Shutdown()
		err = errors.WithStack(<-errCh)
	}

	return errors.WithStack(err)
}

func (s *server) handle(w godns.ResponseWriter, req *godns.Msg) {
	q := req.Question[0]
	resp := new(godns.Msg)
	resp.SetReply(req)

	if ip, ok := s.lookup(q); ok {
		resp.Answer = append(resp.Answer, &godns.A{
			Hdr: godns.RR_Header{
				Name:   q.Name,
				Rrtype: godns.TypeA,
				Class:  godns.ClassINET,
				Ttl:    defaultTTL,
			},
			A: ip,
		})
	} else {
		resp.MsgHdr.Rcode = godns.RcodeNameError
	}

	s.log.Debug("received message", zap.Any("req", q), zap.Any("resp", resp))

	w.WriteMsg(resp)
}

func (s *server) lookup(q godns.Question) (ip net.IP, ok bool) {
	if q.Qtype == godns.TypeA && q.Qclass == godns.ClassINET {
		ip, ok = s.mappingRepo.LookupIP(context.TODO(), strings.TrimSuffix(q.Name, "."))
	}
	return
}
