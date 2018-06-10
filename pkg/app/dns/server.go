package dns

import (
	"context"
	"net"
	"strings"

	godns "github.com/miekg/dns"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/util/netutil"
)

var (
	defaultTTL     uint32 = 60
	defaultNetwork        = "udp"
	defaultAddr           = ":53"
)

// NewServer creates a DNS server instance.
func NewServer(mappingRepo domain.MappingRepository) app.Server {
	return &server{
		mappingRepo: mappingRepo,
		addr:        defaultAddr,
		log:         zap.L().Named("dns"),
	}
}

type server struct {
	mappingRepo domain.MappingRepository
	server      *godns.Server
	addr        string
	log         *zap.Logger
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
		errCh <- errors.WithStack(s.server.ListenAndServe())
	}()

	select {
	case err = <-errCh:
		err = errors.WithStack(err)
	case <-ctx.Done():
		s.log.Debug("shutdowning DNS server...", zap.Error(ctx.Err()))
		s.server.Shutdown()
		err = errors.WithStack(<-errCh)
	}

	return errors.WithStack(err)
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
			A: s.localhost(),
		})
	} else {
		resp.MsgHdr.Rcode = godns.RcodeNameError
	}

	w.WriteMsg(resp)
}

func (s *server) handlable(q godns.Question) (ok bool) {
	if q.Qtype == godns.TypeA && q.Qclass == godns.ClassINET {
		ok, _ = s.mappingRepo.HasHost(context.TODO(), strings.TrimSuffix(q.Name, "."))
	}
	return
}

func (s *server) localhost() net.IP {
	return netutil.LocalIP() // FIXME
}
