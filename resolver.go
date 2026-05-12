package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

type ResponseBlock struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   string `json:"ttl"`
}

func string_to_type(s string) uint16 {

	switch s {
	case "A":
		return dns.TypeA
	case "AAAA":
		return dns.TypeAAAA
	case "TXT":
		return dns.TypeTXT
	case "CNAME":
		return dns.TypeCNAME
	case "NS":
		return dns.TypeNS
	case "MX":
		return dns.TypeMX
	case "SOA":
		return dns.TypeSOA
	case "SRV":
		return dns.TypeSRV
	case "CAA":
		return dns.TypeCAA
	case "PTR":
		return dns.TypePTR
	case "HINFO":
		return dns.TypeHINFO
	case "CERT":
		return dns.TypeCERT
	case "DNSKEY":
		return dns.TypeDNSKEY
	case "DS":
		return dns.TypeDS
	case "HTTPS":
		return dns.TypeHTTPS
	case "LOC":
		return dns.TypeLOC
	case "NAPTR":
		return dns.TypeNAPTR
	default:
		return 0
	}

}

func result_to_json(answer []dns.RR, _type string) ([]byte, error) {

	var result []ResponseBlock

	for _, ans := range answer {
		var value string

		varray := strings.Split(ans.String(), "\t")
		if len(varray) != 5 {
			continue
		}
		value = varray[4]

		block := ResponseBlock{Type: _type, Value: value, TTL: varray[1]}

		result = append(result, block)
	}

	return json.Marshal(result)
}

func dns_to_url(s string) string {
	switch s {
	case "GoogleDNS":
		return "https://8.8.8.8/dns-query"
	case "YandexDNS":
		return "https://77.88.8.8/dns-query"
	case "CloudflareDNS":
		return "https://1.1.1.1/dns-query"
	case "Quad9DNS":
		return "https://9.9.9.9/dns-query"
	case "OpenDNS":
		return "https://208.67.222.222/dns-query"
	default:
		return "https://8.8.8.8/dns-query"
	}
}

func resolve(domain string, _type string, dns_server string) ([]byte, error) {
	m := new(dns.Msg)
	m.SetQuestion(fmt.Sprintf("%s.", domain), string_to_type(_type))

	msg, err := m.Pack()

	if err != nil {
		fmt.Println(err)
	}

	buf := bytes.NewBuffer(msg)

	resp, e := http.Post(dns_to_url(dns_server), "application/dns-message", buf)

	if e != nil {
		fmt.Println(e)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	m1 := new(dns.Msg)
	m1.Unpack(body)

	return result_to_json(m1.Answer, _type)

}
