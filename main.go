package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/goburrow/serial"
	ams "github.com/karl-gustav/ams-han"
)

var (
	address  string
	baudrate int
	databits int
	stopbits int
	parity   string
	verbose  bool
)

func main() {
	flag.StringVar(&address, "a", "/dev/ttyUSB2", "address")
	flag.IntVar(&baudrate, "b", 2400, "baud rate")
	flag.IntVar(&databits, "d", 8, "data bits")
	flag.IntVar(&stopbits, "s", 1, "stop bits")
	flag.StringVar(&parity, "p", "E", "parity (N/E/O)")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.Parse()

	if verbose == true {
		os.Setenv("DEBUG", "DEBUG")
	}
	serialPort := getSerialPort(address, baudrate, databits, stopbits, parity)
	byteStream := createByteStream(serialPort)

	next := ams.ByteReader(byteStream)
	for {
		bytePackage, err := next()
		if err != nil {
			fmt.Println("[ERROR]", err)
			if err == ams.CHANNEL_IS_CLOSED_ERROR {
				return
			}
			continue
		}

		if verbose {
			fmt.Printf("\nBuffer(%d):\n%s\n", len(bytePackage), byteArrayToHexStringArray(bytePackage))
		}

		message, err := ams.BytesParser(bytePackage)
		if err != nil {
			fmt.Println("[ERROR]", err)
			continue
		}
		jsonString, _ := json.Marshal(message)
		fmt.Printf("%s\n", jsonString)
	}
}

func getSerialPort(Address string, BaudRate int, DataBits int, StopBits int, Parity string) (port serial.Port) {
	config := serial.Config{
		Address:  address,
		BaudRate: baudrate,
		DataBits: databits,
		StopBits: stopbits,
		Parity:   parity,
		Timeout:  60 * time.Second,
	}
	if verbose {
		log.Printf("connecting %+v\n", config)
	}
	port, err := serial.Open(&config)
	if err != nil {
		log.Fatal(err)
	}
	if verbose {
		log.Println("connected")
	}
	return
}

func createByteStream(port serial.Port) chan byte {
	serialChannel := make(chan byte)

	go func() {
		var buf [8]byte
		for {
			n, err := port.Read(buf[:])
			if err == io.EOF {
				log.Fatalln("Reached end of stream")
				break
			} else if err != nil {
				log.Println("[ERROR]:", err)
				break
			}
			for i := 0; i < n; i++ {
				serialChannel <- buf[i]
			}
		}

		err := port.Close()
		log.Println("Closed connection!")
		close(serialChannel)
		if err != nil {
			log.Println(err)
		}
	}()
	return serialChannel
}

func byteArrayToHexStringArray(bytes []byte) string {
	var bytesAsString []string
	for _, b := range bytes {
		bytesAsString = append(bytesAsString, fmt.Sprintf("0x%02x", b))
	}
	return fmt.Sprintf("[%s]", strings.Join(bytesAsString, ", "))
}
