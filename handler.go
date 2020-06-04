package coredns_etcd_backend

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	// "fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
)

// https://golangcookbook.com/chapters/arrays/reverse/
func reverse(x []string) []string {
	for i := 0; i < len(x)/2; i++ {
		j := len(x) - i - 1
		x[i], x[j] = x[j], x[i]
	}
	return x
}

// https://github.com/arvancloud/redis/blob/8a6e3c44c0ac1ae258addbe741365c2b69992350/redis.go
func split255(s string) []string {
	if len(s) < 255 {
		return []string{s}
	}
	sx := []string{}
	p, i := 0, 255
	for {
		if i <= len(s) {
			sx = append(sx, s[p:i])
		} else {
			sx = append(sx, s[p:])
			break

		}
		p, i = p+255, i+255
	}

	return sx
}

type EtcdService struct {
	Host     string `json:"host"`
	Text     string `json:"text"`
	Cname    string `json:"cname"`
	Target   string `json:"target"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Priority uint16 `json:"priority"`
}

// ServeDNS implements the plugin.Handler interface.
func (e *Etcd) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	//fmt.Println(r)
	state := request.Request{W: w, Req: r}

	qname := state.Name()
	qtype := state.Type()
	fmt.Println(qname)
	fmt.Println(qtype)

	zone := plugin.Zones(e.Zones).Matches(qname)

	fmt.Println("zone : ", zone)

	if zone == "" {
		return plugin.NextOrFailure(qname, e.Next, ctx, w, r)
	}

	// z := redis.load(zone)
	// if z == nil {
	// 	return redis.errorResponse(state, zone, dns.RcodeServerFailure, nil)
	// }

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	// record := redis.get(location, z)

	//fmt.Println(r)
	splitDomain := strings.Split(qname, ".")
	reverseSplitDomain := reverse(splitDomain)
	reversedDomain := strings.Join(reverseSplitDomain, "/")
	etcdPath := defaultPrefix + reversedDomain + "/-" + qtype + "-"
	fmt.Println(etcdPath)

	rCode := dns.RcodeSuccess

	switch qtype {
	case "A":
		resp, err := e.Client.Get(ctx, etcdPath, clientv3.WithPrefix())
		if err != nil {
			fmt.Println(err)
			rCode = dns.RcodeServerFailure
			break
		}
		for _, ev := range resp.Kvs {
			var record EtcdService
			err := json.Unmarshal(ev.Value, &record)
			if err != nil {
				fmt.Println(err)
				rCode = dns.RcodeServerFailure
				continue
			}
			r := new(dns.A)
			r.Hdr = dns.RR_Header{Name: dns.Fqdn(qname), Rrtype: dns.TypeA,
				Class: dns.ClassINET, Ttl: 65}
			r.A = net.ParseIP(record.Host)
			answers = append(answers, r)
		}
	case "TXT":
		resp, err := e.Client.Get(ctx, etcdPath, clientv3.WithPrefix())
		if err != nil {
			fmt.Println(err)
			rCode = dns.RcodeServerFailure
			break
		}
		for _, ev := range resp.Kvs {
			var record EtcdService
			err := json.Unmarshal(ev.Value, &record)
			if err != nil {
				fmt.Println(err)
				rCode = dns.RcodeServerFailure
				continue
			}
			r := new(dns.TXT)
			r.Hdr = dns.RR_Header{Name: dns.Fqdn(qname), Rrtype: dns.TypeTXT,
				Class: dns.ClassINET, Ttl: 65}
			r.Txt = split255(record.Text)
			answers = append(answers, r)
		}
	case "CNAME":
		resp, err := e.Client.Get(ctx, etcdPath) //Dont use prefix so that we only get one cname for that record
		if err != nil {
			fmt.Println(err)
			rCode = dns.RcodeServerFailure
			break
		}
		for _, ev := range resp.Kvs {
			var record EtcdService
			err := json.Unmarshal(ev.Value, &record)
			if err != nil {
				fmt.Println(err)
				rCode = dns.RcodeServerFailure
				continue
			}
			r := new(dns.CNAME)
			r.Hdr = dns.RR_Header{Name: dns.Fqdn(qname), Rrtype: dns.TypeCNAME,
				Class: dns.ClassINET, Ttl: 65}
			r.Target = dns.Fqdn(record.Cname)
			answers = append(answers, r)
		}
	case "SRV":
		resp, err := e.Client.Get(ctx, etcdPath, clientv3.WithPrefix())
		if err != nil {
			fmt.Println(err)
			rCode = dns.RcodeServerFailure
			break
		}
		for _, ev := range resp.Kvs {
			var record EtcdService
			err := json.Unmarshal(ev.Value, &record)
			//err = json.NewDecoder(ev.Value).Decode(&record)
			if err != nil {
				fmt.Println(err)
				rCode = dns.RcodeServerFailure
				continue
			}
			r := new(dns.SRV)
			r.Hdr = dns.RR_Header{Name: dns.Fqdn(qname), Rrtype: dns.TypeSRV,
				Class: dns.ClassINET, Ttl: 65}
			r.Target = dns.Fqdn(record.Target)
			r.Weight = record.Weight
			r.Port = record.Port
			r.Priority = record.Priority
			answers = append(answers, r)
		}
	case "SOA":
		r := new(dns.SOA)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(qname), Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: 3600}
		// https://github.com/coredns/coredns/blob/b9b513c48c1edc737fac2106efcc0d4922145a90/plugin/backend_lookup.go#L452
		Ns := "ns.dns."
		Mbox := "hostmaster."
		if zone[0] != '.' {
			Mbox += zone
			Ns += zone
		}
		r.Ns = Ns
		r.Mbox = Mbox
		r.Refresh = 86400
		r.Retry = 7200
		r.Expire = 604800
		r.Minttl = 3600
		r.Serial = uint32(time.Now().Unix())
		answers = append(answers, r)
	default:
		rCode = dns.RcodeNotImplemented
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	state.SizeAndDo(m)
	m = state.Scrub(m)
	_ = w.WriteMsg(m)
	return rCode, nil
}

// Name implements the Handler interface.
func (e *Etcd) Name() string { return "coredns_etcd_backend" }
