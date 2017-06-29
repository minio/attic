/*
 * Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
 *
 * This file is part of Xray.
 *
 * Xray is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"fmt"
	"net"
	"net/http"

	router "github.com/gorilla/mux"
	"github.com/minio/cli"
)

var (
	// global flags for minio.
	globalFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: ":8080",
			Usage: `Bind to a specific IP:PORT.`,
		},
		cli.StringFlag{
			Name:  "cert",
			Value: globalXrayCertFile,
			Usage: "Path to SSL certificate file.",
		},
		cli.StringFlag{
			Name:  "key",
			Value: globalXrayKeyFile,
			Usage: "Path to SSL key file.",
		},
	}
)

// Help template for xray.
var xrayHelpTemplate = `NAME:
  {{.HelpName}}

DESCRIPTION:
  {{.Description}}

USAGE:
  {{.HelpName}} {{if .Flags}}[FLAGS] {{end}}

ENVIRONMENT VARIABLES:
  CASCADE:
     LBP_CASCADE: To enable LBP cascade image detector. Defaults to [Haar Cascade].
{{if .Commands}}
COMMANDS:
  {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
  {{end}}{{end}}{{if .Flags}}
FLAGS:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
`

// init - check the environment before main starts
func init() {
	// Check if minio was compiled using a supported version of Golang.
	checkGoVersion()
}

// getListenIPs - gets all the ips to listen on.
func getListenIPs(serverAddr string) (hosts []string, port string, err error) {
	var host string
	host, port, err = net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, port, fmt.Errorf("Unable to parse host address %s", err)
	}
	if host == "" {
		var ipv4s []net.IP
		ipv4s, err = getInterfaceIPv4s()
		if err != nil {
			return nil, port, fmt.Errorf("Unable reverse sort ips from hosts %s", err)
		}
		for _, ip := range ipv4s {
			hosts = append(hosts, ip.String())
		}
		return hosts, port, nil
	} // if host != "" {

	// Proceed to append itself, since user requested a specific endpoint.
	hosts = append(hosts, host)

	// Success.
	return hosts, port, nil
}

func registerApp() *cli.App {
	// Set up app.
	app := cli.NewApp()
	app.Name = "xray"
	app.Version = Version
	app.HideHelpCommand = true
	app.Author = "Minio.io"
	app.Description = `Deep learning based object detection for video.`
	app.Flags = globalFlags
	app.CustomAppHelpTemplate = xrayHelpTemplate
	app.Action = func(ctx *cli.Context) error {
		// Initialize a mux router.
		mux := router.NewRouter().SkipClean(true)
		httpServer := &http.Server{
			Addr:           ctx.String("address"),
			Handler:        configureXrayHandler(mux),
			MaxHeaderBytes: 1 << 20,
		}

		hosts, port, err := getListenIPs(httpServer.Addr)
		fatalIf(err, "Unable to get listen ips.")
		for _, host := range hosts {
			rlog.Printf("Started listening on ws://%s:%s", host, port)
		}

		// Start server, automatically configures TLS if certs are available.
		cert, key := ctx.String("cert"), ctx.String("key")
		if isCertFileExists(cert) && isKeyFileExists(key) {
			fatalIf(httpServer.ListenAndServeTLS(cert, key), "Failed to start xray server.")
		} else {
			fatalIf(httpServer.ListenAndServe(), "Failed to start xray server.")
		}
		return nil
	}
	return app
}

// Main - Xray server entry point.
func Main() {
	registerApp().RunAndExitOnError()
}
