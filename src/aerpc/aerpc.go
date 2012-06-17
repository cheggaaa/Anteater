package aerpc

import (
	"net/rpc"
	"os"
	"flag"
	"strings"
)

const (
	DEFAULT_HOST = "127.0.0.1"
	DEFAULT_PORT = "32032"
	SERVER_FLAG_FORMAT = `
	Server addr format:
	default addr: 127.0.0.1:32032
	examples:
		-s=192.168.1.2 will be 192.168.1.2:32032
		-s=:32033 will be 127.0.0.1:32033
		-s=anteater.local:3234 will be anteater.local:3234
	`;
)


type Conn struct {
	ServerAddr string
	Client *rpc.Client
	Command string 
	Args []string
	ShowHelp bool
}

func ParseFAndConnect() (*Conn, error) {
	var err error
	c := &Conn{}
	ServerAddr := flag.String("s", DEFAULT_HOST + ":" + DEFAULT_PORT, "Server addr")
	
	flag.Parse();
	var s int = 1
	s += flag.NFlag()

	if len(os.Args) - s >= 1 {
		c.Command = os.Args[s]
		s++
	} else {
		c.ShowHelp = true
		return c, nil
	}
	
	if len(os.Args) - s >= 1 {
		c.Args = os.Args[s:]
	}
	
	// check server addr
	addr := strings.Split(*ServerAddr, ":")
	var host, port string
	if len(addr) == 1 {
		host = addr[0]
		port = DEFAULT_PORT
	} else if len(addr) == 2 {
		host = addr[0]
		port = addr[1]
	}
	if len(host) == 0 {
		host = DEFAULT_HOST 
	}
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	c.ServerAddr = host + ":" + port
	c.Client, err = Connect(c.ServerAddr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Connect(addr string) (*rpc.Client, error) {
	return rpc.DialHTTP("tcp", addr)	
}
