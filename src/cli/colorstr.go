package main

import "fmt"

func greenStr(str string) string {
	return fmt.Sprintf("\033[32m%s\033[0m", str)
}

func yellowStr(str string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", str)
}

func redStr(str string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", str)
}
