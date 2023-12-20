package library

import (
	"fmt"
	"log"
	"s2dnglobby/config"
)

// TODO log to file
// see https://stackoverflow.com/questions/18361750/correct-approach-to-global-logging

const info 	= "INFO "
const debug = "DEBUG"
const err 	= "ERROR"

type logger struct {
	name   string
	debug  bool
	logger *log.Logger
}

func (l *logger) Info(v ...any) {
	l.logger.Print(l.name, info, v)
}
func (l *logger) Infoln(v ...any) {
	l.logger.Println(l.name, info, v)
}
func (l *logger) Infof(format string, v ...any) {
	l.logger.Printf(format, l.name, info, v)
}

func (l *logger) Debug(v ...any) {
	if l.debug {
		l.logger.Print(l.name, debug, v)
	}
}
func (l *logger) Debugln(v ...any) {
	if l.debug {
		l.logger.Println(l.name, debug, v)
	}
}
func (l *logger) Debugf(format string, v ...any) {
	if l.debug {
		str := fmt.Sprintf(format, v...)
		l.logger.Print(l.name, debug, str)
	}
}

func (l *logger) Error(v ...any) {
	l.logger.Print(l.name, err, v)
}
func (l *logger) Errorln(v ...any) {
	l.logger.Println(l.name, err, v)
}
func (l *logger) Errorf(format string, v ...any) {
	l.logger.Printf(format, l.name, err, v)
}

func (l *logger) Panic(v ...any) {
	l.logger.Panic(l.name, v)
}
func (l *logger) Panicln(v ...any) {
	l.logger.Panicln(l.name, v)
}
func (l *logger) Panicf(format string, v ...any) {
	l.logger.Panicf(format, l.name, v)
}


func GetLogger(name string) *logger {
	dfl := log.Default()

	return &logger{
		name:   fmt.Sprintf("[%s]", name),
		debug:  config.DEBUGGING,
		logger: dfl,
	}
}
