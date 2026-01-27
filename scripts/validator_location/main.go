package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Configuration
var (
	rpcURL  = getEnv("RPC_URL", "https://rpc-fuel-seq.simplystaking.xyz")
	restURL = getEnv("REST_URL", "https://rest-fuel-seq.simplystaking.xyz")
)

// Data structures for Cosmos SDK responses
type ValidatorsResponse struct {
	Validators []Validator `json:"validators"`
}

type Validator struct {
	OperatorAddress string `json:"operator_address"`
	ConsensusPubkey struct {
		Type string `json:"@type"`
		Key  string `json:"key"`
	} `json:"consensus_pubkey"`
	Status      string `json:"status"`
	Tokens      string `json:"tokens"`
	Description struct {
		Moniker         string `json:"moniker"`
		Identity        string `json:"identity"`
		Website         string `json:"website"`
		SecurityContact string `json:"security_contact"`
		Details         string `json:"details"`
	} `json:"description"`
}

type NetInfoResponse struct {
	Result struct {
		Peers []Peer `json:"peers"`
	} `json:"result"`
}

type Peer struct {
	NodeInfo struct {
		ID      string `json:"id"`
		Moniker string `json:"moniker"`
	} `json:"node_info"`
	RemoteIP string `json:"remote_ip"`
}

type ConsensusValidatorsResponse struct {
	Result struct {
		Validators []ConsensusValidator `json:"validators"`
	} `json:"result"`
}

type ConsensusValidator struct {
	Address string `json:"address"`
	PubKey  struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
}

// Geolocation result from a single source
type GeoResult struct {
	Source     string
	Country    string
	Region     string
	City       string
	Lat        float64
	Lon        float64
	ISP        string
	Org        string
	ASN        string
	ASNOrg     string
	Datacenter string
	Success    bool
	Error      string
}

// Aggregated location with confidence
type AggregatedLocation struct {
	Country    string
	Region     string
	City       string
	ISP        string
	ASN        string
	Datacenter string
	Confidence float64
	Sources    []GeoResult
	Consensus  string // Summary of agreement between sources
}

