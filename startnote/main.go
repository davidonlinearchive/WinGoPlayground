package main

import (
	"fmt"
	"golang.org/x/sys/windows"
)

/*
Native C++ syntax
BOOL CreateProcessW(
	  [in, optional]      LPCWSTR               lpApplicationName,
	    [in, out, optional] LPWSTR                lpCommandLine,
	      [in, optional]      LPSECURITY_ATTRIBUTES lpProcessAttributes,
	        [in, optional]      LPSECURITY_ATTRIBUTES lpThreadAttributes,
	          [in]                BOOL                  bInheritHandles,
	            [in]                DWORD                 dwCreationFlags,
	              [in, optional]      LPVOID                lpEnvironment,
	                [in, optional]      LPCWSTR               lpCurrentDirectory,
	                  [in]                LPSTARTUPINFOW        lpStartupInfo,
	                    [out]               LPPROCESS_INFORMATION lpProcessInformation
	                    );
)
*/

/*
func CreateProcess(appName *uint16, commandLine *uint16,

	procSecurity *SecurityAttributes, threadSecurity *SecurityAttributes,
	inheritHandles bool, creationFlags uint32,
	env *uint16, currentDir *uint16,
	startupInfo *StartupInfo, outProcInfo *ProcessInformation)
	 (err error)
*/
func main() {
	appName := windows.StringToUTF16Ptr(`C:\\Windows\\system32\\notepad.exe`)

	var si windows.StartupInfo
	var pi windows.ProcessInformation
	err := windows.CreateProcess(
		appName,
		nil,
		nil,
		nil,
		false,
		0,
		nil,
		nil,
		&si,
		&pi)

	if err != nil {
		fmt.Printf("%v\n", err)
	}

}
