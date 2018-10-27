package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	filevolConf = "/etc/docker/filevol-plugin"
	filevolHome = "/var/lib/docker-filevol-plugin"
	filevolLogs = "/var/log/docker-filevol-plugin.log"
)

var (
	flVersion *bool
	flDebug   *bool

	tfSize  = "208896"
	tfFStyp = "ext4"
	tfPath  = "/mnt/data/apps"
)

func init() {
	flVersion = flag.Bool("version", false, "Print version information and quit.")
	flDebug = flag.Bool("debug", false, "Enable debug logging.")
}

func main() {
	flag.Parse()

	if *flVersion {
		fmt.Fprint(os.Stdout, "Docker filevol plugin version: 1.0.0\n")
		return
	}

	if *flDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if _, err := os.Stat(filevolHome); err != nil {
		if !os.IsNotExist(err) {
			logrus.Fatal(err)
		}

		logrus.Debugf("Created home dir at %s.", filevolHome)
		if err := os.MkdirAll(filevolHome, 0700); err != nil {
			logrus.Fatal(err)
		}
	}

	var val []string
	fp, errp := os.Open(filevolConf)
	if errp == nil {
		scanner := bufio.NewScanner(fp)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "#") {
				continue
			} else if strings.Contains(line, "size=") {
				val = strings.Split(line, "=")
				tfSize = val[1]
			} else if strings.Contains(line, "path=") {
				val = strings.Split(line, "=")
				tfPath = val[1]
			} else if strings.Contains(line, "fstyp=") {
				val = strings.Split(line, "=")
				tfFStyp = val[1]
			}
		}
		fp.Close()
	}

	filevol, err := newDriver(tfSize, tfPath, tfFStyp, filevolHome)
	if err != nil {
		logrus.Fatalf("Error initializing fileVolDriver %v.", err)
	}

	h := volume.NewHandler(filevol)
	if err := h.ServeUnix("filevol", 0); err != nil {
		logrus.Fatal(err)
	}
}
