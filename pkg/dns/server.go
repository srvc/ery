package dns

import (
	"context"
	"net"
	"strings"

	godns "github.com/miekg/dns"

	"github.com/srvc/ery/pkg/discovery"
)

var (
	defaultTTL     uint32 = 60
	defaultNetwork        = "udp"
)

// NewServer creates a DNS server instance.
func NewServer(mapper discovery.Mapper, localhost net.IP, addr string) discovery.Server {
	return &server{
		mapper:    mapper,
		localhost: localhost,
		addr:      addr,
	}
}

type server struct {
	mapper    discovery.Mapper
	server    *godns.Server
	localhost net.IP
	addr      string
}

func (s *server) Serve() error {
	s.server = &godns.Server{
		Handler: godns.HandlerFunc(s.handle),
		Addr:    s.addr,
		Net:     defaultNetwork,
	}
	return s.server.ListenAndServe()
}

func (s *server) Shutdown(context.Context) error {
	return s.server.Shutdown()
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
