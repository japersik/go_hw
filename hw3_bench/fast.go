package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const fPath string = "./data/users.txt"

type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}
func FastSearch(out io.Writer) {
	file, err := os.Open(fPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := bufio.NewScanner(file)
	if err != nil {
		panic(err)
	}

	seenBrowsers := make(map[string]bool,256)

	var users []User

	for ;reader.Scan(); {
		line := reader.Bytes()
		user := User{}
		user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}
	fmt.Fprintln(out, "found users:")
	for i, user := range users {
		isAndroid := false
		isMSIE := false
		for  _, browser := range user.Browsers {
			if strings.Contains(browser,"Android") {
				isAndroid = true
				seenBrowsers[browser] = true
			}else if strings.Contains(browser,"MSIE") {
				isMSIE = true
				seenBrowsers[browser] = true
			}
		}

		if isAndroid && isMSIE {
			email := strings.Replace(user.Email,"@"," [at] ",1)
			fmt.Fprintf(out,"[%d] %s <%s>\n", i, user.Name, email)
		}
	}
	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
