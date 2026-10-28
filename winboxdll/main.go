package main

/*
  Define DLL ( NewLazyDLL Ex: User32.dll)
  create process via (Ex: NewProc() )
  Call function process Call(args...)
*/
import (
	"fmt"
	"golang.org/x/sys/windows"
	"unsafe"
)

func main() {
	text := windows.StringToUTF16Ptr(`This is not a fire drill, it's the real thing`)
	caption := windows.StringToUTF16Ptr(`MB dll`)

	user32dll := windows.NewLazyDLL("user32.dll")

	procMessageBox := user32dll.NewProc("MessageBoxW")

	// windows.MessageBox(windows.HWND(0), t *uint16, c *uint16, windows.MB_OK)
	r1, _, err := procMessageBox.Call(uintptr(0), uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(caption)), uintptr(windows.MB_OK))
	if r1 == 0 {
		fmt.Print(err)
	}

}
