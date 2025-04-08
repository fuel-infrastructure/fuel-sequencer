/*
 * Usage:
 * go run main.go --upgrade-time "2025-04-10 14:00:00"
 *
 * This will calculate the block height at which the upgrade will happen
 * based on the average block time of the last 1000 blocks.
 *
 * It will then find the "nice" block heights that end in 000, 00, and 0
 * and print them out.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"
)

const (
	rpcURL = "https://testnet-rpc-fuel-seq.simplystaking.xyz"
)

// Block response structure
type BlockResponse struct {
	Result struct {
		Block struct {
			Header struct {
				Height    string    `json:"height"`
				Time      time.Time `json:"time"`
				ChainID   string    `json:"chain_id"`
				LastBlock struct {
					Height string `json:"height"`
				} `json:"last_block_id"`
			} `json:"header"`
		} `json:"block"`
	} `json:"result"`
}

// Latest block height response
type StatusResponse struct {
	Result struct {
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"sync_info"`
	} `json:"result"`
}

func main() {
	// Define command line arguments
	upgradeDate := flag.String("upgrade-time", "", "Upgrade time in format 'YYYY-MM-DD HH:MM:SS' in CET timezone")
	flag.Parse()

	// Get latest block height
	latestHeight, err := getLatestBlockHeight()
	if err != nil {
		fmt.Printf("Error getting latest block height: %v\n", err)
		os.Exit(1)
	}

	// Get genesis block (height 1)
	genesisBlock, err := getBlock("1")
	if err != nil {
		fmt.Printf("Error getting genesis block: %v\n", err)
		os.Exit(1)
	}

	// Get latest block
	latestBlock, err := getBlock(latestHeight)
	if err != nil {
		fmt.Printf("Error getting latest block: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Chain ID: %s\n", genesisBlock.Result.Block.Header.ChainID)
	fmt.Printf("Genesis Block Time: %s\n", genesisBlock.Result.Block.Header.Time)
	fmt.Printf("Latest Block: %s at %s\n", latestHeight, latestBlock.Result.Block.Header.Time)

	// Setup table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, "Sample Size\tStart Block\tEnd Block\tTotal Time\tAvg Block Time (sec)")

	// Calculate for all blocks from genesis
	allAvgTime := calculateAverageBlockTime(w, "1", latestHeight, "All")

	// Calculate for the last 100 blocks
	b100AvgTime := calculateForLastN(w, latestHeight, 100)

	// Calculate for the last 1000 blocks
	b1000AvgTime := calculateForLastN(w, latestHeight, 1000)

	// Calculate for the last 10000 blocks
	calculateForLastN(w, latestHeight, 10000)

	// Calculate for the last 100000 blocks (if chain has enough blocks)
	calculateForLastN(w, latestHeight, 100000)

	w.Flush()

	// Handle upgrade time estimation if provided
	if *upgradeDate != "" {
		// Parse the upgrade time
		loc, _ := time.LoadLocation("CET")
		upgradeTime, err := time.ParseInLocation("2006-01-02 15:04:05", *upgradeDate, loc)
		if err != nil {
			fmt.Printf("Error parsing upgrade time: %v\n", err)
			os.Exit(1)
		}

		// Estimate block height at upgrade time using different averages
		fmt.Println("\nEstimated Upgrade Block Heights:")
		fmt.Println("=================================")

		// Use the most recent block time averages (more accurate for near-term projections)
		// Prioritize using 1000-block average if available
		var avgBlockTime float64
		if b1000AvgTime > 0 {
			avgBlockTime = b1000AvgTime
			fmt.Println("Using last 1000 blocks average time for projection")
		} else if b100AvgTime > 0 {
			avgBlockTime = b100AvgTime
			fmt.Println("Using last 100 blocks average time for projection")
		} else {
			avgBlockTime = allAvgTime
			fmt.Println("Using all-time average block time for projection")
		}

		latestBlockTime := latestBlock.Result.Block.Header.Time
		timeUntilUpgrade := upgradeTime.Sub(latestBlockTime)
		blocksUntilUpgrade := int(timeUntilUpgrade.Seconds() / avgBlockTime)

		estimatedHeight := toInt(latestHeight) + blocksUntilUpgrade
		fmt.Printf("Current time: %s\n", latestBlockTime)
		fmt.Printf("Target upgrade time: %s\n", upgradeTime)
		fmt.Printf("Time until upgrade: %s\n", timeUntilUpgrade)
		fmt.Printf("Blocks until upgrade (@ %.2fs per block): %d\n", avgBlockTime, blocksUntilUpgrade)
		fmt.Printf("Estimated raw block height at upgrade time: %d\n", estimatedHeight)

		// Find a "nice" block height (ending in zeros)
		niceHeights := findNiceHeights(estimatedHeight, avgBlockTime, upgradeTime)

		fmt.Println("\nRecommended Upgrade Heights (with adjusted times):")
		fmt.Println("=================================================")
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "Height\tTime (CET)\tDelay\tNice Factor")

		for _, nh := range niceHeights {
			delay := nh.time.Sub(upgradeTime)
			var niceFactor string
			if nh.height%1000 == 0 {
				niceFactor = "⭐⭐⭐" // Ends with 3 zeros
			} else if nh.height%100 == 0 {
				niceFactor = "⭐⭐" // Ends with 2 zeros
			} else if nh.height%10 == 0 {
				niceFactor = "⭐" // Ends with 1 zero
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", nh.height, nh.time.Format("2006-01-02 15:04:05"), formatDelay(delay), niceFactor)
		}
		w.Flush()
	}
}

// NiceHeight represents a block height that ends in zeros and its estimated time
type NiceHeight struct {
	height int
	time   time.Time
}

// Find block heights ending in zeros near the estimated height
func findNiceHeights(estimatedHeight int, avgBlockTime float64, targetTime time.Time) []NiceHeight {
	var results []NiceHeight

	// Look at heights ending in 000, 00, and 0 within a reasonable range
	rangeSize := 2000 // Search within 2000 blocks (before and after)

	// Helper function to calculate time for a height and add to results if it meets criteria
	addIfNice := func(height int) {
		// Calculate time difference for this height
		blockDiff := height - estimatedHeight
		timeDiff := time.Duration(float64(blockDiff) * avgBlockTime * float64(time.Second))
		blockTime := targetTime.Add(timeDiff)

		// Only consider heights that are either exact or delay the upgrade (not before)
		delay := blockTime.Sub(targetTime)
		// Allow heights slightly before target time (within 30 minutes)
		if delay > -30*time.Minute {
			results = append(results, NiceHeight{height: height, time: blockTime})
		}
	}

	// Find the next height ending in 000 after the estimated height
	nextHeight1000 := ((estimatedHeight / 1000) * 1000) + 1000
	if nextHeight1000-estimatedHeight <= rangeSize {
		addIfNice(nextHeight1000)
	}

	// Find heights ending in 00 (within range)
	start100 := ((estimatedHeight - rangeSize) / 100) * 100
	if start100 < estimatedHeight {
		start100 += 100
	}

	for h := start100; h <= estimatedHeight+rangeSize; h += 100 {
		if h%1000 != 0 { // Skip if already added as 000
			addIfNice(h)
		}
	}

	// Add some heights ending in 0 (cherry-pick a few close to target)
	start10 := ((estimatedHeight - 200) / 10) * 10
	if start10 < estimatedHeight {
		start10 += 10
	}

	for h := start10; h <= estimatedHeight+200; h += 10 {
		if h%100 != 0 { // Skip if already added as 00 or 000
			addIfNice(h)
		}
	}

	// Sort by absolute delay from target time
	// Prioritize heights that are closest to the target time
	// But avoid complex sorting in this script

	return results
}

// Format time delay in a readable format
func formatDelay(delay time.Duration) string {
	if delay < 0 {
		return fmt.Sprintf("-%s", formatDuration(-delay))
	}
	return formatDuration(delay)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func getLatestBlockHeight() (string, error) {
	resp, err := http.Get(rpcURL + "/status")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var statusResp StatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return "", err
	}

	return statusResp.Result.SyncInfo.LatestBlockHeight, nil
}

func getBlock(height string) (*BlockResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/block?height=%s", rpcURL, height))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var blockResp BlockResponse
	if err := json.Unmarshal(body, &blockResp); err != nil {
		return nil, err
	}

	return &blockResp, nil
}

func calculateAverageBlockTime(w *tabwriter.Writer, startHeight, endHeight, label string) float64 {
	startBlock, err := getBlock(startHeight)
	if err != nil {
		fmt.Printf("Error getting block %s: %v\n", startHeight, err)
		return 0
	}

	endBlock, err := getBlock(endHeight)
	if err != nil {
		fmt.Printf("Error getting block %s: %v\n", endHeight, err)
		return 0
	}

	timeDiff := endBlock.Result.Block.Header.Time.Sub(startBlock.Result.Block.Header.Time)
	blockDiff := toInt(endHeight) - toInt(startHeight)
	avgBlockTime := timeDiff.Seconds() / float64(blockDiff)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.4f\n",
		label,
		startHeight,
		endHeight,
		timeDiff.String(),
		avgBlockTime)

	return avgBlockTime
}

func calculateForLastN(w *tabwriter.Writer, latestHeight string, n int) float64 {
	latestHeightInt := toInt(latestHeight)

	// Make sure we don't go below block 1
	startHeightInt := latestHeightInt - n + 1
	if startHeightInt < 1 {
		startHeightInt = 1
	}

	startHeight := fmt.Sprintf("%d", startHeightInt)

	// Calculate with actual sample size
	actualSampleSize := latestHeightInt - startHeightInt + 1
	label := fmt.Sprintf("Last %d", actualSampleSize)

	return calculateAverageBlockTime(w, startHeight, latestHeight, label)
}

func toInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
