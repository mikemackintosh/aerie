package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/mikemackintosh/aerie"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	// flagListPort is supplied to ListenAndServe.
	flagListenPort string

	// flagEnv is the environment file.
	flagEnv string

	// Default log format
	logFormat = `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
		`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
		`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n"
)

// init
func init() {
	flag.StringVar(&flagListenPort, "p", ":8080", "HTTP Listening IP:Port")
	flag.StringVar(&flagEnv, "env", ".env", "The name of the env file to load")
}

func main() {
	var err error
	err = aerie.GetReportEvents("all", "user_accounts")
	if err != nil {
		log.Fatalf("error starting watch: %s", err)
	}

	err = aerie.StartWatching("all", "user_accounts", []string{"email_forwarding_out_of_domain", "recovery_phone_edit"})
	if err != nil {
		log.Fatalf("error starting watch: %s", err)
	}

	err = aerie.StartWatching("all", "login", []string{"login_success", "account_disabled_spamming", "login_verification", "recovery_phone_edit"})
	if err != nil {
		log.Fatalf("error starting watch: %s", err)
	}

	err = aerie.StartWatching("all", "admin", []string{"DOWNLOAD_USERLIST_CSV", "CHANGE_USER_CUSTOM_FIELD"})
	if err != nil {
		log.Fatalf("error starting watch: %s", err)
	}

	watchDriveEvents := []string{"create", "edit"}
	err = aerie.StartWatching("all", "drive", watchDriveEvents)
	if err != nil {
		log.Fatalf("error starting watch: %s", err)
	}

	// Crteate the new echo router
	e := echo.New()
	e.HideBanner = true
	e.Debug = true

	// Configure the middlewares
	e.Use(middlewareServerHeader)
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           logFormat,
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}))

	// Default Google Workspace Push Notification route handler
	e.POST("/google/workspace/notification", aerie.HandlerWorkspaceNotifier)

	// Print the configured routes for debugging.
	for _, r := range e.Routes() {
		fmt.Printf("- %s\t%s\t-> %s\n", r.Method, r.Path, r.Name)
	}

	// Start listening on HTTP
	fmt.Printf("Listening HTTP on port: %s\n", flagListenPort)
	if err := e.Start(flagListenPort); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// echoWrap will wrap a standard-lib http handler with the echo.HandlerFunc.
func echoWrap(f func(http.ResponseWriter, *http.Request)) echo.HandlerFunc {
	return echo.WrapHandler(http.HandlerFunc(f))
}

// middlewareServerHeader middleware adds a `Server` header to the response.
func middlewareServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, "Aerie/0.2")
		return next(c)
	}
}
