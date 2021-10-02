package main

import (
	"io"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
<<<<<<< HEAD
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

	fmt.Fprintln(out, "found users:")
	for i:=0;reader.Scan();i++ {
		line := reader.Bytes()
		user := User{}
		user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}
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
=======
	SlowSearch(out)
}
>>>>>>> parent of 6fed1d4 (add hw3)
