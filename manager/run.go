package manager

// cfssl genkey csr.json
// cfssl genkey -initca server-csr.json | cfssljson -bare ca

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"semaygateway.com/balance"
	"semaygateway.com/gatelogger"
	"semaygateway.com/gateparse"
)

var (
	start = &cobra.Command{
		Use:   "start",
		Short: "This app is domain forwarding proxy server ",
		Long:  `This proxy server forwards incoming request to provided services. It manages Hash keys and serves as api gateway.`,
		Run: func(cmd *cobra.Command, args []string) {
			app()
		},
	}
)

func app() {

	// for tls config
	// this is loading certificates
	serverTLSCert, err := tls.LoadX509KeyPair("ca.crt", "ca-key.pem")
	if err != nil {
		gatelogger.GateLoggerFatal(err.Error())
	}

	//  pining the certificate
	// cert, err := os.ReadFile("ca.csr")
	// if err != nil {
	// gatelogger.GateLoggerFatal(err.Error())
	// }

	// certPool := x509.NewCertPool()
	// if ok := certPool.AppendCertsFromPEM(cert); !ok {
	// 	gatelogger.GateLoggerFatal("failed to append certificate to pool")
	// }

	// pining ends here

	tlsConfig := &tls.Config{
		Certificates:     []tls.Certificate{serverTLSCert},
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion:       tls.VersionTLS13,
		// RootCAs:          certPool,
	}
	// instance of the server
	server := http.Server{
		Addr:      ":4500",
		TLSConfig: tlsConfig,
	}

	services, _ := gateparse.GetServiceLists()
	fmt.Println("In Manager run.go command file")

	for i := range services {
		if !(services[i] == "redis") && !(services[i] == "rabbit") {
			service_balance_method, _ := gateparse.GetLoadBalanceMethod(services[i])
			switch service_balance_method.Option {
			case "round_robbin":
				http.Handle("/"+services[i]+"/", http.HandlerFunc(balance.RobbinBalance))
			case "least_connection":
				http.Handle("/"+services[i]+"/", http.HandlerFunc(balance.RobbinBalance))
			default:
				http.Handle("/"+services[i]+"/", http.HandlerFunc(balance.RobbinBalance))

			}
		}

	}

	// http.Handle("/blue/", http.HandlerFunc(balance.RobbinBalance))
	gatelogger.GateLoggerFatal(server.ListenAndServeTLS("ca.crt", "ca-key.pem"))
	// gatelogger.GateLoggerFatal(http.ListenAndServe(":4500", nil))

}

func init() {
	bluegateway.AddCommand(start)

}
