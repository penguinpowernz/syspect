package main

// AI SLOP

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type FileInfo struct {
	Path        string
	Exists      bool
	MD5Sum      string
	SHA256Sum   string
	Package     string
	PackageInfo string
	MD5Valid    bool
	Processes   []ProcessInfo
	Connections []ConnectionInfo
}

type ProcessInfo struct {
	PID        int32
	Name       string
	Command    string
	User       string
	Executable string
}

type ConnectionInfo struct {
	PID      int32
	Protocol string
	Local    string
	Remote   string
	Status   string
}

type PathSearchResult struct {
	Path   string
	MD5Sum string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file_path_or_filename>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s /usr/bin/telegraf    # Full analysis of specific file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s telegraf             # Search for 'telegraf' in PATH\n", os.Args[0])
		os.Exit(1)
	}

	input := os.Args[1]

	// Check if input is just a filename (no path separators)
	if !strings.Contains(input, "/") && !strings.Contains(input, "\\") {
		// Search in PATH
		searchInPath(input)
		return
	}

	// Get absolute path for full analysis
	absPath, err := filepath.Abs(input)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		absPath = input
	}

	info := analyzeFile(absPath)
	displayResults(info)
}

func searchInPath(filename string) {
	fmt.Printf("🔍 Searching for '%s' in PATH...\n\n", filename)

	// Get PATH environment variable
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		fmt.Printf("❌ PATH environment variable is empty\n")
		return
	}

	// Split PATH by the appropriate separator
	pathSeparator := ":"
	if os.PathSeparator == '\\' { // Windows
		pathSeparator = ";"
	}

	pathDirs := strings.Split(pathEnv, pathSeparator)
	var results []PathSearchResult

	fmt.Printf("📂 Searching in %d directories...\n\n", len(pathDirs))

	for _, dir := range pathDirs {
		if dir == "" {
			continue
		}

		// Clean the directory path
		dir = strings.TrimSpace(dir)
		fullPath := filepath.Join(dir, filename)

		// Check if file exists and is executable
		if info, err := os.Stat(fullPath); err == nil {
			// Check if it's a regular file
			if info.Mode().IsRegular() {
				// Calculate MD5
				md5sum := calculateMD5Only(fullPath)

				result := PathSearchResult{
					Path:   fullPath,
					MD5Sum: md5sum,
				}
				results = append(results, result)

				fmt.Printf("✅ Found: %s\n", fullPath)
				fmt.Printf("   MD5: %s\n", md5sum)
				fmt.Printf("   Size: %d bytes\n", info.Size())
				fmt.Printf("   Mode: %s\n", info.Mode())
				fmt.Printf("   Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))

				// Check if it's executable
				if info.Mode()&0111 != 0 {
					fmt.Printf("   ✓ Executable\n")
				} else {
					fmt.Printf("   ⚠ Not executable\n")
				}

				// Try to get file type
				if fileType := getFileType(fullPath); fileType != "" {
					fmt.Printf("   Type: %s\n", fileType)
				}

				fmt.Println()
			}
		}
	}

	// Summary
	fmt.Printf("📊 Search Summary:\n")
	if len(results) == 0 {
		fmt.Printf("❌ No files named '%s' found in PATH\n", filename)
	} else {
		fmt.Printf("✅ Found %d instance(s) of '%s':\n", len(results), filename)
		for i, result := range results {
			fmt.Printf("%d. %s (MD5: %s)\n", i+1, result.Path, result.MD5Sum)
		}

		// Check for duplicates (same MD5)
		fmt.Printf("\n🔍 Duplicate Analysis:\n")
		md5Map := make(map[string][]string)
		for _, result := range results {
			md5Map[result.MD5Sum] = append(md5Map[result.MD5Sum], result.Path)
		}

		hasDuplicates := false
		for md5sum, paths := range md5Map {
			if len(paths) > 1 {
				hasDuplicates = true
				fmt.Printf("🔄 Identical files (MD5: %s):\n", md5sum)
				for _, path := range paths {
					fmt.Printf("   - %s\n", path)
				}
				fmt.Println()
			}
		}

		if !hasDuplicates {
			fmt.Printf("✅ All found files are unique (different MD5 hashes)\n")
		}

		// Offer to analyze a specific file
		fmt.Printf("\n💡 To perform full analysis on any of these files, run:\n")
		for _, result := range results {
			fmt.Printf("   %s %s\n", os.Args[0], result.Path)
		}
	}
}

