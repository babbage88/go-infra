package main

import (
	"fmt"
	"io"
	"log/slog"

	customlogger "git.trahan.dev/go-infra/utils"
)

func main() {

	fmt.Print("test")

	config := customlogger.NewCustomLogger()

	clog := customlogger.SetupLogger(config)

	defer func() {
		if file, ok := clog.Handler().(io.Closer); ok {
			file.Close()
		}
	}()

	clog.Info("Starting Infra tasks", slog.String("Action", "DNS Updated"))

}
