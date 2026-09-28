package main

import (
	"math/rand/v2"
	"os"
	"strings"
)

const LOWERCASE_ALPHABETS = "abcdefghijklmnopqrstuvwxyz"
const UPPERCASE_ALPHABETS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const DIGITS = "0123456789"
const SYMBOLS = "!@#$%^&*()-_=+[]{}|;:,.<>?/"

func getRandomByte(charType int) byte {
	switch charType {
	case 1:
		index := randomInt(0, 25)
		return LOWERCASE_ALPHABETS[index]
	case 2:
		index := randomInt(0, 25)
		return UPPERCASE_ALPHABETS[index]
	case 3:
		index := randomInt(0, len(DIGITS)-1)
		return DIGITS[index]
	case 4:
		index := randomInt(0, len(SYMBOLS)-1)
		return SYMBOLS[index]
	}
	return 0
}
func randomInt(min int, max int) int {
	return rand.IntN(max-min+1) + min
}

func GeneratePassword(length int) string {
	var output string
	for range length {
		style := randomInt(1, 4)
		output += string(getRandomByte(style))
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
	content, err := os.ReadFile(".data/word-list.txt")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(content)), "\n"), nil
}
