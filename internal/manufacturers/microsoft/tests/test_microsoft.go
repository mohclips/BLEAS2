package main

// go run internal/manufacturers/microsoft/tests/test_microsoft.go

import (
	"fmt"

	microsoft "github.com/mohclips/BLEAS2/internal/manufacturers/microsoft"
)

var testData = [][]byte{
	//grep "Manufacturer: Microsoft" scan6.log | cut -b128- | sort | uniq
	{6, 0, 1, 9, 32, 2, 100, 249, 61, 127, 54, 229, 151, 142, 235, 115, 40, 106, 108, 208, 176, 132, 121, 119, 51, 228, 127, 15, 12},
	{6, 0, 1, 9, 32, 2, 109, 117, 23, 172, 181, 200, 183, 22, 7, 230, 58, 127, 90, 197, 25, 34, 246, 177, 114, 84, 191, 35, 241},
	{6, 0, 1, 9, 32, 2, 119, 157, 103, 179, 22, 243, 183, 252, 98, 240, 45, 60, 121, 150, 31, 38, 56, 167, 119, 157, 200, 60, 116},
	{6, 0, 1, 9, 32, 2, 153, 196, 24, 20, 167, 77, 2, 66, 88, 213, 80, 235, 199, 183, 115, 212, 115, 148, 164, 119, 235, 143, 9},
	{6, 0, 1, 9, 32, 2, 189, 225, 200, 70, 128, 50, 172, 192, 45, 225, 241, 219, 244, 24, 142, 195, 245, 100, 173, 24, 115, 171, 63},
	{6, 0, 1, 9, 32, 2, 192, 42, 218, 13, 217, 139, 167, 1, 82, 119, 179, 174, 201, 174, 235, 58, 112, 111, 217, 49, 250, 119, 224},
	{6, 0, 1, 9, 32, 2, 27, 0, 48, 153, 31, 185, 154, 70, 216, 81, 19, 90, 46, 45, 59, 60, 82, 104, 36, 76, 0, 104, 62},
	{6, 0, 1, 9, 32, 2, 63, 57, 247, 1, 173, 31, 188, 174, 40, 223, 135, 24, 250, 214, 250, 93, 177, 96, 124, 159, 194, 174, 221},
}

func main() {
	for i := 0; i < len(testData); i++ {
		var d []byte = testData[i]
		fmt.Println("test data: ", d)

		pkt := microsoft.ParseMF(testData[i])

		fmt.Printf("Final Microsoft: %s\n", pkt)

	}
}
