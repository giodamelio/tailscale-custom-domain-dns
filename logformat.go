package main

import (
	"fmt"
	"os"

	"github.com/logrusorgru/aurora/v4"
	"github.com/rs/zerolog"
)

func createFormatter() zerolog.ConsoleWriter {
	formatter := zerolog.ConsoleWriter{
		Out: os.Stdout,
		// Disable timestamp
		FormatTimestamp: func(i interface{}) string {
			return ""
		},
		FormatLevel: func(i interface{}) string {
			if stringLevel, ok := i.(string); ok {
				switch stringLevel {
				case zerolog.LevelTraceValue:
					return fmt.Sprintf("%s", aurora.Magenta("TRACE"))
				case zerolog.LevelDebugValue:
					return fmt.Sprintf("%s", aurora.Yellow("DEBUG"))
				case zerolog.LevelInfoValue:
					return fmt.Sprintf("%s", aurora.Green("INFO "))
				case zerolog.LevelWarnValue:
					return fmt.Sprintf("%s", aurora.Red("WARN "))
				case zerolog.LevelErrorValue:
					return fmt.Sprintf("%s", aurora.Red("ERROR").Bold())
				case zerolog.LevelFatalValue:
					return fmt.Sprintf("%s", aurora.Red("FATAL").Bold())
				case zerolog.LevelPanicValue:
					return fmt.Sprintf("%s", aurora.Red("PANIC").Bold())
				default:
					return fmt.Sprintf("%s", i)
				}
			} else {
				return fmt.Sprintf("%s", i)
			}
		},
		FormatFieldName: func(i interface{}) string {
			return ""
		},
		FormatFieldValue: func(i interface{}) string {
			return ""
		},
	}

	return formatter
}

func createTraceFormatter() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out: os.Stdout,
	}
}
