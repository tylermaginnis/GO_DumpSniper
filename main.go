package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	TH32CS_SNAPPROCESS        = 0x00000002
	PROCESS_VM_READ           = 0x0010
	PROCESS_QUERY_INFORMATION = 0x0400
	MEM_COMMIT                = 0x00001000
	PAGE_READWRITE            = 0x04
	PAGE_READONLY             = 0x02
)

type PROCESSENTRY32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

func main() {
	args := os.Args[1:]

	if len(args) == 1 {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Invalid PID: %s\n", args[0])
			return
		}
		SnapshotMemory(pid)
		EnumerateProcessMemory(pid)
	} else if len(args) == 2 {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Invalid PID: %s\n", args[0])
			return
		}

		memoryAddress, err := strconv.ParseUint(args[1], 0, 64)
		if err != nil {
			fmt.Printf("Invalid Memory Address: %s\n", args[1])
			return
		}

		PrintMemoryValue(pid, uintptr(memoryAddress))
	} else {
		fmt.Println("Usage: program.exe <PID> [MemoryAddress]")
	}
}

func listProcesses() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	createToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	process32First := kernel32.NewProc("Process32FirstW")
	process32Next := kernel32.NewProc("Process32NextW")

	handle, _, _ := createToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if handle == 0 {
		fmt.Println("Error creating snapshot.")
		return
	}
	defer syscall.CloseHandle(syscall.Handle(handle))

	var entry PROCESSENTRY32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if ret, _, _ := process32First.Call(handle, uintptr(unsafe.Pointer(&entry))); ret == 0 {
		return
	}

	for {
		fmt.Printf("Process ID: %d, Executable Name: %s\n", entry.ProcessID, syscall.UTF16ToString(entry.ExeFile[:]))
		if ret, _, _ := process32Next.Call(handle, uintptr(unsafe.Pointer(&entry))); ret == 0 {
			break
		}
	}
}

func SnapshotMemory(pid int) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	openProcess := kernel32.NewProc("OpenProcess")
	virtualQueryEx := kernel32.NewProc("VirtualQueryEx")
	readProcessMemory := kernel32.NewProc("ReadProcessMemory")
	closeHandle := kernel32.NewProc("CloseHandle")

	handle, _, _ := openProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
	if handle == 0 {
		fmt.Println("Error opening process.")
		return
	}
	defer closeHandle.Call(handle)

	var mbi MEMORY_BASIC_INFORMATION
	var address uintptr

	for {
		// Query the memory region
		ret, _, _ := virtualQueryEx.Call(handle, address, uintptr(unsafe.Pointer(&mbi)), unsafe.Sizeof(mbi))
		if ret == 0 {
			break
		}

		// Read the memory region if it is committed and readable
		if mbi.State == MEM_COMMIT && (mbi.Protect&PAGE_READWRITE != 0 || mbi.Protect&PAGE_READONLY != 0) {
			bufferSize := mbi.RegionSize
			buffer := make([]byte, bufferSize)
			var bytesRead uint32

			ret, _, _ := readProcessMemory.Call(handle, mbi.BaseAddress, uintptr(unsafe.Pointer(&buffer[0])), uintptr(bufferSize), uintptr(unsafe.Pointer(&bytesRead)))
			if ret != 0 {
				fmt.Printf("Read %d bytes from memory address %#x\n", bytesRead, mbi.BaseAddress)
				// Process the buffer data as needed
			} else {
				fmt.Printf("Failed to read memory at address %#x\n", mbi.BaseAddress)
			}
		}

		address = uintptr(unsafe.Pointer(mbi.BaseAddress)) + uintptr(mbi.RegionSize)
	}
}

func EnumerateProcessMemory(pid int) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	openProcess := kernel32.NewProc("OpenProcess")
	virtualQueryEx := kernel32.NewProc("VirtualQueryEx")
	closeHandle := kernel32.NewProc("CloseHandle")

	handle, _, _ := openProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
	if handle == 0 {
		fmt.Println("Error opening process.")
		return
	}
	defer closeHandle.Call(handle)

	var mbi MEMORY_BASIC_INFORMATION
	var address uintptr

	for {
		ret, _, _ := virtualQueryEx.Call(handle, address, uintptr(unsafe.Pointer(&mbi)), unsafe.Sizeof(mbi))
		if ret == 0 {
			break
		}

		fmt.Printf("Memory Region: %#x, Size: %x\n", mbi.BaseAddress, mbi.RegionSize)

		address = uintptr(unsafe.Pointer(mbi.BaseAddress)) + uintptr(mbi.RegionSize)
	}
}

func PrintMemoryValue(pid int, address uintptr) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	openProcess := kernel32.NewProc("OpenProcess")
	readProcessMemory := kernel32.NewProc("ReadProcessMemory")
	closeHandle := kernel32.NewProc("CloseHandle")

	handle, _, _ := openProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
	if handle == 0 {
		fmt.Println("Error opening process.")
		return
	}
	defer closeHandle.Call(handle)

	var buffer [256]byte // You can adjust the buffer size as needed

	ret, _, _ := readProcessMemory.Call(handle, address, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0)
	if ret != 0 {
		fmt.Printf("Memory content at address %#x:\n", address)
		for i := 0; i < len(buffer); i++ {
			fmt.Printf("%02X ", buffer[i])
			if (i+1)%16 == 0 {
				fmt.Println()
			}
		}
	} else {
		fmt.Printf("Failed to read memory at address %#x\n", address)
	}
}
