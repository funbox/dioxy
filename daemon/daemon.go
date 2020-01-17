package daemon

import (
	"os"
	"runtime"
	"strings"
	"time"

	"pkg.re/essentialkaos/ek.v10/fmtc"
	"pkg.re/essentialkaos/ek.v10/fsutil"
	"pkg.re/essentialkaos/ek.v10/knf"
	"pkg.re/essentialkaos/ek.v10/log"
	"pkg.re/essentialkaos/ek.v10/options"
	"pkg.re/essentialkaos/ek.v10/pid"
	"pkg.re/essentialkaos/ek.v10/signal"
	"pkg.re/essentialkaos/ek.v10/usage"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic info
const (
	APP  = "Dioxy"
	VER  = "1.0.0"
	DESC = "Aggregating proxy for MQTT broker in JSON format"
)

// Options
const (
	OPT_CONFIG   = "c:config"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VERSION  = "v:version"
)

// Configuration file props
const (
	HTTP_IP              = "http:ip"
	HTTP_PORT            = "http:port"
	MQTT_IP              = "mqtt:ip"
	MQTT_PORT            = "mqtt:port"
	MQTT_USER            = "mqtt:user"
	MQTT_PASSWORD        = "mqtt:password"
	MQTT_TOPIC           = "mqtt:topic"
	STORE_TTL            = "store:ttl"
	STORE_CLEAN_INTERVAL = "store:clean-interval"
	LOG_DIR              = "log:dir"
	LOG_FILE             = "log:file"
	LOG_PERMS            = "log:perms"
	LOG_LEVEL            = "log:level"
)

// Pid info
const (
	PID_DIR  = "/var/run/dioxy"
	PID_FILE = "dioxy.pid"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options map
var optMap = options.Map{
	OPT_CONFIG:   {Value: "/etc/dioxy.knf"},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL, Alias: "u:usage"},
	OPT_VERSION:  {Type: options.BOOL, Alias: "ver"},
}

// ////////////////////////////////////////////////////////////////////////////////// //

func Init() {
	runtime.GOMAXPROCS(8)

	_, errs := options.Parse(optMap)

	if len(errs) != 0 {
		for _, err := range errs {
			printError(err.Error())
		}

		os.Exit(1)
	}

	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}

	if options.GetB(OPT_VERSION) {
		showAbout()
		return
	}

	if options.GetB(OPT_HELP) {
		showUsage()
		return
	}

	loadConfig()
	//validateConfig()
	registerSignalHandlers()
	setupLogger()
	createPidFile()

	log.Aux(strings.Repeat("-", 88))
	log.Aux("%s %s starting...", APP, VER)

	start()
}

// loadConfig read and parse configuration file
func loadConfig() {
	err := knf.Global(options.GetS(OPT_CONFIG))

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// validateConfig validate configuration file values
func validateConfig() {
	var permsChecker = func(config *knf.Config, prop string, value interface{}) error {
		if !fsutil.CheckPerms(value.(string), config.GetS(prop)) {
			switch value.(string) {
			case "DW":
				return fmtc.Errorf("Property %s must be path to writable directory", prop)
			case "DX":
				return fmtc.Errorf("Property %s must be path to executable directory", prop)
			}
		}

		return nil
	}

	errs := knf.Validate([]*knf.Validator{
		{MQTT_IP, knf.Empty, nil},
		{MQTT_PORT, knf.Empty, nil},
		{MQTT_PORT, knf.Less, 1024},
		{MQTT_PORT, knf.Greater, 65535},
		{MQTT_USER, knf.Empty, nil},
		{MQTT_PASSWORD, knf.Empty, nil},
		{MQTT_TOPIC, knf.Empty, nil},

		{STORE_TTL, knf.Empty, nil},
		{STORE_TTL, knf.Less, 1},
		{STORE_CLEAN_INTERVAL, knf.Empty, 0},
		{STORE_CLEAN_INTERVAL, knf.Less, 1},

		{LOG_DIR, knf.Empty, nil},
		{LOG_FILE, knf.Empty, nil},
		{HTTP_PORT, knf.Empty, nil},

		{HTTP_PORT, knf.Less, 1024},
		{HTTP_PORT, knf.Greater, 65535},

		{LOG_DIR, permsChecker, "DW"},
		{LOG_DIR, permsChecker, "DX"},
		{LOG_LEVEL, knf.NotContains, []string{"debug", "info", "warn", "error", "crit"}},
	})

	if len(errs) != 0 {
		printError("Error while configuration file validation:")

		for _, err := range errs {
			printError("  %v", err)
		}

		os.Exit(1)
	}
}

// registerSignalHandlers register signal handlers
func registerSignalHandlers() {
	signal.Handlers{
		signal.TERM: termSignalHandler,
		signal.INT:  intSignalHandler,
		signal.HUP:  hupSignalHandler,
	}.TrackAsync()
}

// setupLogger setup logger
func setupLogger() {
	err := log.Set(knf.GetS(LOG_FILE), knf.GetM(LOG_PERMS, 644))

	if err != nil {
		printErrorAndExit(err.Error())
	}

	err = log.MinLevel(knf.GetS(LOG_LEVEL))

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// createPidFile create PID file
func createPidFile() {
	pid.Dir = PID_DIR

	err := pid.Create(PID_FILE)

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// start start service
func start() {
	err := startObserver(
		knf.GetS(MQTT_IP),
		knf.GetS(MQTT_PORT),
		knf.GetS(MQTT_USER),
		knf.GetS(MQTT_PASSWORD),
		knf.GetS(MQTT_TOPIC),
		knf.GetI(STORE_TTL),
	)

	if err != nil {
		log.Crit(err.Error())
		shutdown(1)
	}

	go storeJanitor()

	err = startHTTPServer(
		knf.GetS(HTTP_IP),
		knf.GetS(HTTP_PORT),
	)

	if err != nil {
		log.Crit(err.Error())
		shutdown(1)
	}

	shutdown(0)
}

// storeJanitor periodically cleans store
func storeJanitor() {
	cleanInterval := time.Duration(knf.GetI(STORE_CLEAN_INTERVAL)) * time.Second
	for {
		time.Sleep(cleanInterval)

		if datastore != nil {
			log.Debug("Cleaning datastore...")
			datastore.Clean()
		}
	}
}

// INT signal handler
func intSignalHandler() {
	log.Aux("Received INT signal, shutdown...")
	shutdown(0)
}

// TERM signal handler
func termSignalHandler() {
	log.Aux("Received TERM signal, shutdown...")
	shutdown(0)
}

// HUP signal handler
func hupSignalHandler() {
	log.Info("Received HUP signal, log will be reopened...")
	log.Reopen()
	log.Info("Log reopened by HUP signal")
}

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
}

// printError prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{y}"+f+"{!}\n", a...)
}

// printErrorAndExit print error mesage and exit with exit code 1
func printErrorAndExit(f string, a ...interface{}) {
	printError(f, a...)
	os.Exit(1)
}

// shutdown stop deamon
func shutdown(code int) {
	pid.Remove(PID_FILE)
	os.Exit(code)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func showUsage() {
	info := usage.NewInfo()

	info.AddOption(OPT_CONFIG, "Path to configuraion file", "file")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VERSION, "Show version")

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2007,
		Owner:   "Gleb Goncharov",
		License: "MIT License",
	}

	about.Render()
}
