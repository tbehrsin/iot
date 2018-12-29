package main

import (
	"db"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
)

const Domain = db.Domain
const NS1 = "iot-ns1.behrsin.com."
const NS2 = "iot-ns2.behrsin.com."

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
			if out, err := dns.NewRR(fmt.Sprintf("%s 30 IN A %s", name, rr.A.String())); err == nil {
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

	if rr, err := dns.NewRR(fmt.Sprintf("%s 1209600 IN NS %s", Domain, NS1)); err == nil {
		m.Ns = append(m.Ns, rr)
	}
	if rr, err := dns.NewRR(fmt.Sprintf("%s 1209600 IN NS %s", Domain, NS2)); err == nil {
		m.Ns = append(m.Ns, rr)
	}

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			name := strings.ToLower(q.Name)
			if q.Qtype == dns.TypeSOA && name == Domain {
				if rr, err := dns.NewRR(fmt.Sprintf("%s 900 IN SOA %s hostmaster.iot.behrsin.com. 2018102901 900 900 1800 60", name, NS1)); err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else if q.Qtype == dns.TypeNS && name == Domain {
				if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN NS %s", name, NS1)); err == nil {
					m.Answer = append(m.Answer, rr)
				}
				if rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN NS %s", name, NS2)); err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else if q.Qtype == dns.TypeA && name == Domain {
				alias(name, NS1, m)
				alias(name, NS2, m)
			} else if q.Qtype == dns.TypeA {
				id := strings.TrimSuffix(name, fmt.Sprintf(".%s", Domain))

				if id == "ca" || id == "proxy" {
					alias(name, NS1, m)
					alias(name, NS2, m)
				} else if strings.HasPrefix(id, "local.") {
					if gw, _ := db.GetGateway(strings.TrimPrefix(id, "local.")); gw != nil {
						if rr, err := dns.NewRR(fmt.Sprintf("%s 10 IN A %s", name, gw.LocalAddress)); err == nil {
							m.Answer = append(m.Answer, rr)
						}
					}
				} else if gw, _ := db.GetGateway(id); gw != nil {
					if rr, err := dns.NewRR(fmt.Sprintf("%s 30 IN A %s", name, gw.Address)); err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		}
	}

	w.WriteMsg(m)
}

func main() {
	if err := db.Initialize(); err != nil {
		panic(err)
	}

	port := "53"

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	dns.HandleFunc(Domain, server)
	go func() {
		log.Fatal(dns.ListenAndServe(fmt.Sprintf(":%s", port), "udp", nil))
	}()
	log.Fatal(dns.ListenAndServe(fmt.Sprintf(":%s", port), "tcp", nil))
}
