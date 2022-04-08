package server

import (
	"context"
	"fmt"
	. "godopi/internal/app/configs"
	. "godopi/internal/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Init() {
	router := NewRouter()
	serverAddress := Config().GetString(SERVER_ADDRESS)

	server := &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	gracefullyClosedChannel := make(chan struct{})

	go func() {
		shutdownChannel := make(chan os.Signal, 1)

		signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		receivedSignal := <-shutdownChannel

		Logger().Info(fmt.Sprintf("Received an os signal: %s", receivedSignal.String()))

		// We received an expected os signal, shut down.
		if err := server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			Logger().Error(fmt.Sprintf("Error received at Godopi server Shutdown: %v", err))
		}

		close(gracefullyClosedChannel)
	}()

	Logger().Info(fmt.Sprintf("Godopi server listening at: %s", serverAddress))

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		Logger().Fatal(fmt.Sprintf("Error received at Godopi server ListenAndServe: %v", err))
	}

	<-gracefullyClosedChannel

	Logger().Info("Godopi server gracefully closed.")
}
