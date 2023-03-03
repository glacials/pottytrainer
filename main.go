package main

import "twos.dev/pottytrainer/cmd"

var (
	version = "development"
)

func main() {
	cmd.Execute(version)
}
