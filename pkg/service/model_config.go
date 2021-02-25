// generated-from:d5d5aa0731228b23a10b49fa2c69819a212e4565205c0258432d2b35dba9f169 DO NOT REMOVE, DO UPDATE

package service

type GlobalConfig struct {
	ACHTestHarness Config
}

// Config defines all the configuration for the app
type Config struct {
	Servers   ServerConfig
	Responses []Response
}

// ServerConfig - Groups all the http configs for the servers and ports that get opened.
type ServerConfig struct {
	FTP   *FTPConfig
	Admin HTTPConfig
}

// FTPConfig configuration for running an FTP server
type FTPConfig struct {
	RootPath string
	Hostname string
	Port     int
}

// HTTPConfig configuration for running an http server
type HTTPConfig struct {
	Bind BindAddress
}

// BindAddress specifies where the http server should bind to.
type BindAddress struct {
	Address string
}

type Response struct {
	// TODO(adam):
}
