package main

import "fmt"

func main() {
	length := 15
	for range 5 {

		fmt.Println(GeneratePassword(length))
		fmt.Println()
	}

	for range 5 {
		passphrase, _ := GeneratePassphrase()
		fmt.Println(passphrase)
		fmt.Println()
	}

}
