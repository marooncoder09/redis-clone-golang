package parser

import (
	"fmt"
	"io"
	"os"
)

// ParseRDB opens the RDB file and returns a slice of keys
// from the database section (ignoring metadata).
func ParseRDB(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to open RDB file: %v", err)
	}
	defer file.Close()

	// Read header (9 bytes: "REDIS" + version)
	header := make([]byte, 9)
	if _, err := io.ReadFull(file, header); err != nil {
		return nil, fmt.Errorf("failed to read RDB header: %v", err)
	}
	if string(header[:5]) != "REDIS" {
		return nil, fmt.Errorf("invalid RDB file format")
	}

	var keys []string

	for {
		marker, err := readByte(file)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch marker {
		case 0xFF:
			// End-of-file marker
			return keys, nil
		case 0xFA:
			// Metadata subsection: skip metadata key and value
			_, err := readString(file)
			if err != nil {
				return nil, fmt.Errorf("error reading metadata key: %v", err)
			}
			_, err = readString(file)
			if err != nil {
				return nil, fmt.Errorf("error reading metadata value: %v", err)
			}
		case 0xFE:
			// Database subsection marker.
			// Read database index (size encoded) and ignore it.
			_, err := readSize(file)
			if err != nil {
				return nil, fmt.Errorf("error reading database index: %v", err)
			}
			// Expect hash table size marker: should be 0xFB.
			marker2, err := readByte(file)
			if err != nil {
				return nil, err
			}
			if marker2 != 0xFB {
				return nil, fmt.Errorf("expected hash table size marker 0xFB, got: %x", marker2)
			}
			// Read two size encoded values (hash table sizes) and ignore them.
			_, err = readSize(file)
			if err != nil {
				return nil, fmt.Errorf("error reading hash table size: %v", err)
			}
			_, err = readSize(file)
			if err != nil {
				return nil, fmt.Errorf("error reading keys-with-expiry count: %v", err)
			}

			// Check for an expire marker.
			peek, err := peekByte(file)
			if err != nil {
				return nil, err
			}
			if peek == 0xFC || peek == 0xFD {
				// Consume the expire marker.
				expireMarker, _ := readByte(file)
				if expireMarker == 0xFC {
					// For milliseconds: read 8 bytes.
					expireBytes := make([]byte, 8)
					if _, err := io.ReadFull(file, expireBytes); err != nil {
						return nil, err
					}
				} else { // 0xFD: seconds -> read 4 bytes.
					expireBytes := make([]byte, 4)
					if _, err := io.ReadFull(file, expireBytes); err != nil {
						return nil, err
					}
				}
			}

			// Read the value type marker (assume 0 for string).
			_, err = readByte(file)
			if err != nil {
				return nil, err
			}
			// Read key and value.
			key, err := readString(file)
			if err != nil {
				return nil, fmt.Errorf("error reading key: %v", err)
			}
			// Skip the value.
			_, err = readString(file)
			if err != nil {
				return nil, fmt.Errorf("error reading value: %v", err)
			}

			keys = append(keys, key)
		default:
			return nil, fmt.Errorf("unknown marker: %x", marker)
		}
	}

	return keys, nil
}

// readByte reads a single byte from the file.
func readByte(file *os.File) (byte, error) {
	buf := make([]byte, 1)
	_, err := file.Read(buf)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

// peekByte reads a single byte without advancing the file pointer.
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

// readSize reads a size-encoded value for non-special cases (only supports 0b00 encoding).
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
		// Here in this code i've added the support for both the simple and the special encoded strings
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
		/*
				in the code above i'am using the firstByte & 0x3F to get the encoding type
				because in a string the first 6 bits are used to encode the length of the string
				and the last 2 bits are used to encode the encoding type

				Expl:
				firstByte =110110101
				firstByte & 0x3F → 1101 1010 & 0011 1111
				Result → 0001 1010 (which is 26 in decimal), this means that the encoding type is 26

			Currently I've only implemented the support till the 3 cases but there could be total of 0-63 cases
			because encType is set to firstByte & 0x3F and 0x3F is 0011 1111 in binary (which is 63 in decimal)

			// TODO: Add support for the other cases
		*/
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
