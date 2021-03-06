package golanglibs

import (
	"crypto/sha256"

	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
)

type kcpStruct struct {
	Listen  func(host string, port int, key string, salt string) *kcpServerSideListener
	Connect func(host string, port int, key string, salt string) *kcp.UDPSession
}

var kcpstruct kcpStruct

func init() {
	kcpstruct = kcpStruct{
		Listen:  kcpListen,
		Connect: kcpConnect,
	}
}

type kcpServerSideListener struct {
	listener *kcp.Listener
}

func kcpListen(host string, port int, key string, salt string) *kcpServerSideListener {
	block, err := kcp.NewAESBlockCrypt(pbkdf2.Key([]byte(key), []byte(salt), 4096, 32, sha256.New))
	Panicerr(err)

	l, err := kcp.ListenWithOptions(host+":"+Str(port), block, 10, 3)
	Panicerr(err)

	l.SetDSCP(46)
	l.SetReadBuffer(4194304)
	l.SetWriteBuffer(4194304)

	return &kcpServerSideListener{listener: l}
}

func (m *kcpServerSideListener) Accept() chan *kcp.UDPSession {
	ch := make(chan *kcp.UDPSession)

	go func() {
		for {
			c, err := m.listener.AcceptKCP()
			if err != nil {
				if String("io: read/write on closed pipe").In(err.Error()) || String("use of closed network connection").In(err.Error()) {
					close(ch)
					break
				}
				Panicerr(err)
			}

			c.SetNoDelay(0, 20, 2, 1)
			c.SetMtu(1400)
			c.SetWindowSize(1024, 1024)
			c.SetACKNoDelay(false)

			ch <- c
		}
	}()

	return ch
}

func kcpConnect(host string, port int, key string, salt string) *kcp.UDPSession {
	block, err := kcp.NewAESBlockCrypt(pbkdf2.Key([]byte(key), []byte(salt), 4096, 32, sha256.New))
	Panicerr(err)
	conn, err := kcp.DialWithOptions(host+":"+Str(port), block, 10, 3)
	Panicerr(err)

	conn.SetMtu(1400)
	conn.SetWriteDelay(false)
	conn.SetNoDelay(0, 20, 2, 1)
	conn.SetWindowSize(128, 1024)
	conn.SetACKNoDelay(false)
	conn.SetDSCP(46)
	conn.SetReadBuffer(4194304)
	conn.SetWriteBuffer(4194304)

	return conn
}
