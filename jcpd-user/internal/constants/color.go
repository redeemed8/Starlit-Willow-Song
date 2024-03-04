package constants

import (
	"github.com/fatih/color"
)

func Err(msg string) string {
	return color.HiRedString(msg)
}

func Hint(msg string) string {
	return color.HiYellowString(msg)
}

func Info(msg string) string {
	return color.HiGreenString(msg)
}
