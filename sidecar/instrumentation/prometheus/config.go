package prometheus

type Config struct {
	Enabled            bool
	ListenAddress      string
	MaxOpenConnections int
	Namespace          string
}
