package storage_pools

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mahcks/serra/pkg/structures"
)

type StoragePoolService struct{}

func NewStoragePoolService() *StoragePoolService {
	return &StoragePoolService{}
}

// DetectAllPools detects all storage pools on the system
func (s *StoragePoolService) DetectAllPools() ([]structures.StoragePool, error) {
	var allPools []structures.StoragePool

	// Detect ZFS pools
	if zfsPools, err := s.DetectZFSPools(); err == nil {
		for _, pool := range zfsPools {
			allPools = append(allPools, pool.StoragePool)
		}
	}

	// Detect UnRAID arrays
	if unraidArrays, err := s.DetectUnRAIDArrays(); err == nil {
		for _, array := range unraidArrays {
			allPools = append(allPools, array.StoragePool)
		}
	}

	return allPools, nil
}

// DetectZFSPools detects ZFS pools on the system
func (s *StoragePoolService) DetectZFSPools() ([]structures.ZFSPool, error) {
	var pools []structures.ZFSPool

	// Check if zpool command exists
	if !s.commandExists("zpool") {
		return pools, fmt.Errorf("zpool command not found")
	}

	// Get basic pool information
	cmd := exec.Command("zpool", "list", "-H", "-o", "name,size,alloc,free,health")
	output, err := cmd.Output()
	if err != nil {
		return pools, fmt.Errorf("failed to list ZFS pools: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		poolName := fields[0]
		totalSize := s.parseSize(fields[1])
		usedSize := s.parseSize(fields[2])
		availableSize := s.parseSize(fields[3])
		health := fields[4]

		var usagePercentage float64
		if totalSize > 0 {
			usagePercentage = float64(usedSize) / float64(totalSize) * 100
		}

		pool := structures.ZFSPool{
			StoragePool: structures.StoragePool{
				Name:            poolName,
				Type:            "zfs",
				Health:          health,
				Status:          s.getZFSPoolStatus(poolName),
				TotalSize:       totalSize,
				UsedSize:        usedSize,
				AvailableSize:   availableSize,
				UsagePercentage: usagePercentage,
				Redundancy:      s.getZFSPoolRedundancy(poolName),
				Devices:         s.getZFSPoolDevices(poolName),
				LastChecked:     time.Now(),
			},
			Compression:     s.getZFSProperty(poolName, "compression"),
			Deduplication:   s.getZFSProperty(poolName, "dedup"),
			ScrubStatus:     s.getZFSScrubStatus(poolName),
			FragmentationPct: s.getZFSFragmentation(poolName),
		}

		pools = append(pools, pool)
	}

	return pools, nil
}

// DetectUnRAIDArrays detects UnRAID arrays
func (s *StoragePoolService) DetectUnRAIDArrays() ([]structures.UnRAIDArray, error) {
	var arrays []structures.UnRAIDArray

	// Check if we're on UnRAID by looking for specific files
	if !s.fileExists("/proc/mdcmd") && !s.fileExists("/var/local/emhttp/array.ini") {
		return arrays, fmt.Errorf("UnRAID not detected")
	}

	// Parse /proc/mdstat for array information
	mdstat, err := s.parseMdstat()
	if err != nil {
		return arrays, err
	}

	for arrayName, arrayInfo := range mdstat {
		if !strings.HasPrefix(arrayName, "md") {
			continue
		}

		// Get UnRAID-specific array information
		array := structures.UnRAIDArray{
			StoragePool: structures.StoragePool{
				Name:            arrayName,
				Type:            "unraid",
				Health:          arrayInfo.Health,
				Status:          arrayInfo.Status,
				TotalSize:       arrayInfo.TotalSize,
				UsedSize:        arrayInfo.UsedSize,
				AvailableSize:   arrayInfo.AvailableSize,
				UsagePercentage: arrayInfo.UsagePercentage,
				Redundancy:      fmt.Sprintf("Parity: %d", arrayInfo.ParityCount),
				Devices:         arrayInfo.Devices,
				LastChecked:     time.Now(),
			},
			ParityDevices: arrayInfo.ParityDevices,
			DataDevices:   arrayInfo.DataDevices,
			SyncStatus:    arrayInfo.SyncStatus,
			SyncProgress:  arrayInfo.SyncProgress,
		}

		arrays = append(arrays, array)
	}

	return arrays, nil
}

// Helper functions

func (s *StoragePoolService) commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (s *StoragePoolService) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *StoragePoolService) parseSize(sizeStr string) int64 {
	// Remove any whitespace
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "-" || sizeStr == "" {
		return 0
	}

	// Handle ZFS size formats (e.g., "1.23T", "456G", "789M")
	multiplier := int64(1)
	unit := sizeStr[len(sizeStr)-1:]
	
	switch strings.ToUpper(unit) {
	case "T":
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	case "G":
		multiplier = 1024 * 1024 * 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	case "M":
		multiplier = 1024 * 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	case "K":
		multiplier = 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	}

	if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
		return int64(size * float64(multiplier))
	}

	return 0
}

