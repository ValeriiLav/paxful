// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeebo/errs"

	"paxful"
	"paxful/internal/logger/zaplog"
	"paxful/paxfuldb"
)

var Error = errs.Class("paxful payments CLI error")

// Config is the global configuration to interact with paxful payment service through CLI.
type Config struct {
	DatabaseURL   string `json:"databaseUrl"`
	paxful.Config `json:"config"`
}

// commands
var (
	// payments root cmd.
	rootCmd = &cobra.Command{
		Use:   "payments",
		Short: "CLI for interacting with paxful payment service",
	}

	// payments setup cmd.
	setupCmd = &cobra.Command{
		Use:         "setup",
		Short:       "setups the program config, creates database",
		RunE:        cmdSetup,
		Annotations: map[string]string{"type": "setup"},
	}
	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the program",
		RunE:        cmdRun,
		Annotations: map[string]string{"type": "run"},
	}
	runCfg   Config
	setupCfg Config

	defaultConfigDir = applicationDir("paxful")
)

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()
	log := zaplog.NewLog()

	runCfg, err = readConfig()
	if err != nil {
		log.Error("Could not read config from default place", Error.Wrap(err))
		return Error.Wrap(err)
	}

	db, err := paxfuldb.NewDatabase(runCfg.DatabaseURL)
	if err != nil {
		log.Error("Error starting master database on paxful payment service", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, db.Close())
	}()

	peer, err := paxful.NewPeer(log, db, runCfg.Config)
	if err != nil {
		log.Error("Error starting paxful payment service", Error.Wrap(err))
		return Error.Wrap(err)
	}

	runError := peer.Run(ctx)
	closeError := peer.Close()
	return Error.Wrap(errs.Combine(runError, closeError))
}

func cmdSetup(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()
	log := zaplog.NewLog()

	setupDir, err := filepath.Abs(defaultConfigDir)
	if err != nil {
		return Error.Wrap(err)
	}

	err = os.MkdirAll(setupDir, os.ModePerm)
	if err != nil {
		return Error.Wrap(err)
	}

	configFile, err := os.Create(path.Join(setupDir, "config.json"))
	if err != nil {
		log.Error("could not create config file", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, configFile.Close())
	}()

	jsonData, err := json.MarshalIndent(setupCfg, "", "    ")
	if err != nil {
		log.Error("could not marshal config", Error.Wrap(err))
		return Error.Wrap(err)
	}

	_, err = configFile.Write(jsonData)
	if err != nil {
		log.Error("could not write to config", Error.Wrap(err))
		return Error.Wrap(err)
	}

	conn, err := sql.Open("postgres", setupCfg.DatabaseURL)
	if err != nil {
		log.Error("could not connect to database server", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, conn.Close())
	}()

	createDBQuery := "CREATE DATABASE paxfuldb;"

	_, err = conn.ExecContext(ctx, createDBQuery)
	if err != nil {
		log.Error("can not create paxfuldb", Error.Wrap(err))
		return Error.Wrap(err)
	}

	createTableQuery :=
		`
		CREATE TABLE transactions (
			id            TEXT   NOT NULL,
			currency 	  TEXT   NOT NULL,
			amount        bigint NOT NULL,
			fee           bigint NOT NULL,
			fromAddress   TEXT   NOT NULL,
			toAddress     TEXT   NOT NULL,
			created_at    timestamp with time zone NOT NULL
		);
		`

	_, err = conn.ExecContext(ctx, createTableQuery)
	if err != nil {
		log.Error("can not create transactions table", Error.Wrap(err))
		return Error.Wrap(err)
	}

	return err
}

// TODO: below functions should be placed in another place and be refactored, but i'm facing real lack of time.

// applicationDir returns best base directory for specific OS.
func applicationDir(subdir ...string) string {
	for i := range subdir {
		if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
			subdir[i] = strings.Title(subdir[i])
		} else {
			subdir[i] = strings.ToLower(subdir[i])
		}
	}
	var appdir string
	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "windows":
		// Windows standards: https://msdn.microsoft.com/en-us/library/windows/apps/hh465094.aspx?f=255&MSPPError=-2147217396
		for _, env := range []string{"AppData", "AppDataLocal", "UserProfile", "Home"} {
			val := os.Getenv(env)
			if val != "" {
				appdir = val
				break
			}
		}
	case "darwin":
		// Mac standards: https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/MacOSXDirectories/MacOSXDirectories.html
		appdir = filepath.Join(home, "Library", "Application Support")
	case "linux":
		fallthrough
	default:
		// Linux standards: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
		appdir = os.Getenv("XDG_DATA_HOME")
		if appdir == "" && home != "" {
			appdir = filepath.Join(home, ".local", "share")
		}
	}
	return filepath.Join(append([]string{appdir}, subdir...)...)
}

// readConfig reads config from default config dir.
func readConfig() (config Config, err error) {
	configBytes, err := ioutil.ReadFile(path.Join(defaultConfigDir, "config.json"))
	if err != nil {
		return Config{}, err
	}

	return config, json.Unmarshal(configBytes, &config)
}
