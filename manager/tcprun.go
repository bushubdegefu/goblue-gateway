package manager

import (
	"fmt"
	"io"
	"net"

	"github.com/spf13/cobra"
	"semaygateway.com/gatelogger"
)

var (
	start_tcp = &cobra.Command{
		Use:   "starttcp",
		Short: "This app is domain forwarding proxy server",
		Long:  `This proxy server forwards incoming request to provided services. It manages Hash keys and serves as api gateway.`,
		Run: func(cmd *cobra.Command, args []string) {
			gatetcpapp()
		},
	}
)

func gatetcpapp() {

	fmt.Println("In Manager tcprun.go command file")

	listener, err := net.Listen("tcp", ":9500")
	fmt.Printf("listeneing on : %s", listener.Addr().String())
	if err != nil {
		panic("connection error:" + err.Error())
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			// fmt.Println("Accept Error:", err)
			gatelogger.GateLoggerInfo(err.Error())
			continue
		}
		fmt.Println(conn.RemoteAddr())
		copyConn(conn)
	}

}

func copyConn(src net.Conn) {
	dst, err := net.Dial("tcp", "localhost:7500")
	if err != nil {
		panic("Dial Error:" + err.Error())
	}

	done := make(chan struct{})

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(dst, src)
		done <- struct{}{}
	}()

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	<-done
	<-done
}

func init() {
	bluegateway.AddCommand(start_tcp)

}