func (s *StoragePoolService) getZFSPoolStatus(poolName string) string {
	cmd := exec.Command("zpool", "status", poolName)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "state:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return "unknown"
}

func (s *StoragePoolService) getZFSPoolRedundancy(poolName string) string {
	cmd := exec.Command("zpool", "status", poolName)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse the pool configuration to determine redundancy
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "mirror") {
			return "mirror"
		} else if strings.Contains(trimmed, "raidz3") {
			return "raidz3"
		} else if strings.Contains(trimmed, "raidz2") {
			return "raidz2"
		} else if strings.Contains(trimmed, "raidz1") || strings.Contains(trimmed, "raidz") {
			return "raidz1"
		}
	}

	return "stripe"
}

func (s *StoragePoolService) getZFSPoolDevices(poolName string) []structures.StoragePoolDevice {
	var devices []structures.StoragePoolDevice

	cmd := exec.Command("zpool", "status", poolName)
	output, err := cmd.Output()
	if err != nil {
		return devices
	}

	lines := strings.Split(string(output), "\n")
	inConfig := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.Contains(line, "config:") {
			inConfig = true
			continue
		}
		
		if inConfig && (strings.Contains(line, "errors:") || strings.Contains(line, "spares:")) {
			break
		}
		
		if inConfig && strings.HasPrefix(line, "\t") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 5 && !strings.Contains(fields[0], poolName) {
				device := structures.StoragePoolDevice{
					Name:   fields[0],
					Path:   fields[0],
					Status: fields[1],
					Health: fields[1],
				}
				
				// Parse error counts if available
				if len(fields) >= 5 {
					if readErr, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
						device.ReadErrors = readErr
					}
					if writeErr, err := strconv.ParseInt(fields[3], 10, 64); err == nil {
						device.WriteErrors = writeErr
					}
					if chkErr, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
						device.ChecksumErrors = chkErr
					}
				}
				
				devices = append(devices, device)
			}
		}
	}

	return devices
}

func (s *StoragePoolService) getZFSProperty(poolName, property string) string {
	cmd := exec.Command("zfs", "get", "-H", "-o", "value", property, poolName)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func (s *StoragePoolService) getZFSScrubStatus(poolName string) string {
	cmd := exec.Command("zpool", "status", poolName)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "scrub") || strings.Contains(line, "resilver") {
			return strings.TrimSpace(line)
		}
	}

	return "none in progress"
}

func (s *StoragePoolService) getZFSFragmentation(poolName string) float64 {
	cmd := exec.Command("zpool", "list", "-H", "-o", "frag", poolName)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	fragStr := strings.TrimSpace(string(output))
	fragStr = strings.TrimSuffix(fragStr, "%")
	
	if frag, err := strconv.ParseFloat(fragStr, 64); err == nil {
		return frag
	}

	return 0
}

// UnRAID-specific types for parsing
type UnRAIDArrayInfo struct {
	Health          string
	Status          string
	TotalSize       int64
	UsedSize        int64
	AvailableSize   int64
	UsagePercentage float64
	ParityCount     int
	Devices         []structures.StoragePoolDevice
	ParityDevices   []structures.StoragePoolDevice
	DataDevices     []structures.StoragePoolDevice
	SyncStatus      string
	SyncProgress    float64
}

func (s *StoragePoolService) parseMdstat() (map[string]UnRAIDArrayInfo, error) {
	arrays := make(map[string]UnRAIDArrayInfo)

	file, err := os.Open("/proc/mdstat")
	if err != nil {
		return arrays, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentArray string
	var arrayInfo UnRAIDArrayInfo

	for scanner.Scan() {
		line := scanner.Text()
		
		// Look for array lines (e.g., "md1 : active raid5 sdb1[0] sdc1[1] sdd1[2]")
		if strings.HasPrefix(line, "md") && strings.Contains(line, " : ") {
			// Save previous array if exists
			if currentArray != "" {
				arrays[currentArray] = arrayInfo
			}
			
			// Parse new array
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				currentArray = parts[0]
				arrayInfo = UnRAIDArrayInfo{
					Status: parts[2], // active/inactive
					Health: "healthy", // Default, will be updated if issues found
				}
				
				// Parse device list
				for i := 4; i < len(parts); i++ {
					deviceInfo := parts[i]
					deviceName := strings.Split(deviceInfo, "[")[0]
					
					device := structures.StoragePoolDevice{
						Name:   deviceName,
						Path:   "/dev/" + deviceName,
						Status: "active",
						Health: "healthy",
					}
					
					arrayInfo.Devices = append(arrayInfo.Devices, device)
				}
			}
		} else if currentArray != "" && strings.Contains(line, "blocks") {
			// Parse size information from lines like "1234567 blocks super 1.2 level 5, 512k chunk, algorithm 2 [3/3] [UUU]"
			parts := strings.Fields(line)
			if len(parts) > 0 {
				if blocks, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					arrayInfo.TotalSize = blocks * 1024 // Convert from blocks to bytes
				}
			}
		}
	}

	// Save the last array
	if currentArray != "" {
		arrays[currentArray] = arrayInfo
	}

	return arrays, nil
}