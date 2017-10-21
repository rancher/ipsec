package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/leodotcloud/log"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/ipsec/arp"
	"github.com/rancher/ipsec/backend/ipsec"
	"github.com/rancher/ipsec/server"
	"github.com/rancher/ipsec/store"
)

const (
	metadataURLTemplate = "http://%v/2015-12-19"

	// DefaultMetadataAddress specifies the default value to use if nothing is specified
	DefaultMetadataAddress = "169.254.169.250"
)

var (
	// VERSION Of the binary
	VERSION = "0.0.0-dev"
)

const (
	metadataAddressFlag = "metadata-address"
)

func main() {
	app := cli.NewApp()
	app.Version = VERSION
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "ipsec-config, c",
			Value: ".",
			Usage: "Configuration directory",
		},
		cli.BoolTFlag{
			Name:  "gcm",
			Usage: "GCM mode Supported",
		},
		cli.StringFlag{
			Name: "charon-log",
		},
		cli.BoolFlag{
			Name: "charon-launch",
		},
		cli.BoolFlag{
			Name: "test-charon",
		},
		cli.BoolFlag{
			Name: "debug",
		},
		cli.StringFlag{
			Name:   "listen",
			Value:  "localhost:8111",
			EnvVar: "RANCHER_SERVICE_LISTEN_PORT",
		},
		cli.StringFlag{
			Name:   metadataAddressFlag,
			Value:  store.DefaultMetadataAddress,
			Usage:  "metadata address to use",
			EnvVar: "RANCHER_METADATA_ADDRESS",
		},
		cli.StringFlag{
			Name:   "ipsec-ike-sa-rekey-interval",
			Value:  ipsec.DefaultIkeSaRekeyInterval,
			Usage:  "IKE_SA rekey interval time",
			EnvVar: "IPSEC_IKE_SA_REKEY_INTERVAL",
		},
		cli.StringFlag{
			Name:   "ipsec-child-sa-rekey-interval",
			Value:  ipsec.DefaultChildSaRekeyInterval,
			Usage:  "CHILD_SA rekey interval time",
			EnvVar: "IPSEC_CHILD_SA_REKEY_INTERVAL",
		},
		cli.StringFlag{
			Name:   "ipsec-replay-window-size",
			Value:  ipsec.DefaultReplayWindowSize,
			Usage:  "IPSec Replay Window Size",
			EnvVar: "IPSEC_REPLAY_WINDOW_SIZE",
		},
	}
	app.Action = func(ctx *cli.Context) {
		if err := appMain(ctx); err != nil {
			log.Fatalf("error: %v", err)
		}
	}

	app.Run(os.Args)
}

func appMain(ctx *cli.Context) error {
	if ctx.GlobalBool("test-charon") {
		if err := ipsec.Test(); err != nil {
			log.Fatalf("Failed to talk to charon: %v", err)
		}
		os.Exit(0)
	}

	if ctx.GlobalBool("debug") {
		log.SetLevelString("debug")
	}

	done := make(chan error)

	log.Infof("Reading info from metadata")
	metadataAddress := ctx.GlobalString(metadataAddressFlag)
	if metadataAddress == "" {
		metadataAddress = DefaultMetadataAddress
	}
	metadataURL := fmt.Sprintf(metadataURLTemplate, metadataAddress)
	mc, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		log.Errorf("couldn't create metadata client: %v", err)
		return nil
	}

	db, err := store.NewMetadataStore(mc)
	if err != nil {
		log.Errorf("Error creating metadata store: %v", err)
		return err
	}

	db.Reload()

	ipsecOverlay := ipsec.NewOverlay(ctx.GlobalString("ipsec-config"), db, mc)
	ipsecOverlay.ReplayWindowSize = ctx.GlobalString("ipsec-replay-window-size")
	ipsecOverlay.IPSecIkeSaRekeyInterval = ctx.GlobalString("ipsec-ike-sa-rekey-interval")
	ipsecOverlay.IPSecChildSaRekeyInterval = ctx.GlobalString("ipsec-child-sa-rekey-interval")
	if !ctx.GlobalBool("gcm") {
		ipsecOverlay.Blacklist = []string{"aes128gcm16"}
	}
	overlay := ipsecOverlay
	overlay.Start(ctx.GlobalBool("charon-launch"), ctx.GlobalString("charon-log"))

	go func() {
		done <- arp.ListenAndServe(db, "eth0")
	}()

	listenPort := ctx.GlobalString("listen")
	log.Debugf("About to start server and listen on port: %v", listenPort)
	go func() {
		s := server.Server{
			Backend: overlay,
		}
		done <- s.ListenAndServe(listenPort)
	}()

	if err := overlay.Reload(); err != nil {
		log.Errorf("couldn't reload the overlay: %v", err)
		return err
	}

	return <-done
}