// Validator with location info
type ValidatorLocation struct {
	Rank            int
	OperatorAddress string
	Moniker         string
	Tokens          int64
	VotingPowerPct  float64
	Website         string
	PeerIP          string
	ResolvedIPs     []string
	Location        AggregatedLocation
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func main() {
	useTraceroute := flag.Bool("traceroute", true, "Use traceroute for additional verification (disable with --traceroute=false)")
	flag.BoolVar(useTraceroute, "t", true, "Use traceroute (shorthand)")
	outputJSON := flag.Bool("json", false, "Output results as JSON")
	flag.BoolVar(outputJSON, "j", false, "Output JSON (shorthand)")
	pct := flag.Float64("percentage", 66.7, "Voting power percentage threshold (e.g. 66.7 for top 2/3)")
	flag.Float64Var(pct, "p", 66.7, "Voting power percentage (shorthand)")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Fetch all required data
	fmt.Fprintln(os.Stderr, "Fetching validators from REST API...")
	validators, err := fetchValidators(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching validators: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Fetching peer information from RPC...")
	peers, err := fetchNetInfo(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not fetch peer info: %v\n", err)
		peers = []Peer{}
	}

	fmt.Fprintln(os.Stderr, "Fetching consensus validators...")
	consensusVals, err := fetchConsensusValidators(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not fetch consensus validators: %v\n", err)
		consensusVals = []ConsensusValidator{}
	}

	// Filter bonded validators and sort by voting power
	bondedValidators := filterAndSortValidators(validators)
	if len(bondedValidators) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No bonded validators found")
		os.Exit(1)
	}

	// Calculate total voting power and threshold
	totalPower := calculateTotalPower(bondedValidators)
	threshold := int64(float64(totalPower) * (*pct / 100.0))

	fmt.Fprintf(os.Stderr, "Total voting power: %d\n", totalPower)
	fmt.Fprintf(os.Stderr, "%.1f%% threshold: %d\n\n", *pct, threshold)

	// Get top validators up to threshold
	topValidators := getTopValidators(bondedValidators, totalPower, threshold)

	// Create peer lookup maps
	peerByMoniker := createPeerMonikerMap(peers)
	peerByNodeID := createPeerNodeIDMap(peers)
	consensusByPubkey := createConsensusPubkeyMap(consensusVals)

	fmt.Fprintf(os.Stderr, "Processing %d validators...\n\n", len(topValidators))

	// Process validators concurrently but with rate limiting
	results := processValidatorsWithLocation(ctx, topValidators, totalPower, peerByMoniker, peerByNodeID, consensusByPubkey, *useTraceroute)

	// Output results
	if *outputJSON {
		outputJSONResults(results, totalPower, threshold, *pct)
	} else {
		outputTextResults(results, totalPower, threshold, *pct)
	}
}

func fetchValidators(ctx context.Context) ([]Validator, error) {
	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?pagination.limit=500", restURL)
	resp, err := httpGetWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ValidatorsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Validators, nil
}

func fetchNetInfo(ctx context.Context) ([]Peer, error) {
	url := fmt.Sprintf("%s/net_info", rpcURL)
	resp, err := httpGetWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result NetInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Result.Peers, nil
}

func fetchConsensusValidators(ctx context.Context) ([]ConsensusValidator, error) {
	url := fmt.Sprintf("%s/validators?per_page=100", rpcURL)
	resp, err := httpGetWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ConsensusValidatorsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Result.Validators, nil
}

func httpGetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func filterAndSortValidators(validators []Validator) []Validator {
	var bonded []Validator
	for _, v := range validators {
		if v.Status == "BOND_STATUS_BONDED" {
			bonded = append(bonded, v)
		}
	}

	sort.Slice(bonded, func(i, j int) bool {
		ti := parseTokens(bonded[i].Tokens)
		tj := parseTokens(bonded[j].Tokens)
		return ti > tj
	})

	return bonded
}

func parseTokens(s string) int64 {
	var tokens int64
	fmt.Sscanf(s, "%d", &tokens)
	return tokens
}

func calculateTotalPower(validators []Validator) int64 {
	var total int64
	for _, v := range validators {
		total += parseTokens(v.Tokens)
	}
	return total
}

func getTopValidators(validators []Validator, totalPower, threshold int64) []Validator {
	var result []Validator
	var cumulative int64

	for _, v := range validators {
		if cumulative >= threshold {
			break
		}
		result = append(result, v)
		cumulative += parseTokens(v.Tokens)
	}

	return result
}

func createPeerMonikerMap(peers []Peer) map[string]Peer {
	m := make(map[string]Peer)
	for _, p := range peers {
		key := strings.ToLower(strings.TrimSpace(p.NodeInfo.Moniker))
		m[key] = p
	}
	return m
}

func createPeerNodeIDMap(peers []Peer) map[string]Peer {
	m := make(map[string]Peer)
	for _, p := range peers {
		m[p.NodeInfo.ID] = p
	}
	return m
}

func createConsensusPubkeyMap(validators []ConsensusValidator) map[string]ConsensusValidator {
	m := make(map[string]ConsensusValidator)
	for _, v := range validators {
		m[v.PubKey.Value] = v
	}
	return m
}

func processValidatorsWithLocation(ctx context.Context, validators []Validator, totalPower int64,
	peerByMoniker map[string]Peer, peerByNodeID map[string]Peer,
	consensusByPubkey map[string]ConsensusValidator, useTraceroute bool) []ValidatorLocation {

	results := make([]ValidatorLocation, len(validators))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3) // Limit concurrent requests to avoid rate limiting

	for i, v := range validators {
		wg.Add(1)
		go func(idx int, val Validator) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			tokens := parseTokens(val.Tokens)
			pct := float64(tokens) / float64(totalPower) * 100

			result := ValidatorLocation{
				Rank:            idx + 1,
				OperatorAddress: val.OperatorAddress,
				Moniker:         val.Description.Moniker,
				Tokens:          tokens,
				VotingPowerPct:  pct,
				Website:         val.Description.Website,
			}

			// Try to find peer IP
			ip := findPeerIP(val, peerByMoniker, peerByNodeID, consensusByPubkey)
			result.PeerIP = ip

			// Resolve additional IPs from website if available
			if val.Description.Website != "" {
				resolved := resolveWebsiteIPs(val.Description.Website)
				result.ResolvedIPs = resolved
			}

			// Aggregate location from multiple sources
			if ip != "" {
				fmt.Fprintf(os.Stderr, "[%d/%d] Looking up location for %s (IP: %s)...\n",
					idx+1, len(validators), val.Description.Moniker, ip)
				result.Location = aggregateLocation(ctx, ip, useTraceroute)
			} else if len(result.ResolvedIPs) > 0 {
				fmt.Fprintf(os.Stderr, "[%d/%d] Looking up location for %s (Website IP: %s)...\n",
					idx+1, len(validators), val.Description.Moniker, result.ResolvedIPs[0])
				result.Location = aggregateLocation(ctx, result.ResolvedIPs[0], useTraceroute)
			} else {
				fmt.Fprintf(os.Stderr, "[%d/%d] No IP found for %s\n",
					idx+1, len(validators), val.Description.Moniker)
				result.Location = AggregatedLocation{
					Confidence: 0,
					Consensus:  "No IP address found",
				}
			}

			results[idx] = result
		}(i, v)
	}

	wg.Wait()
	return results
}

