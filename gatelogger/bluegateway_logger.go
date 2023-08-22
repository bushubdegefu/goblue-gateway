package gatelogger

import (
	"log"
	"os"
	"time"
)

func GateLoggerInfo(message string) {

	File, _ := os.OpenFile("gateway.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer File.Close()
	GateLogger := log.Logger{}
	GateLogger.SetOutput(File)
	GateLogger.Println(time.Now().String() + ": " + message)
}

func GateLoggerFatal(object interface{}) {
	File, _ := os.OpenFile("gateway.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer File.Close()
	GateLogger := log.Logger{}
	GateLogger.SetOutput(File)
	GateLogger.Fatal(object)

}
