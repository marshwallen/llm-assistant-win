package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"winds-assistant/utils"
	"log"
)

// 获取所有Windows盘符
func GetDrives() ([]string, error) {
	out, _, err := utils.RunCommand("wmic", "logicaldisk", "get", "deviceid")
	if err != nil {
		return nil, err
	}
	
	drives := strings.Split(out, "\n")
	var result []string
	for _, drive := range drives {
		drive = strings.TrimSpace(drive)
		if drive != "" && drive != "DeviceID" {
			result = append(result, drive+"/")
		}
	}
	return result, nil
}

// 递归遍历目录并打印结构/大小
// 因小文件过多，可选是否打印非文件夹文件
func traverseDir(path string, depth int, maxPrintDepth int, maxSearchDepth int, contain_files bool) (ftree string, size int64, err error) {
	if depth > maxSearchDepth {
		return
	}

	// 获取当前路径下的所有文件和目录
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	sftree := ""
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			// 递归子目录
			_s, _size, _ := traverseDir(fullPath, depth+1, maxPrintDepth, maxSearchDepth, contain_files)
			size += _size
			sftree += _s
		} else {
			// 处理非目录文件
			info, _ := entry.Info()
			size += info.Size()
			if contain_files {
				sftree += fmt.Sprintf(
					"%*s %s %.2f MB\n", 
					(depth+1)*2, "-", 
					entry.Name(), 
					float64(info.Size())/(1024.00*1024.00),
				)
			}
		}
	}

	if depth > maxPrintDepth {
		return
	}

	_l := strings.Split(path,"\\")
	_path := _l[len(_l)-1] + "/"
	ftree = fmt.Sprintf(
		"%*s %s %.2f MB\n", 
		depth*2, "-", 
		_path, 
		float64(size)/(1024.0*1024.0),
	)

	ftree += sftree
	return
}

// 获取指定深度的文件树结构。
// 参数:
//   diskname: 要获取文件树的磁盘名称。
//   maxPrintDepth: 打印文件树的最大深度。
//   maxSearchDepth: 搜索文件的最大深度。
//   contain_files: 是否打印非文件夹文件信息。
func GetFileTree(diskname string, maxPrintDepth int, maxSearchDepth int, contain_files bool) (ftree string, err error) {
	// 获取所有盘符
	drives, err := GetDrives()
	if err != nil {
		log.Fatalf("Error getting drives: %v", err)
		return 
	}

	for _, drive := range drives {
		if diskname != "" && diskname != drive {
			continue
		}
		ftree += fmt.Sprintf("%s\n", drive)
		s, _, _ := traverseDir(drive, 0, maxPrintDepth, maxSearchDepth, contain_files)
		ftree += s
	}
	return
}

// 函数用于保存指定磁盘名称的文件树结构到文件中。
// 参数:
//   diskname: 要获取文件树的磁盘名称。
//   maxPrintDepth: 打印文件树的最大深度。
//   maxSearchDepth: 搜索文件的最大深度。
//   contain_files: 是否打印非文件夹文件信息。
// 此函数会调用 GetFileTree 获取文件树数据，并将其保存到 "data/filetree_disk_<diskname>.txt" 文件中
func SaveFileTree(diskname string, maxPrintDepth int, maxSearchDepth int, contain_files bool) {
	data, _ := GetFileTree(diskname, maxPrintDepth, maxSearchDepth, contain_files)

	// 初始化数据存放路径
	if err := utils.EnsureDir("data/"); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

    err := os.WriteFile(fmt.Sprintf("data/filetree_disk_%s.txt", strings.Split(diskname, ":")[0]), []byte(data), 0644)
    if err != nil {
        log.Fatalf("%v", err)
    }
	fmt.Println("Filetree saved.")
}