func findPeerIP(v Validator, peerByMoniker map[string]Peer, peerByNodeID map[string]Peer,
	consensusByPubkey map[string]ConsensusValidator) string {

	moniker := strings.ToLower(strings.TrimSpace(v.Description.Moniker))

	// Try exact moniker match
	if peer, ok := peerByMoniker[moniker]; ok {
		return peer.RemoteIP
	}

	// Try first word of moniker
	words := strings.Fields(moniker)
	if len(words) > 0 {
		if peer, ok := peerByMoniker[words[0]]; ok {
			return peer.RemoteIP
		}
	}

	// Try prefix matching
	for key, peer := range peerByMoniker {
		if strings.HasPrefix(key, moniker) || strings.HasPrefix(moniker, key) {
			return peer.RemoteIP
		}
	}

	// Try matching via consensus pubkey
	if v.ConsensusPubkey.Key != "" {
		// The key from REST API might be in a different format
		// Try to normalize and match
		for pubkey := range consensusByPubkey {
			if pubkey == v.ConsensusPubkey.Key {
				// Found match, but we'd need to map consensus address to node ID
				// This is complex as Tendermint uses different address formats
				break
			}
		}
	}

	return ""
}

func resolveWebsiteIPs(website string) []string {
	// Extract domain from website URL
	domain := website
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")
	parts := strings.Split(domain, "/")
	if len(parts) > 0 {
		domain = parts[0]
	}

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil
	}

	var result []string
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			result = append(result, ipv4.String())
		}
	}
	return result
}

func aggregateLocation(ctx context.Context, ip string, useTraceroute bool) AggregatedLocation {
	var sources []GeoResult
	var wg sync.WaitGroup

	// Query multiple geolocation services concurrently
	geoFuncs := []func(context.Context, string) GeoResult{
		queryIPAPI,
		queryIPInfo,
		queryFreeIPAPI,
		queryIPWhoIs,
		queryIPWhois,
		queryReverseDNS,
	}

	resultsChan := make(chan GeoResult, len(geoFuncs)+1)

	for _, fn := range geoFuncs {
		wg.Add(1)
		go func(f func(context.Context, string) GeoResult) {
			defer wg.Done()
			resultsChan <- f(ctx, ip)
			time.Sleep(500 * time.Millisecond) // Rate limiting
		}(fn)
	}

	// Optional traceroute
	if useTraceroute {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resultsChan <- tracerouteLocation(ip)
		}()
	}

	wg.Wait()
	close(resultsChan)

	for r := range resultsChan {
		sources = append(sources, r)
	}

	return calculateConsensus(sources)
}

// ip-api.com - Free, 45 req/min
func queryIPAPI(ctx context.Context, ip string) GeoResult {
	result := GeoResult{Source: "ip-api.com"}

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,regionName,city,lat,lon,isp,org,as,hosting", ip)
	resp, err := httpGetWithContext(ctx, url)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	var data struct {
		Status     string  `json:"status"`
		Message    string  `json:"message"`
		Country    string  `json:"country"`
		RegionName string  `json:"regionName"`
		City       string  `json:"city"`
		Lat        float64 `json:"lat"`
		Lon        float64 `json:"lon"`
		ISP        string  `json:"isp"`
		Org        string  `json:"org"`
		AS         string  `json:"as"`
		Hosting    bool    `json:"hosting"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		result.Error = err.Error()
		return result
	}

	if data.Status != "success" {
		result.Error = data.Message
		return result
	}

	result.Success = true
	result.Country = data.Country
	result.Region = data.RegionName
	result.City = data.City
	result.Lat = data.Lat
	result.Lon = data.Lon
	result.ISP = data.ISP
	result.Org = data.Org
	result.ASN = data.AS
	if data.Hosting {
		result.Datacenter = "Yes (hosting/datacenter IP)"
	}

	return result
}

// ipinfo.io - Free tier available
func queryIPInfo(ctx context.Context, ip string) GeoResult {
	result := GeoResult{Source: "ipinfo.io"}

	url := fmt.Sprintf("https://ipinfo.io/%s/json", ip)
	resp, err := httpGetWithContext(ctx, url)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	var data struct {
		City    string `json:"city"`
		Region  string `json:"region"`
		Country string `json:"country"`
		Loc     string `json:"loc"`
		Org     string `json:"org"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		result.Error = err.Error()
		return result
	}

	result.Success = true
	result.City = data.City
	result.Region = data.Region
	result.Country = data.Country
	result.Org = data.Org

	// Parse lat/lon from "loc" field (format: "lat,lon")
	if data.Loc != "" {
		fmt.Sscanf(data.Loc, "%f,%f", &result.Lat, &result.Lon)
	}

	// Parse ASN from org field if present (format: "AS12345 Organization Name")
	if strings.HasPrefix(data.Org, "AS") {
		parts := strings.SplitN(data.Org, " ", 2)
		result.ASN = parts[0]
		if len(parts) > 1 {
			result.ASNOrg = parts[1]
		}
	}

	return result
}

