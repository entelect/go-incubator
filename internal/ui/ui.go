package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Selection prompts the user to select from a list of items
func Selection(prompt string, items []string) string {
	fmt.Println(prompt)
	for i, v := range items {
		fmt.Printf("%d) %s\n", i+1, v)
	}
	fmt.Println()

	var selection int64
	valid := false
	for !valid {
		var err error
		s := GetValue(fmt.Sprintf("Make your selection (%d - %d)", 1, len(items)))
		selection, err = strconv.ParseInt(s, 10, 32)
		if err != nil || selection < 1 || int(selection) > len(items) {
			fmt.Printf("invalid selection (%s)\n", s)
		} else {
			valid = true
		}
	}

	return items[selection-1]
}

// GetValue prompts the user for a single string value
func GetValue(prompt string) string {
	var str string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, prompt+" ")
		str, _ = r.ReadString('\n')
		if str != "" {
			break
		}
	}
	return strings.TrimSpace(str)
}
