package golanglibs

import (
	"crypto/tls"
	"crypto/x509"
	"net"
)

type sslStruct struct {
	Listen        func(host string, port int, key string, crt string) *tcpServerSideListener
	ServerWrapper func(conn net.Conn, key string, crt string) *TcpServerSideConn
	Connect       func(host string, port int, cfg ...SSLCfg) *sslClientSideConn
	ClientWrapper func(conn net.Conn, cfg ...SSLCfg) *sslClientSideConn
}

var sslstruct sslStruct

func init() {
	sslstruct = sslStruct{
		Listen:        sslListen,
		ServerWrapper: sslServerWrapper,
		Connect:       sslConnect,
		ClientWrapper: sslClientWrapper,
	}
}

// SSL - Server
// 只实现了一个Listener， 其他的方法是tcp的方法

func sslListen(host string, port int, key string, crt string) *tcpServerSideListener {
	cert, err := tls.X509KeyPair([]byte(crt), []byte(key))
	Panicerr(err)

	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	listener, err := tls.Listen("tcp4", host+":"+Str(port), tlsCfg)
	Panicerr(err)

	return &tcpServerSideListener{listener: listener}
}

func sslServerWrapper(conn net.Conn, key string, crt string) *TcpServerSideConn {
	cert, err := tls.X509KeyPair([]byte(crt), []byte(key))
	Panicerr(err)

	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	tconn := tls.Server(conn, tlsCfg)

	return &TcpServerSideConn{Conn: tconn}
}

// SSL - Client

type sslClientSideConn struct {
	conn *tls.Conn
}

type SSLCfg struct {
	InsecureSkipVerify  bool     // true为跳过证书验证
	AdditionRootCA      []string // 额外的用来验证证书的CA证书
	Domain              string   // 需要认证的域名, 也会在请求证书的时候提供
	WithoutSystemRootCA bool     // true为不使用系统内置的CA
}

func sslConnect(host string, port int, cfg ...SSLCfg) *sslClientSideConn {
	servAddr := host + ":" + Str(port)

	tcfg := tls.Config{}
	if len(cfg) != 0 {
		if cfg[0].InsecureSkipVerify {
			tcfg.InsecureSkipVerify = cfg[0].InsecureSkipVerify
		}
		if len(cfg[0].AdditionRootCA) != 0 {
			var rootCAs *x509.CertPool
			if cfg[0].WithoutSystemRootCA {
				rootCAs = x509.NewCertPool()
			} else {
				rootCAs, _ = x509.SystemCertPool()
				if rootCAs == nil {
					rootCAs = x509.NewCertPool()
				}
			}
			for _, ca := range cfg[0].AdditionRootCA {
				rootCAs.AppendCertsFromPEM([]byte(ca))
			}
			tcfg.RootCAs = rootCAs
		}
		if cfg[0].Domain != "" {
			tcfg.ServerName = cfg[0].Domain
		}
	}

	conn, err := tls.Dial("tcp", servAddr, &tcfg)
	Panicerr(err)
	return &sslClientSideConn{conn: conn}
}

func sslClientWrapper(conn net.Conn, cfg ...SSLCfg) *sslClientSideConn {
	tcfg := tls.Config{}
	if len(cfg) != 0 {
		if cfg[0].InsecureSkipVerify {
			tcfg.InsecureSkipVerify = cfg[0].InsecureSkipVerify
		}
		if len(cfg[0].AdditionRootCA) != 0 {
			var rootCAs *x509.CertPool
			if cfg[0].WithoutSystemRootCA {
				rootCAs = x509.NewCertPool()
			} else {
				rootCAs, _ = x509.SystemCertPool()
				if rootCAs == nil {
					rootCAs = x509.NewCertPool()
				}
			}
			for _, ca := range cfg[0].AdditionRootCA {
				rootCAs.AppendCertsFromPEM([]byte(ca))
			}
			tcfg.RootCAs = rootCAs
		}
		if cfg[0].Domain != "" {
			tcfg.ServerName = cfg[0].Domain
		}
	}

	tconn := tls.Client(conn, &tcfg)
	return &sslClientSideConn{conn: tconn}
}

func (m *sslClientSideConn) Send(str string) {
	_, err := m.conn.Write([]byte(str))
	Panicerr(err)
}

func (m *sslClientSideConn) Recv(buffersize int) string {
	reply := make([]byte, buffersize)
	n, err := m.conn.Read(reply)
	Panicerr(err)
	return string(reply[:n])
}

func (m *sslClientSideConn) Close() {
	err := m.conn.Close()
	Panicerr(err)
}
