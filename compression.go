package polytopiamapmodel

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"

	lz4 "github.com/pierrec/lz4/v4"
)

func DecompressFile(inputFilename string) {
	decompressedContents := GetDecompressedContents(inputFilename)
	decompressedLength := len(decompressedContents)

	decompressedFilename := inputFilename + ".decomp"
	if err := os.WriteFile(decompressedFilename, decompressedContents[:decompressedLength], 0666); err != nil {
		log.Fatal("Error writing decompressed contents", err)
	}
}

func BuildReaderForDecompressedFile(inputFilename string) (*bytes.Reader, int) {
	decompressedContents := GetDecompressedContents(inputFilename)
	decompressedLength := len(decompressedContents)
	return bytes.NewReader(decompressedContents), decompressedLength
}

func GetDecompressedContents(inputFilename string) []byte {
	inputFile, err := os.Open(inputFilename)
	defer inputFile.Close()
	if err != nil {
		log.Fatal("Failed to load state file: ", err)
		return nil
	}

	inputBuffer := new(bytes.Buffer)
	inputBuffer.ReadFrom(inputFile)

	inputBytes := inputBuffer.Bytes()
	firstByte := inputBytes[0]
	sizeOfDiff := ((firstByte >> 6) & 3)
	if sizeOfDiff == 3 {
		sizeOfDiff = 4
	}
	dataOffset := 1 + int(sizeOfDiff)
	var resultDiff int
	if sizeOfDiff == 4 {
		resultDiff = int(binary.LittleEndian.Uint32(inputBytes[1 : 1+sizeOfDiff]))
	} else if sizeOfDiff == 2 {
		resultDiff = int(binary.LittleEndian.Uint16(inputBytes[1 : 1+sizeOfDiff]))
	} else {
		log.Fatal("Header sizeOfDiff is unrecognized value: ", sizeOfDiff)
	}
	dataLength := len(inputBytes) - dataOffset
	resultLength := dataLength + resultDiff

	// decompress
	decompressedContents := make([]byte, resultLength)
	decompressedLength, err := lz4.UncompressBlock(inputBytes[dataOffset:], decompressedContents)
	if err != nil {
		panic(err)
	}

	return decompressedContents[:decompressedLength]
}

func CompressFile(inputFilename string, outputFilename string) {
	inputFile, err := os.Open(inputFilename)
	defer inputFile.Close()
	if err != nil {
		log.Fatal("Failed to load state file: ", err)
		return
	}

	inputBuffer := new(bytes.Buffer)
	inputBuffer.ReadFrom(inputFile)
	inputBytes := inputBuffer.Bytes()

	decompressedLength := len(inputBytes)
	compressedContents := make([]byte, decompressedLength)
	compressedLength, err := lz4.CompressBlock(inputBytes, compressedContents, []int{0})
	if err != nil {
		panic(err)
	}

	newCompressedContents := []byte{}
	if decompressedLength >= 65536 {
		byteArrDecompressedSize := make([]byte, 4)
		binary.LittleEndian.PutUint32(byteArrDecompressedSize, uint32(decompressedLength-compressedLength))

		newCompressedContents = append([]byte{0xC0}, byteArrDecompressedSize...)
		newCompressedContents = append(newCompressedContents, compressedContents[:compressedLength]...)
	} else {
		byteArrDecompressedSize := make([]byte, 2)
		binary.LittleEndian.PutUint16(byteArrDecompressedSize, uint16(decompressedLength-compressedLength))

		newCompressedContents = append([]byte{0x80}, byteArrDecompressedSize...)
		newCompressedContents = append(newCompressedContents, compressedContents[:compressedLength]...)
	}

	if err := os.WriteFile(outputFilename, newCompressedContents, 0666); err != nil {
		log.Fatal("Error writing compressed contents", err)
	}
}
