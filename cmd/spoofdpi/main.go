package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bariiss/SpoofDPI/proxy"
	"github.com/bariiss/SpoofDPI/util"
	"github.com/bariiss/SpoofDPI/util/log"
	"github.com/bariiss/SpoofDPI/version"
)

// main is the entry point of the application.
func main() {
	args := util.ParseArgs()

	if args.Version {
		version.PrintVersion()
		os.Exit(0)
	}

	config := util.GetConfig()
	config.Load(args)

	log.InitLogger(config)

	ctx := util.GetCtxWithScope(context.Background(), "MAIN")
	logger := log.GetCtxLogger(ctx)

	if !config.Silent {
		util.PrintColoredBanner()
	}

	if config.SystemProxy {
		if err := util.SetOsProxy(uint16(config.Port)); err != nil {
			logger.Fatal().Msgf("error setting system proxy: %s", err)
		}
		defer func() {
			if err := util.UnsetOsProxy(); err != nil {
				logger.Fatal().Msgf("error unsetting system proxy: %s", err)
			}
		}()
	}

	pxy := proxy.New(config)
	go pxy.Start(context.Background())

	waitForShutdown()
}

// waitForShutdown listens for OS signals and blocks until one is received.
func waitForShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)

	<-sigs
}