func calculateMD5Only(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return "Error calculating MD5"
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "Error calculating MD5"
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getFileType(filePath string) string {
	output, err := runCommand("file", "-b", filePath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

func analyzeFile(filePath string) FileInfo {
	info := FileInfo{Path: filePath}

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		info.Exists = false
		return info
	}
	info.Exists = true

	fmt.Printf("🔍 Analyzing file: %s\n\n", filePath)

	// Calculate hashes
	info.MD5Sum, info.SHA256Sum = calculateHashes(filePath)

	// Run file command
	runFileCommand(filePath)

	// Run stat command
	runStatCommand(filePath)

	// Find package
	info.Package = findPackage(filePath)

	// Get package info if package found
	if info.Package != "" {
		info.PackageInfo = getPackageInfo(info.Package)
		info.MD5Valid = checkMD5Sum(filePath, info.Package, info.MD5Sum)
	}

	// Find processes using /proc filesystem
	info.Processes = findProcessesViaProcFS(filePath)

	// Find network connections for those processes
	info.Connections = findConnectionsForProcesses(info.Processes)

	return info
}

func calculateHashes(filePath string) (string, string) {
	file, err := os.Open(filePath)
	if err != nil {
		return "Error calculating hashes", "Error calculating hashes"
	}
	defer file.Close()

	md5Hash := md5.New()
	sha256Hash := sha256.New()

	// Use MultiWriter to calculate both hashes in one pass
	multiWriter := io.MultiWriter(md5Hash, sha256Hash)

	if _, err := io.Copy(multiWriter, file); err != nil {
		return "Error calculating hashes", "Error calculating hashes"
	}

	return fmt.Sprintf("%x", md5Hash.Sum(nil)), fmt.Sprintf("%x", sha256Hash.Sum(nil))
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func runFileCommand(filePath string) {
	fmt.Printf("📄 File command output:\n")
	output, err := runCommand("file", filePath)
	if err != nil {
		fmt.Printf("Error running file command: %v\n", err)
	} else {
		fmt.Printf("%s\n", strings.TrimSpace(output))
	}
	fmt.Println()
}

func runStatCommand(filePath string) {
	fmt.Printf("📊 Stat command output:\n")
	output, err := runCommand("stat", filePath)
	if err != nil {
		fmt.Printf("Error running stat command: %v\n", err)
	} else {
		fmt.Printf("%s\n", strings.TrimSpace(output))
	}
	fmt.Println()
}

func findPackage(filePath string) string {
	fmt.Printf("📦 Package information:\n")
	output, err := runCommand("dpkg-query", "-S", filePath)
	if err != nil {
		fmt.Printf("File not found in any package or error occurred: %v\n", err)
		fmt.Println()
		return ""
	}

	// Extract package name from output (format: "package: filepath")
	parts := strings.SplitN(strings.TrimSpace(output), ":", 2)
	if len(parts) < 2 {
		fmt.Printf("Unexpected dpkg-query output format\n")
		fmt.Println()
		return ""
	}

	packageName := strings.TrimSpace(parts[0])
	fmt.Printf("Found in package: %s\n", packageName)
	fmt.Println()
	return packageName
}

func getPackageInfo(packageName string) string {
	fmt.Printf("📋 Package policy information:\n")
	output, err := runCommand("apt-cache", "policy", packageName)
	if err != nil {
		fmt.Printf("Error getting package policy: %v\n", err)
		fmt.Println()
		return ""
	}

	fmt.Printf("%s\n", strings.TrimSpace(output))
	fmt.Println()
	return strings.TrimSpace(output)
}

func checkMD5Sum(filePath, packageName, calculatedMD5 string) bool {
	fmt.Printf("🔐 MD5 Sum verification:\n")
	fmt.Printf("Calculated MD5: %s\n", calculatedMD5)

	// Try to find the md5sums file for the package
	md5sumPath := fmt.Sprintf("/var/lib/dpkg/info/%s.md5sums", packageName)

	file, err := os.Open(md5sumPath)
	if err != nil {
		fmt.Printf("Could not open md5sums file: %v\n", err)
		fmt.Println()
		return false
	}
	defer file.Close()

	// Get relative path from root
	relativePath := strings.TrimPrefix(filePath, "/")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			expectedMD5 := parts[0]
			filePath := strings.Join(parts[1:], " ")

			if filePath == relativePath {
				fmt.Printf("Expected MD5:   %s\n", expectedMD5)
				if expectedMD5 == calculatedMD5 {
					fmt.Printf("✅ MD5 verification: MATCH\n")
					fmt.Println()
					return true
				} else {
					fmt.Printf("❌ MD5 verification: MISMATCH\n")
					fmt.Println()
					return false
				}
			}
		}
	}

	fmt.Printf("File not found in package md5sums\n")
	fmt.Println()
	return false
}

func findProcessesViaProcFS(targetFile string) []ProcessInfo {
	fmt.Printf("🔄 Running processes (scanning /proc filesystem):\n")

	var processes []ProcessInfo

	// Get target file's absolute path and resolve any symlinks
	targetAbs, err := filepath.Abs(targetFile)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		fmt.Println()
		return processes
	}

	targetResolved, err := filepath.EvalSymlinks(targetAbs)
	if err != nil {
		// If we can't resolve symlinks, use the absolute path
		targetResolved = targetAbs
	}

	// Get all processes using gopsutil
	allProcs, err := process.Processes()
	if err != nil {
		fmt.Printf("Error getting processes: %v\n", err)
		fmt.Println()
		return processes
	}

	for _, proc := range allProcs {
		// Get executable path
		exe, err := proc.Exe()
		if err != nil {
			// Skip if we can't get executable info (likely permission issue)
			continue
		}

		// Resolve symlinks in the process executable path
		exeResolved, err := filepath.EvalSymlinks(exe)
		if err != nil {
			exeResolved = exe
		}

		// Check if this process is running our target file
		if exeResolved == targetResolved || exe == targetAbs {
			name, _ := proc.Name()
			cmdline, _ := proc.Cmdline()
			username, _ := proc.Username()

			procInfo := ProcessInfo{
				PID:        proc.Pid,
				Name:       name,
				Command:    cmdline,
				User:       username,
				Executable: exe,
			}
			processes = append(processes, procInfo)

			fmt.Printf("PID: %d, User: %s, Name: %s\n", procInfo.PID, procInfo.User, procInfo.Name)
			fmt.Printf("  Executable: %s\n", procInfo.Executable)
			if procInfo.Command != "" {
				fmt.Printf("  Command: %s\n", procInfo.Command)
			}
			fmt.Println()
		}
	}

	if len(processes) == 0 {
		fmt.Printf("No running processes found\n")
		fmt.Println()
	}

	return processes
}

