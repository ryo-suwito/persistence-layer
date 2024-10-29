package utils

import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "os"
)

func InitLogger() {
    log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func LogInfo(message string, fields map[string]interface{}) {
    event := log.Info()
    for k, v := range fields {
        event = event.Interface(k, v)
    }
    event.Msg(message)
}

func LogError(err error, fields map[string]interface{}) {
    event := log.Error().Err(err)
    for k, v := range fields {
        event = event.Interface(k, v)
    }
    event.Msg("Error occurred")
}
