package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/urfave/cli/v2"
)

const (
	exitCodeNotContains = 1
	exitCodeUsageError  = 2
)

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Println(cCtx.App.Version)
	}

	app := &cli.App{
		Name:    "postgresql-test-webapp",
		Version: Version(),
		Usage:   "a example Web app using PostgreSQL",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				Aliases: []string{"l"},
				Value:   ":8080",
				Usage:   "listen address (ex. :8080)",
			},
			&cli.StringFlag{
				Name:    "database",
				Aliases: []string{"d"},
				Usage:   `PostgreSQL connection string (ex. "postgres://username:password@localhost:5432/database_name")`,
				EnvVars: []string{"DATABASE_URL"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			return execMainCommand(cCtx.Context, cCtx.String("listen"), cCtx.String("database"))
		},
	}
	app.UsageText = fmt.Sprintf("%s [GLOBAL OPTIONS]", app.Name)
	app.OnUsageError = func(cCtx *cli.Context, err error, isSubcommand bool) error {
		return cli.Exit(err.Error(), exitCodeUsageError)
	}

	if err := run(app, os.Args); err != nil {
		cli.HandleExitCoder(err)
		fmt.Fprintf(app.ErrWriter, "\nError: %s\n", err)
		if strings.HasPrefix(err.Error(), "Required flag") {
			os.Exit(exitCodeUsageError)
		}
	}
}

func run(app *cli.App, args []string) error {
	signals := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		signals = append(signals, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()
	return app.RunContext(ctx, args)
}

const gracefulShutdownTimeout = time.Minute

func execMainCommand(ctx context.Context, listenAddr, databaseURL string) (err error) {
	myApp, err := newMyApp(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer myApp.closeDB()

	e := echo.New()
	myApp.addHandlers(e)
	s := http.Server{
		Addr:    listenAddr,
		Handler: e,
	}

	shutdownErrC := make(chan error)
	go func() {
		<-ctx.Done()
		log.Print("got signal, start shutdown")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			shutdownErrC <- err
		}
		close(shutdownErrC)
	}()

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	if err, ok := <-shutdownErrC; ok {
		return err
	}
	return nil
}

func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return info.Main.Version
}
