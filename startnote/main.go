package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

func main() {
	appName := windows.StringToUTF16Ptr(`C:\\Windows\\system32\\notepad.exe`)

	var si windows.StartupInfo
	si.Cb = uint32(unsafe.Sizeof(si)) // Required: Initialize structure size

	var pi windows.ProcessInformation

	// Use CREATE_NEW_CONSOLE (0x00000010) to spawn in the active user session workspace
	err := windows.CreateProcess(
		appName,
		nil,
		nil,
		nil,
		false,
		windows.CREATE_NEW_CONSOLE,
		nil,
		nil,
		&si,
		&pi)

	if err != nil {
		fmt.Printf("CreateProcess failed: %v\n", err)
		return
	}

	fmt.Printf("PID: [%v]\n", pi.ProcessId)

	windows.CloseHandle(pi.Process)
	windows.CloseHandle(pi.Thread)
}
