package main

// go run internal/manufacturers/nhs/tests/test_nhs.go

import (
	"fmt"

	nhs "github.com/mohclips/BLEAS2/internal/manufacturers/nhs"
)

var testData = [][]byte{

	{0x02, 0x01, 0x02, 0x01, 0x56, 0xa9, 0xbe, 0x43, 0x84, 0x49, 0x1c, 0x03, 0x03, 0x6f, 0xfd, 0x17, 0x16, 0x6f, 0xfd, 0xfb, 0x2b, 0xac, 0x2b, 0x83, 0x48, 0x60, 0xc8, 0x88, 0x80, 0x92, 0x34, 0xb5, 0x2f, 0x11, 0x2d, 0x59, 0x93, 0x40, 0x04, 0xa4},
	//{0x02, 0x01, 0x02, 0x01, 0x56, 0xa9, 0xbe, 0x43, 0x84, 0x49, 0x1c, 0x03, 0x03, 0x6f, 0xfd, 0x17, 0x16, 0x6f, 0xfd, 0xfb, 0x2b, 0xac, 0x2b, 0x83, 0x48, 0x60, 0xc8, 0x88, 0x80, 0x92, 0x34, 0xb5, 0x2f, 0x11, 0x2d, 0x59, 0x93, 0x40, 0x04, 0xa5},
	//{0x02, 0x01, 0x02, 0x01, 0x56, 0xa9, 0xbe, 0x43, 0x84, 0x49, 0x1c, 0x03, 0x03, 0x6f, 0xfd, 0x17, 0x16, 0x6f, 0xfd, 0xfb, 0x2b, 0xac, 0x2b, 0x83, 0x48, 0x60, 0xc8, 0x88, 0x80, 0x92, 0x34, 0xb5, 0x2f, 0x11, 0x2d, 0x59, 0x93, 0x40, 0x04, 0xa9},
}

func main() {
	for i := 0; i < len(testData); i++ {
		var d []byte = testData[i]
		fmt.Println("test data: ", d)

		pkt := nhs.Parse(testData[i])

		fmt.Printf("Final NHS: %s\n", pkt)

	}
}