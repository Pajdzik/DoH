package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/miekg/dns"

	"github.com/zenazn/goji/web"
)

func query(c web.C, w http.ResponseWriter, r *http.Request) {
	dnsMessage, format := getDnsMessage(c.URLParams["hostname"], r)
	response := queryInternal(dnsMessage)
	formattedResponse := convertToFormat(format, response)
	fmt.Fprintf(w, "%s", formattedResponse)
}

func queryInternal(dnsRequest *dns.Msg) *dns.Msg {
	dnsClient := new(dns.Client)

	response, _, err := dnsClient.Exchange(dnsRequest, getDnsServer())
	if err != nil {
		fmt.Print(err)
	}

	return response
}

func getDnsServer() string {
	serversCount := len(config.DnsConfig.TrustedDnsServers)
	randomIndex := rand.Int() % serversCount

	return net.JoinHostPort(config.DnsConfig.TrustedDnsServers[randomIndex], "53")
}

func getDnsMessage(hostname string, r *http.Request) (*dns.Msg, string) {
	urlQueries := parseUrlQuery(r.URL.RawQuery)
	queryType := parseQueryType((*urlQueries)["type"])

	dnsMessage := new(dns.Msg)
	dnsMessage.SetQuestion(dns.Fqdn(hostname), queryType)
	dnsMessage.RecursionDesired, _ = strconv.ParseBool(getQueryParameter(urlQueries, "rr", "true"))
	dnsMessage.CheckingDisabled, _ = strconv.ParseBool(getQueryParameter(urlQueries, "cd", "false"))

	return dnsMessage, (*urlQueries)["format"]
}

func getQueryParameter(urlQueries *map[string]string, parameter string, defaultValue string) string {
	value, exists := (*urlQueries)[parameter]

	if exists {
		return value
	}

	return defaultValue
}

func parseUrlQuery(urlQuery string) *map[string]string {
	queries := make(map[string]string)

	if urlQuery == "" {
		return &queries
	}

	parts := strings.Split(urlQuery, "&")

	for _, part := range parts {
		queryParts := strings.Split(part, "=")
		queries[queryParts[0]] = queryParts[1]
	}

	return &queries
}

func parseQueryType(parameter string) uint16 {
	queryType, err := strconv.Atoi(parameter)
	if err == nil {
		return uint16(queryType)
	}

	switch strings.ToLower(parameter) {
	case "a":
		return dns.TypeA
	case "aaaa":
		return dns.TypeAAAA
	default:
		return dns.TypeA
	}
}

func convertToFormat(format string, dnsMessage *dns.Msg) string {
	switch format {
	case "json":
		return convertToJson(dnsMessage)
	case "raw":
		return fmt.Sprint(dnsMessage)
	default:
		return convertToJson(dnsMessage)
	}
}

func convertToJson(dnsMessage *dns.Msg) string {
	jsonOutput, err := json.Marshal(*dnsMessage)

	if err != nil {
		errorMessage := fmt.Sprintf("Failed to convert %+v to JSON", *dnsMessage)
		log.Error(errorMessage)
		return errorMessage
	}

	return string(jsonOutput)
}
