package utils

import (
	"bytes"
	"os/exec"
    "golang.org/x/text/encoding/simplifiedchinese"
    "golang.org/x/text/transform"
    "io"
    "strings"
)

// runCommand 执行指定的命令及其参数，并返回命令的输出和可能发生的错误
func RunCommand(cmd string, args ...string) (string, []byte, error) {
    command := exec.Command(cmd, args...)
    outbyte, err := command.CombinedOutput()
    _out, _ := gbkToUTF8(outbyte)
    outstr := string(_out)
    outstr = strings.ReplaceAll(outstr, "  ", "")

    return outstr, outbyte, err
}

// GbkToUtf8 将 GBK 编码的字节切片转换为 UTF-8 编码
func gbkToUTF8(s []byte) ([]byte, error) {
    reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
    d, e := io.ReadAll(reader)
    if e != nil {
        return nil, e
    }
    return d, nil
}