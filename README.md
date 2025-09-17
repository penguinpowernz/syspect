# Syspect: System Inspection and analysis Toolkit

A collection of tools for system analysis.

## Binaries

### fileanal
A comprehensive file analysis tool that:
- Searches for files in PATH
- Calculates MD5 and SHA256 hashes
- Identifies file details (type, size, permissions)
- Finds running processes using the file
- Checks network connections related to the file
- Links to VirusTotal for additional analysis

**Usage**: 
```
fileanal <file_path_or_filename>
```

Passing an absolute path will perform a full analysis, while a filename will search for the file in PATH and show any results.

### antibus

> [!WARNING]
> Work in progress

An ANSI code stripping utility that:
- Removes ANSI terminal color/formatting codes from files
- Can be used to sanitize log files or terminal output

**Usage**:
```
antibus <file1> [file2] ...
```

### cmdscan

> [!NOTE]
> Requires auditd to be installed and running and monitoring commands being executed

A Linux audit log command execution monitor that:
- Watches `/var/log/audit/audit.log`
- Extracts and displays executed command arguments
- Useful for tracking system command execution in real time

### conndump

A network connection dumper that:
- Retrieves TCP connection information
- Outputs connection details in JSON format
- Uses `/proc/net/tcp` for information gathering

### shotgun

Kills any processes that starts listening on the given TCP port

**Usage**:
```
shotgun <port1> [port2] ... <portN>
```

### netwatcher

A process network connection tracker that:
- Monitors network connections for a specific process
- Tracks new connections in real-time
- Handles process restarts automatically by searching for the process name

