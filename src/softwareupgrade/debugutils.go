package softwareupgrade

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type (
	//DebugLog specifies where to send debugs to, and whether to also print the debug log to the console
	DebugLog struct {
		PrintDebug   bool
		PrintConsole bool
		file         *os.File
	}
)

func init() {
	log.SetOutput(ioutil.Discard) // prevent double output
}

// Debug decides whether to send the log to the debug log file
func (d *DebugLog) Debug(format string, args ...interface{}) {
	if d.PrintDebug {
		d.Print(format, args...)
	}
}

// Debugf decides whether to send the log to the debug log file
func (d *DebugLog) Debugf(format string, args ...interface{}) {
	if d.PrintDebug {
		d.Printf(format, args...)
	}
}

// Debugln adds a newline to the debug string
func (d *DebugLog) Debugln(format string, args ...interface{}) {
	if d.PrintDebug {
		d.Print(format+"\n", args...)
	}
}

// EnablePrintConsole sets the PrintConsole flag
func (d *DebugLog) EnablePrintConsole() {
	d.PrintConsole = true
}

// Print decides whether the debug log is sent to the console, or not, and also logs it to the debug log
func (d *DebugLog) Print(format string, args ...interface{}) {
	if d.PrintConsole {
		fmt.Printf(format, args...)
	}
	log.Printf(format, args...)
}

// Printf prints the specified debug log
func (d *DebugLog) Printf(format string, args ...interface{}) {
	log := fmt.Sprintf(format, args...)
	d.Print(log)
}

// Println adds a newline to the specified debug log
func (d *DebugLog) Println(msg string, args ...interface{}) {
	d.Printf(msg+"\n", args...)
}

// SetOutput changes the output writer for the log
func (d *DebugLog) SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// EnableDebugLog changes the log output to a new file specified by the given filename
func (d *DebugLog) EnableDebugLog(LogFilename string) error {
	if LogFilename == "" {
		return errors.New("Log filename is empty")
	}
	var err error
	expandedLogPath, err := Expand(LogFilename)
	if err != nil {
		return err
	}
	if !FileExists(expandedLogPath) {
		d.file, err = os.Create(expandedLogPath)
	} else {
		d.file, err = os.OpenFile(expandedLogPath, os.O_APPEND|os.O_WRONLY, 0)
	}
	if err == nil {
		log.SetOutput(d.file)
	}
	return err
}

// CloseDebugLog flushes the debug log and closes it.
func (d *DebugLog) CloseDebugLog() {
	if d.file == nil {
		return
	}
	d.file.Sync()
	d.file.Close()
	d.file = nil
}