// freeipapi.com - Free, no API key required, generous rate limits
func queryFreeIPAPI(ctx context.Context, ip string) GeoResult {
	result := GeoResult{Source: "freeipapi.com"}

	url := fmt.Sprintf("https://freeipapi.com/api/json/%s", ip)
	resp, err := httpGetWithContext(ctx, url)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	var data struct {
		CountryName string  `json:"countryName"`
		RegionName  string  `json:"regionName"`
		CityName    string  `json:"cityName"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		IsProxy     bool    `json:"isProxy"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		result.Error = err.Error()
		return result
	}

	if data.CountryName != "" {
		result.Success = true
		result.Country = data.CountryName
		result.Region = data.RegionName
		result.City = data.CityName
		result.Lat = data.Latitude
		result.Lon = data.Longitude
	}

	return result
}

// ipwho.is - Free, no API key required
func queryIPWhoIs(ctx context.Context, ip string) GeoResult {
	result := GeoResult{Source: "ipwho.is"}

	url := fmt.Sprintf("https://ipwho.is/%s", ip)
	resp, err := httpGetWithContext(ctx, url)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	var data struct {
		Success    bool    `json:"success"`
		Country    string  `json:"country"`
		Region     string  `json:"region"`
		City       string  `json:"city"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		Connection struct {
			ASN int    `json:"asn"`
			Org string `json:"org"`
			ISP string `json:"isp"`
		} `json:"connection"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		result.Error = err.Error()
		return result
	}

	if data.Success {
		result.Success = true
		result.Country = data.Country
		result.Region = data.Region
		result.City = data.City
		result.Lat = data.Latitude
		result.Lon = data.Longitude
		result.ISP = data.Connection.ISP
		result.Org = data.Connection.Org
		if data.Connection.ASN != 0 {
			result.ASN = fmt.Sprintf("AS%d", data.Connection.ASN)
		}
	}

	return result
}

// WHOIS-based lookup using whois command
func queryIPWhois(ctx context.Context, ip string) GeoResult {
	result := GeoResult{Source: "whois"}

	// Check if whois command is available
	whoisPath, err := exec.LookPath("whois")
	if err != nil {
		result.Error = "whois command not available"
		return result
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, whoisPath, ip)
	output, err := cmd.Output()
	if err != nil {
		result.Error = err.Error()
		return result
	}

	whoisData := string(output)

	// Parse WHOIS output for relevant fields
	result.Success = true

	// Extract country
	countryRegex := regexp.MustCompile(`(?i)country:\s*(\S+)`)
	if matches := countryRegex.FindStringSubmatch(whoisData); len(matches) > 1 {
		result.Country = strings.ToUpper(matches[1])
	}

	// Extract organization/netname
	orgRegex := regexp.MustCompile(`(?i)(?:org-?name|organization|netname):\s*(.+)`)
	if matches := orgRegex.FindStringSubmatch(whoisData); len(matches) > 1 {
		result.Org = strings.TrimSpace(matches[1])
	}

	// Extract ASN
	asnRegex := regexp.MustCompile(`(?i)origin(?:as)?:\s*(AS?\d+)`)
	if matches := asnRegex.FindStringSubmatch(whoisData); len(matches) > 1 {
		result.ASN = strings.ToUpper(matches[1])
		if !strings.HasPrefix(result.ASN, "AS") {
			result.ASN = "AS" + result.ASN
		}
	}

	// Extract description for ASN org name
	descRegex := regexp.MustCompile(`(?i)descr:\s*(.+)`)
	if matches := descRegex.FindStringSubmatch(whoisData); len(matches) > 1 {
		result.ASNOrg = strings.TrimSpace(matches[1])
	}

	return result
}

