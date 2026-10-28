package main

import (
	"fmt"
	"golang.org/x/sys/windows"
)

func main() {
	text, _ := windows.UTF16PtrFromString("This is not a fire drill, it's the real thing")
	caption, _ := windows.UTF16PtrFromString("Testing")

	_, err := windows.MessageBox(0, text, caption, windows.MB_OK)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%T\n", text)
}
