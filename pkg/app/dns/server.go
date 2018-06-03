package dns

import (
	"context"
	"net"
	"strings"

	godns "github.com/miekg/dns"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
)

var (
	defaultTTL     uint32 = 60
	defaultNetwork        = "udp"
	defaultAddr           = ":53"
)

// NewServer creates a DNS server instance.
func NewServer(mapper domain.Mapper, localhost net.IP) app.Server {
	return &server{
		mapper:    mapper,
		localhost: localhost,
		addr:      defaultAddr,
		log:       zap.L().Named("dns"),
	}
}

type server struct {
	mapper    domain.Mapper
	server    *godns.Server
	localhost net.IP
	addr      string
	log       *zap.Logger
}

func (s *server) Serve(ctx context.Context) error {
	s.server = &godns.Server{
		Handler: godns.HandlerFunc(s.handle),
		Addr:    s.addr,
		Net:     defaultNetwork,
	}

	var err error
	errCh := make(chan error, 1)
	go func() {
		s.log.Debug("starting DNS server...", zap.String("addr", s.addr))
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err = <-errCh:
		// do nothing
	case <-ctx.Done():
		s.log.Debug("shutdowning DNS server...", zap.Error(ctx.Err()))
		s.server.Shutdown()
		err = <-errCh
	}

	return err
}

func (s *server) Addr() string {
	return s.server.Addr
}

func (s *server) handle(w godns.ResponseWriter, req *godns.Msg) {
	q := req.Question[0]
	resp := new(godns.Msg)
	resp.SetReply(req)

	if s.handlable(q) {
		resp.Answer = append(resp.Answer, &godns.A{
			Hdr: godns.RR_Header{
				Name:   q.Name,
				Rrtype: godns.TypeA,
				Class:  godns.ClassINET,
				Ttl:    defaultTTL,
			},
			A: s.localhost,
		})
	} else {
		resp.MsgHdr.Rcode = godns.RcodeNameError
	}

	w.WriteMsg(resp)
}

func (s *server) handlable(q godns.Question) (ok bool) {
	if q.Qtype == godns.TypeA && q.Qclass == godns.ClassINET {
		_, ok = s.mapper.Lookup(strings.TrimSuffix(q.Name, "."))
	}
	return
}
