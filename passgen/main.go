package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// if len(os.Args) < 2 {
	// 	fmt.Fprintln(os.Stderr, "Expected Usage: passgen <cmd> <flags>")
	// 	os.Exit(1)
	// }

	if len(os.Args) == 1 {
		for range 5 {
			fmt.Println(GeneratePasswordWithOptions(15, All))
			fmt.Println()
		}
		for range 5 {
			passphrase, _ := GeneratePassphrase()
			fmt.Println(passphrase)
			fmt.Println()
		}
		os.Exit(0)
	}
	switch os.Args[1] {
	case "password":
		passwordFlags := flag.NewFlagSet("password", flag.ExitOnError)
		lenPtr := passwordFlags.Int("len", 15, "Length of password")
		numPtr := passwordFlags.Int("num", 1, "Number of passwords to generate")

		passwordFlags.Parse(os.Args[2:])
		for range *numPtr {
			fmt.Println()
			fmt.Println(GeneratePasswordWithOptions(*lenPtr, All))
		}
	case "passphrase":
		passphrase, _ := GeneratePassphrase()
		fmt.Println(passphrase)

	}

}
