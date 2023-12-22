
# Go Process Memory Reader

A Go program that allows you to list processes running on a Windows system and snapshot and print the memory content from a specific memory address.

## Usage

### List Processes

To list all running processes:

```bash
$ go-process-memory-reader
```

### Snapshot and Print Memory Content

To snapshot and print the memory content from a specific process and memory address:

```bash
$ go-process-memory-reader <PID> [MemoryAddress]
```

- `<PID>`: Process ID of the target process.
- `[MemoryAddress]`: Optional memory address within the target process. If provided, the program will print the memory content at that address.

## Examples

List all running processes:

```bash
$ go-process-memory-reader
```

Snapshot and print memory content from a specific process (e.g., PID 1234):

```bash
$ go-process-memory-reader 1234
```

Snapshot and print memory content from a specific process at a memory address (e.g., PID 1234, MemoryAddress 0x1000):

```bash
$ go-process-memory-reader 1234 0x1000
```

## Requirements

- Go (Golang)
- Windows operating system

## Building

To build the program, run:

```bash
$ go build
```

## Dependencies

- None

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- The program utilizes Windows API functions via the `kernel32.dll` library to interact with processes and memory.
