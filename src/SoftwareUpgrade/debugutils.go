package softwareupgrade

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type DebugLog struct {
	PrintDebug bool
	file       *os.File
}

func (d *DebugLog) Debug(format string, args ...interface{}) {
	if d.PrintDebug {
		d.Print(format, args...)
	}
}

func (d *DebugLog) Debugf(format string, args ...interface{}) {
	if d.PrintDebug {
		d.Printf(format, args...)
	}
}

func (d *DebugLog) Print(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (d *DebugLog) Printf(format string, args ...interface{}) {
	log := fmt.Sprintf(format, args...)
	d.Print(log)
}

func (d *DebugLog) Println(msg string, args ...interface{}) {
	d.Printf(msg+"\n", args...)
}

func (d *DebugLog) SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func (d *DebugLog) EnableDebugLog(LogFilePtr *string) error {
	if LogFilePtr == nil || *LogFilePtr == "" {
		return errors.New("Log filename is empty")
	}
	var err error
	expandedLogPath, err := expand(*LogFilePtr)
	if err != nil {
		return err
	}
	d.file, err = os.Create(expandedLogPath)
	if err == nil {
		log.SetOutput(d.file)
	}
	return err
}

func (d *DebugLog) CloseDebugLog() {
	if d.file == nil {
		return
	}
	d.file.Sync()
	d.file.Close()
}
