package ft

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

func ReadMap(file *os.File) (map[string]string, error) {

	pathMap := make(map[string]string)

	fileInfo, err := file.Stat()
	if err != nil {
		return pathMap, errors.New(fmt.Sprint("Error getting file info: ", err))
	}

	// take file size in bytes and make a buffer of that size
	size := fileInfo.Size()
	buff := make([]byte, size)

	// read entire file into memory
	_, err = file.Read(buff)
	if err != nil {
		return pathMap, errors.New(fmt.Sprint("Error reading file into buffer: ", err))
	}

	// key length integer should always fit in 1 byte
	var keyLen uint8
	// value length integer should always fit in 2 bytes
	var valLen uint16
	// sliding pointer to navigate buffer
	var offset uint

	// iterate through buffer and deserialize
	for offset < uint(len(buff)) {

		// read length of key, use length to read in key, adjust offset
		// simple type conversion since length is only 1 byte and not a []byte
		keyLen = uint8(buff[offset])
		offset++
		kl := uint(keyLen)
		keyBytes := buff[offset : offset+kl]
		offset += kl

		// read length of value, use length to read in value, adjust offset
		// length contained in 2 bytes, need to convert []byte to a uint16 value
		valLen = binary.LittleEndian.Uint16(buff[offset : offset+2])
		offset += 2
		vl := uint(valLen)
		valBytes := buff[offset : offset+vl]
		offset += vl
		// add key-value to map
		pathMap[string(keyBytes)] = string(valBytes)

	}

	return pathMap, nil

}

func EnsureData(filepath string) (*os.File, error) {

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error opening file: ", err))
	}

	return file, nil

}

func dataUpdate(hashmap map[string]string, file *os.File) error {

	var buffer []byte
	for key, val := range hashmap {

		keyBytes := []byte(key)
		valBytes := []byte(val)

		keyLen := make([]byte, 1)
		keyLen[0] = byte(uint8(len(keyBytes)))

		valLen := make([]byte, 2)
		binary.LittleEndian.PutUint16(valLen, uint16(len(valBytes)))

		// create an array of array of bytes for optimal concatenation
		allBytes := [][]byte{keyLen, keyBytes, valLen, valBytes}

		// get the length for key-value pair to allocate memory
		var pairLen int
		for _, s := range allBytes {
			pairLen += len(s)
		}
		// create new slice and append all []byte from allBytes
		pair := make([]byte, pairLen)
		var i int
		for _, s := range allBytes {
			i += copy(pair[i:], s)
		}
		// append completed pair to buffer []byte
		allPairs := make([]byte, len(buffer)+len(pair))
		copy(allPairs, buffer)
		copy(allPairs[len(buffer):], pair)
		buffer = allPairs
	}

	err := file.Truncate(0)
	if err != nil {
		return errors.New(fmt.Sprint("Error truncating file: ", err))
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return errors.New(fmt.Sprint("Error seeking to beginning of file: ", err))
	}

	_, err = file.Write(buffer)
	if err != nil {
		return errors.New(fmt.Sprint("Error writing contents of buffer to file: ", err))
	}

	return nil
}
