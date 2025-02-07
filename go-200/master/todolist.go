package main

import (
	"encoding/base64"
	"fmt"
	"github.com/sfeir-open-source/sfeir-school-go/dao"
	"github.com/sfeir-open-source/sfeir-school-go/utils"
	"github.com/sfeir-open-source/sfeir-school-go/web"
	logger "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	// Version is the version of the software
	Version string
	// BuildStmp is the build date
	BuildStmp string
	// GitHash is the git build hash
	GitHash string

	port               = 8020
	logLevel           = "warning"
	db                 = ""
	dbType             = dao.DAOMockStr
	migrationPath      = "migration"
	logFormat          = utils.TextFormatter
	statisticsDuration = 20 * time.Second

	header, _ = base64.StdEncoding.DecodeString(
		"ICAgICAgICAgLF8tLS1+fn5+fi0tLS0uXyAgICAgICAgIAogIF8sLF8sKl5fX19fICAgICAgX19fX19gYCpnKlwiKi" +
			"wgCiAvIF9fLyAvJyAgICAgXi4gIC8gICAgICBcIF5AcSAgIGYgClsgIEBmIHwgQCkpICAgIHwgIHwgQCkpICAgbCA" +
			"gMCBfLyAgCiBcYC8gICBcfl9fX18gLyBfXyBcX19fX18vICAgIFwgICAKICB8ICAgICAgICAgICBfbF9fbF8gICAg" +
			"ICAgICAgIEkgICAKICB9ICAgICAgICAgIFtfX19fX19dICAgICAgICAgICBJICAKICBdICAgICAgICAgICAgfCB8I" +
			"HwgICAgICAgICAgICB8ICAKICBdICAgICAgICAgICAgIH4gfiAgICAgICAgICAgICB8ICAKICB8ICAgICAgICAgIC" +
			"AgICAgICAgICAgICAgICAgIHwgICAKICAgfCAgICAgICAgICAgICAgICAgICAgICAgICAgIHwg")
)

// installExit manages ctrl-c stop signals
func exitHandler() {
	// manage ctrl-c exit waiting for https://github.com/urfave/cli/issues/945 to be released
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		os.Exit(0)
	}()
}

func main() {
	// new app
	app := cli.NewApp()
	app.Name = utils.AppName
	app.Usage = "todolist service launcher"

	timeStmp, err := strconv.Atoi(BuildStmp)
	if err != nil {
		timeStmp = 0
	}
	app.Version = Version + ", build on " + time.Unix(int64(timeStmp), 0).String() + ", git hash " + GitHash
	app.Authors = []*cli.Author{{Name: "sfr"}}
	app.Copyright = "Sfeir " + strconv.Itoa(time.Now().Year())

	// command line flags
	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Value:       port,
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "Set the listening port of the webserver",
			Destination: &port,
			EnvVars:     []string{"APP_PORT"},
		},
		&cli.StringFlag{
			Value:       db,
			Name:        "db",
			Aliases:     []string{"d"},
			Usage:       "Set the database connection string (mongodb or postgresql)",
			Destination: &db,
		},
		&cli.StringFlag{
			Value:       dbType,
			Name:        "dbt",
			Aliases:     []string{"dt"},
			Usage:       "Set the database type to use for the service (mongodb, postgresql or mock)",
			Destination: &dbType,
		},
		&cli.StringFlag{
			Value:       migrationPath,
			Name:        "mp",
			Aliases:     []string{"m"},
			Usage:       "Set the database migration folder path",
			Destination: &migrationPath,
		},
		&cli.StringFlag{
			Value:       logLevel,
			Name:        "logl",
			Aliases:     []string{"l"},
			Usage:       "Set the output log level (debug, info, warning, error)",
			Destination: &logLevel,
		},
		&cli.StringFlag{
			Value:       logFormat,
			Name:        "logf",
			Aliases:     []string{"f"},
			Usage:       "Set the log formatter (logstash or text)",
			Destination: &logFormat,
		},
		&cli.DurationFlag{
			Value:       statisticsDuration,
			Name:        "statd",
			Aliases:     []string{"s"},
			Usage:       "Set the statistics accumulation duration (ex : 1h, 2h30m, 30s, 300ms)",
			Destination: &statisticsDuration,
		},
	}

	// main action
	// sub action are also possible
	app.Action = func(c *cli.Context) error {
		// print header
		fmt.Println(string(header))

		fmt.Print("* --------------------------------------------------- *\n")
		fmt.Printf("|   port                    : %d\n", port)
		fmt.Printf("|   db                      : %s\n", db)
		fmt.Printf("|   dbt                     : %s\n", dbType)
		fmt.Printf("|   mp                      : %s\n", migrationPath)
		fmt.Printf("|   logger level            : %s\n", logLevel)
		fmt.Printf("|   logger format           : %s\n", logFormat)
		fmt.Printf("|   statistic duration(s)   : %0.f\n", statisticsDuration.Seconds())
		fmt.Print("* --------------------------------------------------- *\n")

		// init log options from command line params
		err := utils.InitLog(logLevel, logFormat)
		if err != nil {
			logger.WithField("error", err).Warn("error setting log level, using debug as default")
		}

		// parse the database type
		dbt, err := dao.ParseDBType(dbType)
		if err != nil {
			return err
		}

		// build the web server
		webServer, err := web.BuildWebServer(db, migrationPath, dbt, statisticsDuration)

		if err != nil {
			return err
		}

		// serve
		webServer.Run(":" + strconv.Itoa(port))

		return nil
	}

	// install exit ctrl-c
	exitHandler()

	// run the app
	err = app.Run(os.Args)
	if err != nil {
		logger.Fatalf("Run error %q\n", err)
	}

}
