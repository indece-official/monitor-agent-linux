package utils

import (
	"fmt"

	"github.com/miekg/dns"
)

func ResolveDNS(host string, dnsServer string) (string, error) {
	c := dns.Client{}
	m := dns.Msg{}

	m.SetQuestion(host+".", dns.TypeA)

	r, _, err := c.Exchange(&m, dnsServer)
	if err != nil {
		return "", fmt.Errorf("can't resolve '%s' on %s: %s", host, dnsServer, err)
	}

	if len(r.Answer) == 0 {
		return "", fmt.Errorf("can't resolve '%s' on %s: No results", host, dnsServer)
	}

	aRecord := r.Answer[0].(*dns.A)

	return aRecord.A.String(), nil
}
