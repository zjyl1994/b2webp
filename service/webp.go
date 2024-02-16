package service

import (
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
)

var singleFlight singleflight.Group

func Convert2Webp(infile, outfile string) (err error) {
	infile, err = filepath.Abs(infile)
	if err != nil {
		return err
	}
	outfile, err = filepath.Abs(outfile)
	if err != nil {
		return err
	}

	_, err, _ = singleFlight.Do(infile, func() (interface{}, error) {
		t := time.Now()
		err := exec.Command("cwebp", "-quiet", "-o", outfile, infile).Run()
		if err != nil {
			logrus.Errorln("CWEBP", infile, err.Error())
		} else {
			logrus.Infoln("CWEBP", infile, time.Since(t))
		}
		return nil, err
	})

	return err
}
