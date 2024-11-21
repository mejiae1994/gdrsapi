package logger

import (
	"log"
	"os"
)

type AppLogger struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func NewAppLogger() *AppLogger {
	infoLog := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return &AppLogger{
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}
}
