package ftp

import (
	"bufio"
	"crypto/tls"
	"net"
)

// serverOpts contains parameters for server.NewServer()
type ServerOpts struct {
	Auth Auth

	// Server Name, Default is Go Ftp Server
	Name string

	// The hostname that the FTP server should listen on. Optional, defaults to
	// "::", which means all hostnames on ipv4 and ipv6.
	Hostname string

	// Public IP of the server
	PublicIp string

	// Passive ports
	PassivePorts string

	// use tls, default is false
	TLS bool

	// if tls used, cert file is required
	//CertFile string

	// if tls used, key file is required
	//KeyFile string

	// If ture TLS is used in RFC4217 mode
	ExplicitFTPS bool

	WelcomeMessage string
}

// Server is the root of your FTP application. You should instantiate one
// of these and call ListenAndServe() to start accepting client connections.
//
// Always use the NewServer() method to create a new Server.
type Server struct {
	*ServerOpts
	tlsConfig *tls.Config
}

// serverOptsWithDefaults copies an ServerOpts struct into a new struct,
// then adds any default values that are missing and returns the new data.
func serverOptsWithDefaults(opts *ServerOpts) *ServerOpts {
	var newOpts ServerOpts
	if opts == nil {
		opts = &ServerOpts{}
	}
	if opts.Hostname == "" {
		newOpts.Hostname = "::"
	} else {
		newOpts.Hostname = opts.Hostname
	}
	if opts.Name == "" {
		newOpts.Name = "Go FTP Server"
	} else {
		newOpts.Name = opts.Name
	}

	if opts.WelcomeMessage == "" {
		newOpts.WelcomeMessage = defaultWelcomeMessage
	} else {
		newOpts.WelcomeMessage = opts.WelcomeMessage
	}

	if opts.Auth != nil {
		newOpts.Auth = opts.Auth
	}

	newOpts.TLS = opts.TLS
	newOpts.ExplicitFTPS = opts.ExplicitFTPS

	newOpts.PublicIp = opts.PublicIp
	newOpts.PassivePorts = opts.PassivePorts

	return &newOpts
}

// NewServer initialises a new FTP server. Configuration options are provided
// via an instance of ServerOpts.
func NewServer(opts *ServerOpts) *Server {
	opts = serverOptsWithDefaults(opts)
	s := new(Server)
	s.ServerOpts = opts
	return s
}

// NewConn constructs a new object that will handle the FTP protocol over
// an active net.TCPConn. The TCP connection should already be open before
// it is handed to this functions. driver is an instance of FTPDriver that
// will handle all auth and persistence details.
func (server *Server) newConn(tcpConn net.Conn, driver Driver, recv chan string) *Conn {
	c := &Conn{
		namePrefix:    "/",
		conn:          tcpConn,
		controlReader: bufio.NewReader(tcpConn),
		controlWriter: bufio.NewWriter(tcpConn),
		driver:        driver,
		auth:          server.Auth,
		server:        server,
		sessionid:     newSessionID(),
		tlsConfig:     server.tlsConfig,
		rcv:           recv,
	}

	driver.Init()
	return c
}

func simpleTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	config := &tls.Config{}
	if config.NextProtos == nil {
		config.NextProtos = []string{"ftp"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return config, nil
}
