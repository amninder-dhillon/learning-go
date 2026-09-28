package main

import (
	"embed"
	"math/rand/v2"
	"strings"
)

type Options int

const (
	LowerAlpha Options = iota
	UpperAlpha
	Numbers
	Symbols
	All
)

var PasswordSpace = map[Options]string{
	LowerAlpha: "abcdefghijklmnopqrstuvwxyz",
	UpperAlpha: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	Numbers:    "0123456789",
	Symbols:    "!@#$%^&*()-_=+[]{}|;:,.<>?/",
	All:        "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/",
}

func (opt Options) String() string {
	return PasswordSpace[opt]
}

//go:embed .data/word-list.txt
var wordListFS embed.FS

func getRandomByte(passwordOption Options) byte {
	val, ok := PasswordSpace[passwordOption]
	if !ok {
		return 0
	}
	index := randomInt(0, len(val)-1)
	return PasswordSpace[passwordOption][index]
}
func randomInt(min int, max int) int {
	return rand.IntN(max-min+1) + min
}

func GeneratePassword(length int) string {
	var output string
	for range length {
		output += string(getRandomByte(All))
	}
	return output
}

func GeneratePasswordWithOptions(length int, pswdOption Options) string {
	var output string
	for range length {
		output += string(getRandomByte(pswdOption))
	}
	return output
}

func GeneratePassphrase() (string, error) {
	wordList, err := getWordList()
	if err != nil {
		return "", err
	}
	wordCount := 4
	separator := "-"
	passphrase := ""
	for i := range wordCount {
		passphrase += wordList[randomInt(0, len(wordList)-1)]
		if i != wordCount-1 {
			passphrase += separator
		}
	}
	return passphrase, nil
}

func getWordList() ([]string, error) {
	content, err := wordListFS.ReadFile(".data/word-list.txt")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(content)), "\n"), nil
}
