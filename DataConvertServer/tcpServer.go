package main

import (
    "net"
    "time"
    "fmt"
    "strconv"
    "strings"
    "crypto/rand"
    "io"
    "io/ioutil"
    "os"
    "os/exec"
)

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
    uuid := make([]byte, 16)
    n, err := io.ReadFull(rand.Reader, uuid)
    if n != len(uuid) || err != nil {
        return "", err
    }
    // variant bits; see section 4.1.1
    uuid[8] = uuid[8]&^0xc0 | 0x80
    // version 4 (pseudo-random); see section 4.1.3
    uuid[6] = uuid[6]&^0xf0 | 0x40
    return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func checkErr(e error) bool {
    if e != nil {
        fmt.Println(e)
        return true
    }
    return false
}

func HandleServerConnectionRaw(c net.Conn) {
    defer c.Close()

    timeVal := time.Now().Add(2 * time.Minute)
    c.SetDeadline(timeVal)
    c.SetWriteDeadline(timeVal)
    c.SetReadDeadline(timeVal)

    convertTypeBytes := make([]byte, 8)
    readCount, err := c.Read(convertTypeBytes)
    if checkErr(err) {
        return
    }
    if readCount < 8 {
        return
    }

    dataSizeBytes := make([]byte, 8)
    readCount, err = c.Read(dataSizeBytes)
    if checkErr(err) {
        return
    }
    if readCount < 8 {
        return
    }

    convertTypeStr := string(convertTypeBytes)

    dataSizeStr := string(dataSizeBytes)
    dataSizeStr = strings.Replace(dataSizeStr, " ", "", -1)
    dataSize, err := strconv.Atoi(dataSizeStr)
    if checkErr(err) {
        return
    }

    //fmt.Println(convertTypeStr, dataSize)

    switch convertTypeStr {
    case "pngToPvr":
        dataBytes := make([]byte, dataSize)
        totalReadCount := 0
        for totalReadCount < dataSize {
            bytesRef := dataBytes[totalReadCount:]
            fileReadCount, readErr := c.Read(bytesRef)
            if fileReadCount == 0 {
                break
            }
            if readErr != nil {
                break
            }
            totalReadCount += fileReadCount
        }
        if totalReadCount < dataSize {
            return
        }

        uuid, err := newUUID()
        if checkErr(err) {
            return
        }
        // Save file
        filePath := "/tmp/" + uuid + ".png"
        err = ioutil.WriteFile(filePath, dataBytes, 0644)
        if checkErr(err) {
            return
        }

        // Result file path
        resultFile := "/tmp/" + uuid + ".pvr"

        // Defer remove files
        defer os.Remove(filePath)
        defer os.Remove(resultFile)

        // Convert file
        pvrToolPath := "/Applications/Imagination/PowerVR_Graphics/PowerVR_Tools/PVRTexTool/CLI/OSX_x86/PVRTexToolCLI"
        commandText := fmt.Sprintf("%s -f PVRTC2_4 -dither -q pvrtcbest -i %s -o %s", pvrToolPath, filePath, resultFile)
        command := exec.Command("bash", "-c", commandText)
        err = command.Run()
        if checkErr(err) {
            return
        }

        // Send file
        file, err := os.Open(resultFile)
        if checkErr(err) {
            return
        }

        var currentByte int64 = 0
        fileSendBuffer := make([]byte, 1024)
        for {
            fileReadCount, fileErr := file.ReadAt(fileSendBuffer, currentByte)
            if fileReadCount == 0 {
                break
            }

            writtenCount, writeErr := c.Write(fileSendBuffer[:fileReadCount])
            if checkErr(writeErr) {
                return
            }

            currentByte += int64(writtenCount)

            if (fileErr == io.EOF) && (fileReadCount == writtenCount)  {
                break
            }
        }
    }
}

func server() {
    // Прослушивание сервера
    ln, err := net.Listen("tcp", ":10000")
    if err != nil {
        fmt.Println(err)
        return
    }
    for {
        // Принятие соединения
        c, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        // Запуск горутины
        go HandleServerConnectionRaw(c)
    }
}

func main() {
    go server()

    var input string
    fmt.Scanln(&input)
}