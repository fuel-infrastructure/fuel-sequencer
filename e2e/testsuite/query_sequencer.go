package testsuite

import (
	"regexp"
	"strings"
)

func (s *E2ETestSuite) FindSequencerLogs(logRegexp *regexp.Regexp) (matches []string) {
	logs := s.logsByContainerID(s.valResources[0].Container.ID)

	for _, logStr := range strings.Split(logs, "\n") {
		matches = logRegexp.FindStringSubmatch(logStr)
		if len(matches) > 0 {
			return matches
		}
	}

	return nil
}
