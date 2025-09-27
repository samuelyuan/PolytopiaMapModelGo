package polytopiamapmodel

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

func readVarString(reader *io.SectionReader, varName string) string {
	if DebugMode {
		debugPrint("    Reading %s...\n", varName)
	}

	variableLength := uint8(0)
	if err := binary.Read(reader, binary.LittleEndian, &variableLength); err != nil {
		log.Fatal("Failed to load variable length: ", err)
	}

	stringValue := make([]byte, variableLength)
	if err := binary.Read(reader, binary.LittleEndian, &stringValue); err != nil {
		log.Fatal(fmt.Sprintf("Failed to load string value. Variable length: %v, name: %s. Error:", variableLength, varName), err)
	}

	result := string(stringValue[:])
	if DebugMode {
		debugPrint("    %s: %s\n", varName, result)
	}
	return result
}

func unsafeReadUint32(reader *io.SectionReader) uint32 {
	unsignedIntValue := uint32(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		log.Fatal("Failed to load uint32: ", err)
	}
	return unsignedIntValue
}

func readUint32Safe(reader *io.SectionReader, fieldName string) (uint32, error) {
	if DebugMode {
		debugPrint("    Reading %s...\n", fieldName)
	}
	unsignedIntValue := uint32(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		return 0, fmt.Errorf("failed to read uint32 (%s): %w", fieldName, err)
	}
	if DebugMode {
		debugPrint("    %s: %d\n", fieldName, unsignedIntValue)
	}
	return unsignedIntValue, nil
}

func unsafeReadInt32(reader *io.SectionReader) int32 {
	signedIntValue := int32(0)
	if err := binary.Read(reader, binary.LittleEndian, &signedIntValue); err != nil {
		log.Fatal("Failed to load int32: ", err)
	}
	return signedIntValue
}

func readInt32Safe(reader *io.SectionReader, fieldName string) (int32, error) {
	if DebugMode {
		debugPrint("    Reading %s...\n", fieldName)
	}
	signedIntValue := int32(0)
	if err := binary.Read(reader, binary.LittleEndian, &signedIntValue); err != nil {
		return 0, fmt.Errorf("failed to read int32 (%s): %w", fieldName, err)
	}
	if DebugMode {
		debugPrint("    %s: %d\n", fieldName, signedIntValue)
	}
	return signedIntValue, nil
}

func unsafeReadUint16(reader *io.SectionReader) uint16 {
	unsignedIntValue := uint16(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		log.Fatal("Failed to load uint16: ", err)
	}
	return unsignedIntValue
}

func readUint16Safe(reader *io.SectionReader, fieldName string) (uint16, error) {
	if DebugMode {
		debugPrint("    Reading %s...\n", fieldName)
	}
	unsignedIntValue := uint16(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		return 0, fmt.Errorf("failed to read uint16 (%s): %w", fieldName, err)
	}
	if DebugMode {
		debugPrint("    %s: %d\n", fieldName, unsignedIntValue)
	}
	return unsignedIntValue, nil
}

func unsafeReadInt16(reader *io.SectionReader) int16 {
	signedIntValue := int16(0)
	if err := binary.Read(reader, binary.LittleEndian, &signedIntValue); err != nil {
		log.Fatal("Failed to load int16: ", err)
	}
	return signedIntValue
}

func readInt16Safe(reader *io.SectionReader, fieldName string) (int16, error) {
	if DebugMode {
		debugPrint("    Reading %s...\n", fieldName)
	}
	signedIntValue := int16(0)
	if err := binary.Read(reader, binary.LittleEndian, &signedIntValue); err != nil {
		return 0, fmt.Errorf("failed to read int16 (%s): %w", fieldName, err)
	}
	if DebugMode {
		debugPrint("    %s: %d\n", fieldName, signedIntValue)
	}
	return signedIntValue, nil
}

func unsafeReadUint8(reader *io.SectionReader) uint8 {
	unsignedIntValue := uint8(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		log.Fatal("Failed to load uint8: ", err)
	}
	return unsignedIntValue
}

func readUint8Safe(reader *io.SectionReader, fieldName string) (uint8, error) {
	if DebugMode {
		debugPrint("    Reading %s...\n", fieldName)
	}
	unsignedIntValue := uint8(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		return 0, fmt.Errorf("failed to read uint8 (%s): %w", fieldName, err)
	}
	if DebugMode {
		debugPrint("    %s: %d\n", fieldName, unsignedIntValue)
	}
	return unsignedIntValue, nil
}

func unsafeReadFloat32(reader *io.SectionReader) float32 {
	floatValue := float32(0)
	if err := binary.Read(reader, binary.LittleEndian, &floatValue); err != nil {
		log.Fatal("Failed to load float32: ", err)
	}
	return floatValue
}

func readFixedList(streamReader *io.SectionReader, listSize int) []byte {
	buffer := make([]byte, listSize)
	if err := binary.Read(streamReader, binary.LittleEndian, &buffer); err != nil {
		log.Fatal("Failed to load buffer: ", err)
	}
	return buffer
}
