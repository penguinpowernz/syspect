# Syspect: System Inspection and analysis Toolkit

A collection of tools for system analysis.

## Binaries

### fileanal

A comprehensive file analysis tool that:
- has a really awkward name
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

Here is an example checking for telegraf on my system:

```
$ fileanal telegraf
🔍 Searching for 'telegraf' in PATH...

📂 Searching in 7 directories...

✅ Found: /usr/bin/telegraf
   MD5: 00cb4e2fb9ca46f11a2c04408019d2d0
   Size: 177916312 bytes
   Mode: -rwxr-xr-x
   Modified: 2022-03-18 05:28:18
   ✓ Executable
   Type: ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, Go BuildID=cRtX6yK3x34ZJhBoP02E/EPD8YIW_VTWDttbr1OkL/mGkY6uutE90LelWjUU8t/uNA6DmX9ysuyVFSFXnLC, stripped

✅ Found: /bin/telegraf
   MD5: 00cb4e2fb9ca46f11a2c04408019d2d0
   Size: 177916312 bytes
   Mode: -rwxr-xr-x
   Modified: 2022-03-18 05:28:18
   ✓ Executable
   Type: ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, Go BuildID=cRtX6yK3x34ZJhBoP02E/EPD8YIW_VTWDttbr1OkL/mGkY6uutE90LelWjUU8t/uNA6DmX9ysuyVFSFXnLC, stripped

📊 Search Summary:
✅ Found 2 instance(s) of 'telegraf':
1. /usr/bin/telegraf (MD5: 00cb4e2fb9ca46f11a2c04408019d2d0)
2. /bin/telegraf (MD5: 00cb4e2fb9ca46f11a2c04408019d2d0)

🔍 Duplicate Analysis:
🔄 Identical files (MD5: 00cb4e2fb9ca46f11a2c04408019d2d0):
   - /usr/bin/telegraf
   - /bin/telegraf


💡 To perform full analysis on any of these files, run:
   ./fileanal /usr/bin/telegraf
   ./fileanal /bin/telegraf
```

And then the full analysis:

```
$ sudo fileanal /usr/bin/telegraf                                       
🔍 Analyzing file: /usr/bin/telegraf

📄 File command output:
/usr/bin/telegraf: ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, Go BuildID=cRtX6yK3x34ZJhBoP02E/EPD8YIW_VTWDttbr1OkL/mGkY6uutE90LelWjUU8t/uNA6DmX9ysuyVFSFXnLC, stripped

📊 Stat command output:
File: /usr/bin/telegraf
  Size: 177916312       Blocks: 347504     IO Block: 4096   regular file
Device: 10300h/66304d   Inode: 24510966    Links: 1
Access: (0755/-rwxr-xr-x)  Uid: (    0/    root)   Gid: (    0/    root)
Access: 2022-05-12 21:29:39.000000000 +1200
Modify: 2022-03-18 05:28:18.000000000 +1300
Change: 2022-05-12 21:29:40.544928953 +1200
 Birth: 2022-05-12 21:29:39.808912188 +1200

📦 Package information:
Found in package: telegraf

📋 Package policy information:
telegraf:
  Installed: 1.21.4+ds1-0ubuntu2
  Candidate: 1.21.4+ds1-0ubuntu2
  Version table:
 *** 1.21.4+ds1-0ubuntu2 500
        500 http://apt.pop-os.org/ubuntu jammy/universe amd64 Packages
        100 /var/lib/dpkg/status

🔐 MD5 Sum verification:
Calculated MD5: 00cb4e2fb9ca46f11a2c04408019d2d0
Expected MD5:   00cb4e2fb9ca46f11a2c04408019d2d0
✅ MD5 verification: MATCH

🔄 Running processes (scanning /proc filesystem):
PID: 3456911, User: _telegraf, Name: telegraf
  Executable: /usr/bin/telegraf
  Command: /usr/bin/telegraf -config /etc/telegraf/telegraf.conf -config-directory /etc/telegraf/telegraf.d

🌐 Network connections:
PID 3456911:
  %!s(uint32=1)10 :::9273 -> *:* (LISTEN)

📝 Summary:
File: /usr/bin/telegraf
MD5: 00cb4e2fb9ca46f11a2c04408019d2d0
SHA256: fc761d4a65c33625d2cbe8a6293ae8cb1361b4f21c9db2728dd3471969955049
Package: telegraf
MD5 Integrity: ✅ Valid
Running Processes: 1
Network Connections: 1

🦠 VirusTotal Analysis:
https://www.virustotal.com/gui/file/fc761d4a65c33625d2cbe8a6293ae8cb1361b4f21c9db2728dd3471969955049
```

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

Basically an `ss` clone.

A network connection dumper that:
- Retrieves TCP connection information
- Outputs connection details in JSON format
- Uses `/proc/net/tcp` for information gathering

### shotgun

> [!WARNING]
> Untested

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

```
$ netwatcher 1101
🟢 LISTEN :::9273
👀 watching process /usr/bin/telegraf
👍 started BPF trace
👈 tcp4       127.0.0.1:9273 <- 127.0.0.1:53010     
👈 tcp4       127.0.0.1:9273 <- 127.0.0.1:53026     
👈 tcp4       127.0.0.1:9273 <- 127.0.0.1:53034     
👉 tcp4  192.168.1.102:38276 -> 35.223.238.178:443  
👉 tcp4  192.168.1.102:50880 -> 54.192.221.113:80   
👉 tcp4  192.168.1.102:42546 -> 172.105.169.139:443 
☠️ process we were monitoring died
🔍 watching for new '/usr/bin/telegraf' process to start
👀 new process started on 3450624
🟢 LISTEN :::9273
👍 started BPF trace
```
