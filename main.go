/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import "github.com/threeport/tptctl/cmd"

//go:generate bash -c "./get_version.sh"
func main() {
	cmd.Execute()
}