// Traceroute-based location estimation
func tracerouteLocation(ip string) GeoResult {
	result := GeoResult{Source: "traceroute"}

	// Check if traceroute is available
	traceroutePath, err := exec.LookPath("traceroute")
	if err != nil {
		// Try mtr as alternative
		traceroutePath, err = exec.LookPath("mtr")
		if err != nil {
			result.Error = "traceroute/mtr not available"
			return result
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if strings.Contains(traceroutePath, "mtr") {
		cmd = exec.CommandContext(ctx, traceroutePath, "-r", "-c", "1", "-n", ip)
	} else {
		cmd = exec.CommandContext(ctx, traceroutePath, "-m", "15", "-q", "1", "-w", "2", ip)
	}

	output, err := cmd.Output()
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Extract IPs from traceroute output
	ipRegex := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
	matches := ipRegex.FindAllString(string(output), -1)

	// Only use traceroute if the target IP itself appears in the output,
	// meaning the route actually reached the destination. Otherwise we'd
	// just be geolocating a random intermediate hop (often near the caller).
	reached := false
	for _, m := range matches {
		if m == ip {
			reached = true
			break
		}
	}

	if reached {
		// Target was reached — look up the hop just before it for network info
		hopIP := ""
		for i, m := range matches {
			if m == ip && i > 0 {
				prev := matches[i-1]
				if prev != "0.0.0.0" {
					hopIP = prev
				}
			}
		}
		if hopIP != "" {
			hopResult := queryIPAPI(ctx, hopIP)
			if hopResult.Success {
				result.Success = true
				result.Country = hopResult.Country
				result.Region = hopResult.Region
				result.City = hopResult.City
				result.ISP = hopResult.ISP
			}
		}
		// If no usable penultimate hop, geolocate the target directly
		if !result.Success {
			hopResult := queryIPAPI(ctx, ip)
			if hopResult.Success {
				result.Success = true
				result.Country = hopResult.Country
				result.Region = hopResult.Region
				result.City = hopResult.City
				result.ISP = hopResult.ISP
			}
		}
	} else {
		result.Error = "traceroute did not reach destination"
	}

	return result
}

// queryReverseDNS performs a PTR lookup and parses the hostname for location
// hints. Hosting providers frequently embed city/airport codes in rDNS names
// (e.g. "fsn1-speed.hetzner.com" → Falkenstein, "nbg1-dc3.hetzner.com" → Nuremberg).
func queryReverseDNS(ctx context.Context, ip string) GeoResult {
	result := GeoResult{Source: "rdns"}

	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		result.Error = "no PTR record"
		return result
	}

	hostname := strings.ToLower(strings.TrimSuffix(names[0], "."))

	// Known location codes found in datacenter rDNS hostnames.
	// Maps substring → (city, region, country).
	type locHint struct {
		City, Region, Country string
	}
	hints := []struct {
		pattern string
		loc     locHint
	}{
		// Hetzner
		{"fsn", locHint{"Falkenstein", "Saxony", "Germany"}},
		{"nbg", locHint{"Nuremberg", "Bavaria", "Germany"}},
		{"hel", locHint{"Helsinki", "Uusimaa", "Finland"}},
		// OVH
		{"gra", locHint{"Gravelines", "Hauts-de-France", "France"}},
		{"sbg", locHint{"Strasbourg", "Grand Est", "France"}},
		{"bhs", locHint{"Beauharnois", "Quebec", "Canada"}},
		{"rbx", locHint{"Roubaix", "Hauts-de-France", "France"}},
		// AWS region codes
		{"us-east-1", locHint{"Ashburn", "Virginia", "United States"}},
		{"us-east-2", locHint{"Columbus", "Ohio", "United States"}},
		{"us-west-1", locHint{"San Jose", "California", "United States"}},
		{"us-west-2", locHint{"Portland", "Oregon", "United States"}},
		{"eu-west-1", locHint{"Dublin", "Leinster", "Ireland"}},
		{"eu-west-2", locHint{"London", "England", "United Kingdom"}},
		{"eu-central-1", locHint{"Frankfurt", "Hesse", "Germany"}},
		{"ap-southeast-1", locHint{"Singapore", "", "Singapore"}},
		{"ap-northeast-1", locHint{"Tokyo", "", "Japan"}},
		// Common airport/city codes in hostnames
		{"fra", locHint{"Frankfurt", "Hesse", "Germany"}},
		{"ams", locHint{"Amsterdam", "North Holland", "Netherlands"}},
		{"lhr", locHint{"London", "England", "United Kingdom"}},
		{"cdg", locHint{"Paris", "Ile-de-France", "France"}},
		{"sjc", locHint{"San Jose", "California", "United States"}},
		{"iad", locHint{"Ashburn", "Virginia", "United States"}},
		{"dfw", locHint{"Dallas", "Texas", "United States"}},
		{"ord", locHint{"Chicago", "Illinois", "United States"}},
		{"lax", locHint{"Los Angeles", "California", "United States"}},
		{"nrt", locHint{"Tokyo", "", "Japan"}},
		{"sin", locHint{"Singapore", "", "Singapore"}},
		{"syd", locHint{"Sydney", "New South Wales", "Australia"}},
		{"gru", locHint{"São Paulo", "São Paulo", "Brazil"}},
		{"waw", locHint{"Warsaw", "Masovian", "Poland"}},
	}

	// Split hostname into parts for matching (e.g. "fsn1-dc3.hetzner.com" → ["fsn1-dc3", "hetzner", "com"])
	parts := strings.FieldsFunc(hostname, func(r rune) bool {
		return r == '.' || r == '-'
	})

	for _, h := range hints {
		for _, part := range parts {
			if part == h.pattern || strings.HasPrefix(part, h.pattern) {
				result.Success = true
				result.City = h.loc.City
				result.Region = h.loc.Region
				result.Country = h.loc.Country
				result.Org = hostname
				return result
			}
		}
	}

	// Try to extract country from ccTLD (e.g. ".br" → Brazil, ".de" → Germany).
	// This handles hostnames like "cpanel-100.playax.net.br" where the TLD
	// indicates the country even though there's no datacenter location code.
	ccTLDs := map[string]string{
		"ar": "Argentina", "at": "Austria", "au": "Australia", "be": "Belgium",
		"bg": "Bulgaria", "br": "Brazil", "ca": "Canada", "ch": "Switzerland",
		"cl": "Chile", "cn": "China", "co": "Colombia", "cz": "Czech Republic",
		"de": "Germany", "dk": "Denmark", "ee": "Estonia", "es": "Spain",
		"fi": "Finland", "fr": "France", "gb": "United Kingdom", "gr": "Greece",
		"hk": "Hong Kong", "hr": "Croatia", "hu": "Hungary", "id": "Indonesia",
		"ie": "Ireland", "il": "Israel", "in": "India", "ir": "Iran",
		"is": "Iceland", "it": "Italy", "jp": "Japan", "ke": "Kenya",
		"kr": "South Korea", "lt": "Lithuania", "lu": "Luxembourg", "lv": "Latvia",
		"mx": "Mexico", "my": "Malaysia", "ng": "Nigeria", "nl": "Netherlands",
		"no": "Norway", "nz": "New Zealand", "pe": "Peru", "ph": "Philippines",
		"pk": "Pakistan", "pl": "Poland", "pt": "Portugal", "ro": "Romania",
		"rs": "Serbia", "ru": "Russia", "sa": "Saudi Arabia", "se": "Sweden",
		"sg": "Singapore", "si": "Slovenia", "sk": "Slovakia", "th": "Thailand",
		"tr": "Turkey", "tw": "Taiwan", "ua": "Ukraine", "uk": "United Kingdom",
		"us": "United States", "uy": "Uruguay", "vn": "Vietnam", "za": "South Africa",
	}

	// Strip in-addr.arpa suffix if present (misconfigured PTR records)
	cleaned := hostname
	if idx := strings.Index(cleaned, ".in-addr.arpa"); idx >= 0 {
		cleaned = cleaned[:idx]
	}

	// Extract the TLD (last dot-separated segment)
	dotParts := strings.Split(cleaned, ".")
	if len(dotParts) >= 2 {
		tld := dotParts[len(dotParts)-1]
		if country, ok := ccTLDs[tld]; ok {
			result.Success = true
			result.Country = country
			result.Org = hostname
			return result
		}
	}

	// No location hint found in hostname, but record the PTR for reference
	result.Error = fmt.Sprintf("no location hint in %s", hostname)
	return result
}

// normalizeCountry maps ISO 3166-1 alpha-2 codes and common variants to a
// canonical full country name so that votes from different sources unify.
func normalizeCountry(raw string) string {
	// Common ISO alpha-2 → full name mapping (extend as needed)
	codeToName := map[string]string{
		"AD": "Andorra", "AE": "United Arab Emirates", "AF": "Afghanistan",
		"AG": "Antigua and Barbuda", "AL": "Albania", "AM": "Armenia",
		"AO": "Angola", "AR": "Argentina", "AT": "Austria", "AU": "Australia",
		"AZ": "Azerbaijan", "BA": "Bosnia and Herzegovina", "BB": "Barbados",
		"BD": "Bangladesh", "BE": "Belgium", "BG": "Bulgaria", "BH": "Bahrain",
		"BM": "Bermuda", "BN": "Brunei", "BO": "Bolivia", "BR": "Brazil",
		"BS": "Bahamas", "BT": "Bhutan", "BW": "Botswana", "BY": "Belarus",
		"BZ": "Belize", "CA": "Canada", "CD": "DR Congo", "CF": "Central African Republic",
		"CG": "Congo", "CH": "Switzerland", "CI": "Ivory Coast", "CL": "Chile",
		"CM": "Cameroon", "CN": "China", "CO": "Colombia", "CR": "Costa Rica",
		"CU": "Cuba", "CY": "Cyprus", "CZ": "Czech Republic", "DE": "Germany",
		"DJ": "Djibouti", "DK": "Denmark", "DO": "Dominican Republic",
		"DZ": "Algeria", "EC": "Ecuador", "EE": "Estonia", "EG": "Egypt",
		"ES": "Spain", "ET": "Ethiopia", "FI": "Finland", "FJ": "Fiji",
		"FR": "France", "GA": "Gabon", "GB": "United Kingdom", "GE": "Georgia",
		"GH": "Ghana", "GR": "Greece", "GT": "Guatemala", "GY": "Guyana",
		"HK": "Hong Kong", "HN": "Honduras", "HR": "Croatia", "HT": "Haiti",
		"HU": "Hungary", "ID": "Indonesia", "IE": "Ireland", "IL": "Israel",
		"IN": "India", "IQ": "Iraq", "IR": "Iran", "IS": "Iceland",
		"IT": "Italy", "JM": "Jamaica", "JO": "Jordan", "JP": "Japan",
		"KE": "Kenya", "KG": "Kyrgyzstan", "KH": "Cambodia", "KR": "South Korea",
		"KW": "Kuwait", "KZ": "Kazakhstan", "LA": "Laos", "LB": "Lebanon",
		"LI": "Liechtenstein", "LK": "Sri Lanka", "LT": "Lithuania",
		"LU": "Luxembourg", "LV": "Latvia", "LY": "Libya", "MA": "Morocco",
		"MC": "Monaco", "MD": "Moldova", "ME": "Montenegro", "MG": "Madagascar",
		"MK": "North Macedonia", "ML": "Mali", "MM": "Myanmar", "MN": "Mongolia",
		"MO": "Macau", "MT": "Malta", "MU": "Mauritius", "MV": "Maldives",
		"MW": "Malawi", "MX": "Mexico", "MY": "Malaysia", "MZ": "Mozambique",
		"NA": "Namibia", "NG": "Nigeria", "NI": "Nicaragua", "NL": "Netherlands",
		"NO": "Norway", "NP": "Nepal", "NZ": "New Zealand", "OM": "Oman",
		"PA": "Panama", "PE": "Peru", "PH": "Philippines", "PK": "Pakistan",
		"PL": "Poland", "PR": "Puerto Rico", "PS": "Palestine", "PT": "Portugal",
		"PY": "Paraguay", "QA": "Qatar", "RO": "Romania", "RS": "Serbia",
		"RU": "Russia", "RW": "Rwanda", "SA": "Saudi Arabia", "SC": "Seychelles",
		"SD": "Sudan", "SE": "Sweden", "SG": "Singapore", "SI": "Slovenia",
		"SK": "Slovakia", "SN": "Senegal", "SO": "Somalia", "SV": "El Salvador",
		"SY": "Syria", "TH": "Thailand", "TJ": "Tajikistan", "TN": "Tunisia",
		"TR": "Turkey", "TT": "Trinidad and Tobago", "TW": "Taiwan",
		"TZ": "Tanzania", "UA": "Ukraine", "UG": "Uganda", "US": "United States",
		"UY": "Uruguay", "UZ": "Uzbekistan", "VE": "Venezuela", "VN": "Vietnam",
		"ZA": "South Africa", "ZM": "Zambia", "ZW": "Zimbabwe",
	}

	trimmed := strings.TrimSpace(raw)
	upper := strings.ToUpper(trimmed)

	// If it's a 2-3 letter code, look it up
	if len(trimmed) <= 3 {
		if name, ok := codeToName[upper]; ok {
			return name
		}
	}

	// Already a full name — normalize casing for consistency
	// Capitalize first letter of each word manually to avoid deprecated strings.Title
	words := strings.Fields(strings.ToLower(trimmed))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func calculateConsensus(sources []GeoResult) AggregatedLocation {
	agg := AggregatedLocation{
		Sources: sources,
	}

	if len(sources) == 0 {
		agg.Consensus = "No location data available"
		return agg
	}

	// Count successful sources
	var successful []GeoResult
	for _, s := range sources {
		if s.Success {
			successful = append(successful, s)
		}
	}

	if len(successful) == 0 {
		agg.Consensus = "All sources failed"
		return agg
	}

	// Count votes for each value
	countryVotes := make(map[string]int)
	regionVotes := make(map[string]int)
	cityVotes := make(map[string]int)
	asnVotes := make(map[string]int)
	ispVotes := make(map[string]int)

	for _, s := range successful {
		if s.Country != "" {
			countryVotes[normalizeCountry(s.Country)]++
		}
		if s.Region != "" {
			regionVotes[s.Region]++
		}
		if s.City != "" {
			cityVotes[s.City]++
		}
		if s.ASN != "" {
			asnVotes[s.ASN]++
		}
		if s.ISP != "" {
			ispVotes[s.ISP]++
		} else if s.ASNOrg != "" {
			ispVotes[s.ASNOrg]++
		}
	}

	// Pick most voted values
	agg.Country = getMostVoted(countryVotes)
	agg.Region = getMostVoted(regionVotes)
	agg.City = getMostVoted(cityVotes)
	agg.ASN = getMostVoted(asnVotes)
	agg.ISP = getMostVoted(ispVotes)

	// Detect datacenter
	for _, s := range successful {
		if s.Datacenter != "" {
			agg.Datacenter = s.Datacenter
			break
		}
	}

	// Calculate confidence based on agreement
	maxVotes := len(successful)
	countryAgreement := float64(countryVotes[agg.Country]) / float64(maxVotes)
	cityAgreement := 0.0
	if agg.City != "" {
		cityAgreement = float64(cityVotes[agg.City]) / float64(maxVotes)
	}

	// Confidence formula: weighted average of agreement levels
	agg.Confidence = (countryAgreement*0.4 + cityAgreement*0.6) * 100

	// Build consensus description
	agreementLevel := "Low"
	if agg.Confidence >= 80 {
		agreementLevel = "High"
	} else if agg.Confidence >= 50 {
		agreementLevel = "Medium"
	}

	agg.Consensus = fmt.Sprintf("%s confidence (%d/%d sources agree on country, %d/%d on city)",
		agreementLevel,
		countryVotes[agg.Country], maxVotes,
		cityVotes[agg.City], maxVotes)

	return agg
}

func getMostVoted(votes map[string]int) string {
	var maxKey string
	maxCount := 0
	for k, v := range votes {
		if v > maxCount {
			maxCount = v
			maxKey = k
		}
	}
	return maxKey
}

func outputTextResults(results []ValidatorLocation, totalPower, threshold int64, pct float64) {
	var cumulativePower int64

	for _, r := range results {
		cumulativePower += r.Tokens

		fmt.Printf("Rank: %d\n", r.Rank)
		fmt.Printf("  Operator Address: %s\n", r.OperatorAddress)
		fmt.Printf("  Moniker: %s\n", r.Moniker)
		fmt.Printf("  Voting Power: %d tokens (%.2f%%)\n", r.Tokens, r.VotingPowerPct)

		if r.Website != "" {
			fmt.Printf("  Website: %s\n", r.Website)
		}

		if r.PeerIP != "" {
			fmt.Printf("  Peer IP: %s\n", r.PeerIP)
		}

		if len(r.ResolvedIPs) > 0 {
			fmt.Printf("  Website IPs: %s\n", strings.Join(r.ResolvedIPs, ", "))
		}

		loc := r.Location
		if loc.Country != "" {
			locationStr := loc.Country
			if loc.Region != "" {
				locationStr = loc.Region + ", " + locationStr
			}
			if loc.City != "" {
				locationStr = loc.City + ", " + locationStr
			}
			fmt.Printf("  Geographic Location: %s\n", locationStr)

			if loc.ASN != "" {
				asnInfo := loc.ASN
				if loc.ISP != "" {
					asnInfo += " (" + loc.ISP + ")"
				}
				fmt.Printf("  Network: %s\n", asnInfo)
			} else if loc.ISP != "" {
				fmt.Printf("  ISP: %s\n", loc.ISP)
			}

			if loc.Datacenter != "" {
				fmt.Printf("  Datacenter: %s\n", loc.Datacenter)
			}

			fmt.Printf("  Confidence: %.0f%% - %s\n", loc.Confidence, loc.Consensus)
		} else {
			fmt.Printf("  Geographic Location: Unknown - %s\n", loc.Consensus)
		}

		// Show individual source results
		fmt.Println("  Source Results:")
		for _, s := range loc.Sources {
			if s.Success {
				fmt.Printf("    - %s: %s, %s, %s\n", s.Source, s.City, s.Region, s.Country)
			} else {
				fmt.Printf("    - %s: failed (%s)\n", s.Source, s.Error)
			}
		}

		fmt.Println()
	}

	// Summary
	fmt.Println("========================================")
	fmt.Println("Summary:")
	fmt.Printf("  Validators in top %.1f%%: %d\n", pct, len(results))
	fmt.Printf("  Cumulative voting power: %d tokens (%.2f%%)\n",
		cumulativePower, float64(cumulativePower)/float64(totalPower)*100)
	fmt.Printf("  Target threshold: %d tokens (%.1f%%)\n", threshold, pct)

	// Geographic distribution
	countryCount := make(map[string]int)
	unknownCount := 0
	for _, r := range results {
		if r.Location.Country != "" {
			countryCount[r.Location.Country]++
		} else {
			unknownCount++
		}
	}

	fmt.Println("\nGeographic Distribution:")
	type countryEntry struct {
		name  string
		count int
	}
	var countries []countryEntry
	for c, n := range countryCount {
		countries = append(countries, countryEntry{c, n})
	}
	sort.Slice(countries, func(i, j int) bool {
		return countries[i].count > countries[j].count
	})

	for _, c := range countries {
		fmt.Printf("  %s: %d validators\n", c.name, c.count)
	}
	if unknownCount > 0 {
		fmt.Printf("  Unknown: %d validators\n", unknownCount)
	}
}

func outputJSONResults(results []ValidatorLocation, totalPower, threshold int64, pct float64) {
	output := struct {
		TotalPower     int64               `json:"total_power"`
		Threshold      int64               `json:"threshold"`
		Percentage     float64             `json:"percentage"`
		ValidatorCount int                 `json:"validator_count"`
		Validators     []ValidatorLocation `json:"validators"`
		CountryDistrib map[string]int      `json:"country_distribution"`
	}{
		TotalPower:     totalPower,
		Threshold:      threshold,
		Percentage:     pct,
		ValidatorCount: len(results),
		Validators:     results,
		CountryDistrib: make(map[string]int),
	}

	for _, r := range results {
		if r.Location.Country != "" {
			output.CountryDistrib[r.Location.Country]++
		} else {
			output.CountryDistrib["Unknown"]++
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(output)
}
