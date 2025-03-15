package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

// ParseRDB will now open the RDB file and returns a map of key-value pairs
// from the database section (ignoring metadata).
func ParseRDB(filePath string) (map[string]core.StoreEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]core.StoreEntry{}, nil
		}
		return nil, fmt.Errorf("failed to open RDB file: %v", err)
	}
	defer file.Close()

	header := make([]byte, 9)
	if _, err := io.ReadFull(file, header); err != nil {
		return nil, fmt.Errorf("failed to read RDB header: %v", err)
	}
	if string(header[:5]) != "REDIS" {
		return nil, fmt.Errorf("invalid RDB file format")
	}

	entries := make(map[string]core.StoreEntry)

	for {
		marker, err := readByte(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch marker {
		case 0xFF:
			return entries, nil
		case 0xFA:
			if _, err := readString(file); err != nil {
				return nil, fmt.Errorf("error reading metadata key: %v", err)
			}
			if _, err := readString(file); err != nil {
				return nil, fmt.Errorf("error reading metadata value: %v", err)
			}

		case 0xFE:
			if _, err := readSize(file); err != nil {
				return nil, fmt.Errorf("error reading database index: %v", err)
			}

			marker2, err := readByte(file)
			if err != nil {
				return nil, err
			}
			if marker2 != 0xFB {
				return nil, fmt.Errorf("expected hash table size marker 0xFB, got: %x", marker2)
			}

			numKeys, err := readSize(file) // this is the total no. of key-value pairs.
			if err != nil {
				return nil, fmt.Errorf("error reading total keys: %v", err)
			}
			if _, err := readSize(file); err != nil { // this is the total no. of keys with expiry.
				return nil, fmt.Errorf("error reading keys-with-expiry count: %v", err)
			}

			for i := 0; i < numKeys; i++ {
				var expiresAt int64 = 0

				peek, err := peekByte(file)
				if err != nil {
					return nil, err
				}
				if peek == 0xFC || peek == 0xFD {
					// Consume the expire marker.
					expireMarker, _ := readByte(file)
					if expireMarker == 0xFC {
						buf := make([]byte, 8)
						if _, err := io.ReadFull(file, buf); err != nil {
							return nil, err
						}
						ts := binary.LittleEndian.Uint64(buf)
						expiresAt = int64(ts)
					} else { // 0xFD indicates seconds
						buf := make([]byte, 4)
						if _, err := io.ReadFull(file, buf); err != nil {
							return nil, err
						}
						ts := binary.LittleEndian.Uint32(buf)
						// convert seconds to milliseconds.
						expiresAt = int64(ts) * 1000
					}
				}

				// read value type marker
				valueType, err := readByte(file)
				if err != nil {
					return nil, err
				}
				if valueType != 0 {
					return nil, fmt.Errorf("unsupported value type: %x", valueType)
				}

				// read key and value as length prefixed strings.
				key, err := readString(file)
				if err != nil {
					return nil, fmt.Errorf("error reading key: %v", err)
				}
				value, err := readString(file)
				if err != nil {
					return nil, fmt.Errorf("error reading value: %v", err)
				}

				entries[key] = core.StoreEntry{Value: value, ExpiresAt: expiresAt}
			}
		default:
			return nil, fmt.Errorf("unknown marker: %x", marker)
		}
	}

	return entries, nil
}

func readByte(file *os.File) (byte, error) {
	buf := make([]byte, 1)
	_, err := file.Read(buf)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func peekByte(file *os.File) (byte, error) {
	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 1)
	_, err = file.Read(buf)
	if err != nil {
		return 0, err
	}
	_, err = file.Seek(pos, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func readSize(file *os.File) (int, error) {
	b, err := readByte(file)
	if err != nil {
		return 0, err
	}
	if b>>6 == 0 {
		return int(b & 0x3F), nil
	}
	return 0, fmt.Errorf("size encoding not supported for byte: %x", b)
}

func readString(file *os.File) (string, error) {
	/*
		Here in this code i've added the support for both the simple and the special encoded strings
	*/

	firstByte, err := readByte(file)
	if err != nil {
		return "", err
	}
	switch firstByte >> 6 {
	case 0:
		length := int(firstByte & 0x3F)
		data := make([]byte, length)
		if _, err := io.ReadFull(file, data); err != nil {
			return "", err
		}
		return string(data), nil
	case 3:
		encType := firstByte & 0x3F

		switch encType {
		case 0: // 8-bit integer ( the reason that this is not little endian is because the endianess is only applicable on the multiple bytes and this is a single byte, by multiple bytes i mean 16,32,64,128 etc)
			bVal, err := readByte(file)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", bVal), nil
		case 1: // 16-bit integer (little-endian) (little endian in simple terms means that the least significant byte is stored first)
			data := make([]byte, 2)
			if _, err := io.ReadFull(file, data); err != nil {
				return "", err
			}
			val := int(data[0]) | int(data[1])<<8
			return fmt.Sprintf("%d", val), nil
		case 2: // 32-bit integer (little-endian)
			data := make([]byte, 4)
			if _, err := io.ReadFull(file, data); err != nil {
				return "", err
			}
			val := int(data[0]) | int(data[1])<<8 | int(data[2])<<16 | int(data[3])<<24
			return fmt.Sprintf("%d", val), nil
		default:
			return "", fmt.Errorf("unsupported special string encoding type: %d", encType)
		}
	default:
		return "", fmt.Errorf("size encoding not supported for byte: %x", firstByte)
	}
}
