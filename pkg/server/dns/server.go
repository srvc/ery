package dns

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/ery/domain"
)

var (
	defaultTTL     uint32 = 60
	defaultNetwork        = "udp"
	defaultPort           = 53
)

// NewServer creates a DNS server instance.
func NewServer(
	appRepo domain.AppRepository,
) *Server {
	return &Server{
		appRepo: appRepo,
		log:     zap.L().Named("dns"),
	}
}

type Server struct {
	appRepo domain.AppRepository
	server  *dns.Server
	log     *zap.Logger
}

func (s *Server) Serve(ctx context.Context) error {
	s.server = &dns.Server{
		Handler: dns.HandlerFunc(s.handle),
		Addr:    fmt.Sprintf(":%d", defaultPort),
		Net:     defaultNetwork,
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	go func() {
		select {
		case <-doneCh:
			// no-op
		case <-ctx.Done():
			s.log.Info("shutdowning DNS server...", zap.Error(ctx.Err()))
			s.server.Shutdown()
		}
	}()

	s.log.Info("starting DNS server...", zap.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Server) handle(w dns.ResponseWriter, req *dns.Msg) {
	ctx := context.Background()

	q := req.Question[0]
	resp := new(dns.Msg)
	resp.SetReply(req)

	if ip, ok := s.lookup(ctx, q); ok {
		resp.Answer = append(resp.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    defaultTTL,
			},
			A: ip,
		})
	} else {
		resp.MsgHdr.Rcode = dns.RcodeNameError
	}

	s.log.Debug("received message", zap.Any("req", q), zap.Any("resp", resp))

	w.WriteMsg(resp)
}

func (s *Server) lookup(ctx context.Context, q dns.Question) (ip net.IP, ok bool) {
	if q.Qtype == dns.TypeA && q.Qclass == dns.ClassINET {
		app, err := s.appRepo.GetByHostname(ctx, strings.TrimSuffix(q.Name, "."))
		if err != nil {
			return
		}
		ip, ok = net.ParseIP(app.GetIp()), true
	}
	return
}
