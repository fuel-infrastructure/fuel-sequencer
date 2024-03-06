package types

// The following code was adapted from this library https://github.com/alanchchen/web3go/blob/e0f95297bead/web3/web3.go#L300
import (
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

// IsAddress checks if the given string is an address.
func IsAddress(address string) bool {
	smallCapsMatcher := regexp.MustCompile("^(0x)?[0-9a-f]{40}$")
	smallCapsMatched := smallCapsMatcher.MatchString(address)
	allCapsMatcher := regexp.MustCompile("^(0x)?[0-9A-F]{40}$")
	allCapsMatched := allCapsMatcher.MatchString(address)
	if smallCapsMatched || allCapsMatched {
		return true
	}
	return isChecksumAddress(address)
}

func isChecksumAddress(address string) bool {
	addr := strings.Replace(address, "0x", "", -1)
	addressHash := Sha3(strings.ToLower(addr))
	for i := 0; i < 40; i++ {
		d, err := strconv.ParseInt(string(addressHash[i]), 16, 32)
		if err != nil {
			return false
		}

		if d > 7 && strings.ToUpper(string(address[i])) == string(address[i]) ||
			d <= 7 && strings.ToLower(string(address[i])) == string(address[i]) {
			return false
		}
	}
	return true
}

// Sha3 returns Keccak-256 (not the standardized SHA3-256) of the given data.
func Sha3(data string) string {
	d := sha3.NewLegacyKeccak256()
	d.Write([]byte(data))
	return hex.EncodeToString(d.Sum(nil))
}
