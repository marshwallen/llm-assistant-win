package tools

import (
    "fmt"
    "github.com/StackExchange/wmi"
)

type DriverStatus struct {
    DeviceName    string
    Manufacturer  string
    DriverVersion string
    Status        string
    IsSigned      bool
}

// 查询系统中的所有驱动程序信息，包括设备名称、制造商、驱动版本、状态和签名状态
func GetSysDriver() (drivers []DriverStatus, err error) {
    // WMI 查询
    query := "SELECT DeviceName, Manufacturer, DriverVersion, Status, IsSigned FROM Win32_PnPSignedDriver"
    if err = wmi.Query(query, &drivers); err != nil {
        return
    }
	return
}

// 返回系统驱动的字符串表示，包括驱动名称、版本、状态和签名信息
func GetSysDriverStr() (r string) {
    raw, _ := GetSysDriver()
    for _, v := range raw {
        r += fmt.Sprintf("[%s]\n  版本: %s\n  状态: %s\n  签名: %v\n",
            v.DeviceName, v.DriverVersion, v.Status, v.IsSigned)
    }
    return
}