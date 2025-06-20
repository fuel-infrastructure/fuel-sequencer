package testsuite

// EnableFlags controls which components should be enabled during test setup
type EnableFlags struct {
	Proxy     bool
	Ethereum  bool
	Sequencer bool
}

// newEnableFlags returns a new ComponentFlags with default values
func newEnableFlags() *EnableFlags {
	return &EnableFlags{
		Proxy:     false,
		Ethereum:  true,
		Sequencer: true,
	}
}

// Reset resets all flags to disabled
func (cf *EnableFlags) Reset() {
	cf.Proxy = false
	cf.Ethereum = false
	cf.Sequencer = false
}

// EnableProxy enables the proxy for the test suite. This should be called before SetupTest().
func (s *E2ETestSuite) EnableProxy() {
	s.enabled.Proxy = true
}

// DisableProxy disables the proxy for the test suite. This should be called before SetupTest().
func (s *E2ETestSuite) DisableProxy() {
	s.enabled.Proxy = false
}

// IsProxyEnabled returns whether the proxy is enabled for this test suite.
func (s *E2ETestSuite) IsProxyEnabled() bool {
	return s.enabled.Proxy
}

// EnsureProxyRunning ensures the proxy is running. If it's not enabled, it will be started.
func (s *E2ETestSuite) EnsureProxyRunning() {
	if !s.enabled.Proxy {
		s.enabled.Proxy = true
		s.runProxyContainer()
	} else if s.proxyResource == nil {
		s.runProxyContainer()
	}
}

// EnableEthereum enables the Ethereum components for the test suite. This should be called before SetupTest().
func (s *E2ETestSuite) EnableEthereum() {
	s.enabled.Ethereum = true
}

// DisableEthereum disables the Ethereum components for the test suite. This should be called before SetupTest().
func (s *E2ETestSuite) DisableEthereum() {
	s.enabled.Ethereum = false
}

// IsEthereumEnabled returns whether the Ethereum components are enabled for this test suite.
func (s *E2ETestSuite) IsEthereumEnabled() bool {
	return s.enabled.Ethereum
}

// EnableSequencer enables the sequencer components for the test suite. This should be called before SetupTest().
func (s *E2ETestSuite) EnableSequencer() {
	s.enabled.Sequencer = true
}

// DisableSequencer disables the sequencer components for the test suite. This should be called before SetupTest().
func (s *E2ETestSuite) DisableSequencer() {
	s.enabled.Sequencer = false
}

// IsSequencerEnabled returns whether the sequencer components are enabled for this test suite.
func (s *E2ETestSuite) IsSequencerEnabled() bool {
	return s.enabled.Sequencer
}