func findConnectionsForProcesses(processes []ProcessInfo) []ConnectionInfo {
	fmt.Printf("🌐 Network connections:\n")

	var connections []ConnectionInfo

	if len(processes) == 0 {
		fmt.Printf("No processes to check for network connections\n")
		fmt.Println()
		return connections
	}

	// Create a map of PIDs we're interested in
	pidMap := make(map[int32]bool)
	for _, proc := range processes {
		pidMap[proc.PID] = true
	}

	// Get all network connections
	allConnections, err := net.Connections("all")
	if err != nil {
		fmt.Printf("Error getting network connections: %v\n", err)
		fmt.Println()
		return connections
	}

	// Filter connections by our PIDs
	foundConnections := make(map[string]ConnectionInfo) // Use map to avoid duplicates

	for _, conn := range allConnections {
		if pidMap[conn.Pid] {
			local := fmt.Sprintf("%s:%d", conn.Laddr.IP, conn.Laddr.Port)
			remote := fmt.Sprintf("%s:%d", conn.Raddr.IP, conn.Raddr.Port)

			// Handle listening sockets (no remote address)
			if conn.Raddr.IP == "" || conn.Raddr.Port == 0 {
				remote = "*:*"
			}

			connInfo := ConnectionInfo{
				PID:      conn.Pid,
				Protocol: fmt.Sprintf("%s%d", conn.Type, conn.Family),
				Local:    local,
				Remote:   remote,
				Status:   conn.Status,
			}

			// Create a unique key for deduplication
			key := fmt.Sprintf("%d-%s-%s-%s-%s", conn.Pid, connInfo.Protocol, local, remote, conn.Status)
			foundConnections[key] = connInfo
		}
	}

	// Convert map back to slice
	for _, conn := range foundConnections {
		connections = append(connections, conn)
	}

	if len(connections) == 0 {
		fmt.Printf("No network connections found for the processes\n")
	} else {
		// Group by PID for better display
		pidGroups := make(map[int32][]ConnectionInfo)
		for _, conn := range connections {
			pidGroups[conn.PID] = append(pidGroups[conn.PID], conn)
		}

		for pid, conns := range pidGroups {
			fmt.Printf("PID %d:\n", pid)
			for _, conn := range conns {
				fmt.Printf("  %s %s -> %s (%s)\n",
					conn.Protocol, conn.Local, conn.Remote, conn.Status)
			}
			fmt.Println()
		}
	}

	return connections
}

func displayResults(info FileInfo) {
	if !info.Exists {
		fmt.Printf("❌ File does not exist: %s\n", info.Path)
		return
	}

	fmt.Printf("📝 Summary:\n")
	fmt.Printf("File: %s\n", info.Path)
	fmt.Printf("MD5: %s\n", info.MD5Sum)
	fmt.Printf("SHA256: %s\n", info.SHA256Sum)

	if info.Package != "" {
		fmt.Printf("Package: %s\n", info.Package)
		if info.MD5Valid {
			fmt.Printf("MD5 Integrity: ✅ Valid\n")
		} else {
			fmt.Printf("MD5 Integrity: ❌ Invalid or Unknown\n")
		}
	} else {
		fmt.Printf("Package: Not found in package manager\n")
	}

	fmt.Printf("Running Processes: %d\n", len(info.Processes))
	fmt.Printf("Network Connections: %d\n", len(info.Connections))

	// VirusTotal link with SHA256
	fmt.Printf("\n🦠 VirusTotal Analysis:\n")
	fmt.Printf("https://www.virustotal.com/gui/file/%s\n", info.SHA256Sum)

	fmt.Println()
}
