package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ignoreNames = []string{".git",".idea"}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

type fileStruct struct {
	name string
	childList []fileStruct
}

func (fs fileStruct) PrintChild() (result string) {
	for i, f := range fs.childList {
		result+= f.PrintStruct(i==len(fs.childList)-1)
	}
	return
}

func (fs fileStruct) PrintStruct(isLast bool) (result string) {
	if isLast{
		result+="└───"+fs.name+"\n"
	}else{
		result+="├───"+fs.name+"\n"
	}
	for i, f := range fs.childList {
		for _, s := range strings.Split(f.PrintStruct(i==len(fs.childList)-1),"\n") {
			if s != ""{
				if isLast{
					result+= "\t" + s + "\n"
				}else {
					result+= "│\t" + s + "\n"
				}
			}
		}
	}
	return
}

func (fs *fileStruct) addChild(str string)  {
	str = strings.TrimPrefix(str,string(os.PathSeparator))
	elements := strings.SplitN(str,string(os.PathSeparator),2)
		if len(fs.childList)==0 || fs.childList[len(fs.childList)-1].name != elements[0]  {
			child := fileStruct{name: elements[0], childList: make([]fileStruct, 0)}
			fs.childList = append(fs.childList, child)
		}
		if len(elements)>1 {
			fs.childList[len(fs.childList)-1].addChild(elements[1])
		}

}

func dirTree(writer io.Writer, filePath string, printFiles bool) error{
	fs := fileStruct{name: filePath,childList: make([]fileStruct,0)}
	var prefixes []string
	prefixes = append(prefixes, filePath)
	err := filepath.Walk(filePath, func(path string, f os.FileInfo, err error) error {
		if isIgnored(path) {
			return nil
		}else if !f.IsDir() &&!printFiles  {
			return nil
		}
		path = strings.TrimPrefix(path,filePath)
		if path == ""  {
			return nil
		}

		if !f.IsDir(){
			size:= int(f.Size())
			if size == 0 {
				path+= " (empty)"
			}else{
			path+= " (" + strconv.Itoa(size)+"b)"
		}
		}
		fs.addChild(path)
		return nil
	})
	fmt.Fprintln(writer,fs.PrintChild())
	return err
}

func isIgnored(path string)  bool{
	pathList := strings.Split(path,string(os.PathSeparator))[0]
	for _,val := range  ignoreNames {
		if pathList == val{
			return true
		}
	}
	return false
}
