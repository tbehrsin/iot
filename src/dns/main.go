package main

import (
	"db"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"strings"
)

const Domain = db.Domain
const WebAddress1 = "52.56.139.203"
const WebAddress2 = "35.177.149.9"
const NS1Address = "52.56.139.203"
const NS2Address = "35.177.149.9"

func alias(name string, cname string, res *dns.Msg) {
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(cname, dns.TypeA)
	m.RecursionDesired = true
	r, _, err := c.Exchange(m, config.Servers[0]+":"+config.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		return
	}
	for _, a := range r.Answer {
		if rr, ok := a.(*dns.A); ok {
			if out, err := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", name, rr.A.String())); err == nil {
				res.Answer = append(res.Answer, out)
			}
		}
	}
}

func server(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	m.Authoritative = true

	if rr, err := dns.NewRR(fmt.Sprintf("%s 1209600 IN NS ns1.%s", Domain, Domain)); err == nil {
		m.Ns = append(m.Ns, rr)
	}
	if rr, err := dns.NewRR(fmt.Sprintf("%s 1209600 IN NS ns2.%s", Domain, Domain)); err == nil {
		m.Ns = append(m.Ns, rr)
	}

	if rr, err := dns.NewRR(fmt.Sprintf("ns1.%s 300 IN A %s", Domain, NS1Address)); err == nil {
		m.Extra = append(m.Extra, rr)
	}
	if rr, err := dns.NewRR(fmt.Sprintf("ns2.%s 300 IN A %s", Domain, NS2Address)); err == nil {
		m.Extra = append(m.Extra, rr)
	}

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			name := strings.ToLower(q.Name)
			if q.Qtype == dns.TypeSOA && name == Domain {
				if rr, err := dns.NewRR(fmt.Sprintf("%s 900 IN SOA ns1.%s awsdns-hostmaster.amazon.com. 2018102901 900 900 1800 60", name, Domain)); err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else if q.Qtype == dns.TypeNS && name == Domain {
				if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN NS ns1.%s", name, Domain)); err == nil {
					m.Answer = append(m.Answer, rr)
				}
				if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN NS ns2.%s", name, Domain)); err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else if (q.Qtype == dns.TypeA || q.Qtype == dns.TypeCNAME) && name == Domain {
				if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", name, WebAddress1)); err == nil {
					m.Answer = append(m.Answer, rr)
				}
				if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", name, WebAddress2)); err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else if q.Qtype == dns.TypeA || q.Qtype == dns.TypeCNAME {
				if name == "ns1.z3js.net." {
					if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", name, NS1Address)); err == nil {
						m.Answer = append(m.Answer, rr)
					}
					continue
				}

				if name == "ns2.z3js.net." {
					if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", name, NS2Address)); err == nil {
						m.Answer = append(m.Answer, rr)
					}
					continue
				}

				id := strings.TrimSuffix(name, fmt.Sprintf(".%s", Domain))
				if gw, _ := db.GetGateway(id); gw != nil {
					if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", name, gw.Address)); err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		}
	}

	w.WriteMsg(m)
}

func main() {
	dns.HandleFunc(Domain, server)
	go func() {
		log.Fatal(dns.ListenAndServe(":53", "udp", nil))
	}()
	log.Fatal(dns.ListenAndServe(":53", "tcp", nil))
}
