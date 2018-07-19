package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/miekg/dns"

	"github.com/zenazn/goji/web"
)

func query(c web.C, w http.ResponseWriter, r *http.Request) {
	urlQueries := parseUrlQuery(r.URL.RawQuery)
	queryType := parseQueryType((*urlQueries)["type"])
	response := queryInternal(c.URLParams["name"], queryType)
	fmt.Fprintf(w, "Result:\n%s!", response)
}

func queryInternal(hostname string, queryType uint16) *dns.Msg {
	dnsClient := new(dns.Client)

	dnsMessage := new(dns.Msg)
	dnsMessage.SetQuestion(dns.Fqdn(hostname), queryType)
	dnsMessage.RecursionDesired = true

	response, _, err := dnsClient.Exchange(dnsMessage, net.JoinHostPort("8.8.8.8", "53"))
	if err != nil {
		fmt.Print(err)
	}

	fmt.Print(response)

	return response
}

func parseUrlQuery(urlQuery string) *map[string]string {
	queries := make(map[string]string)

	parts := strings.Split(urlQuery, "&")

	for _, part := range parts {
		queryParts := strings.Split(part, "=")
		queries[queryParts[0]] = queryParts[1]
	}

	return &queries
}

func parseQueryType(parameter string) uint16 {
	switch strings.ToLower(parameter) {
	case "a":
		return dns.TypeA
	case "aaaa":
		return dns.TypeAAAA
	default:
		return dns.TypeNone
	}
}
