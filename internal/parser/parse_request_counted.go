// internal/parser/parse_request_counted.go
package parser

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
)

func ParseRequestWithByteCount(reader *bufio.Reader) ([]string, int64, error) {
	var totalBytes int64 = 0

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, 0, err
	}
	totalBytes += int64(len(line))

	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] != '*' {
		return nil, totalBytes, errors.New("invalid request format")
	}

	argCount, err := strconv.Atoi(line[1:])
	if err != nil || argCount < 1 {
		return nil, totalBytes, errors.New("invalid argument count")
	}

	args := make([]string, 0, argCount)

	for i := 0; i < argCount; i++ {

		lengthLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, totalBytes, errors.New("unexpected end of input (length line)")
		}
		totalBytes += int64(len(lengthLine))

		argLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, totalBytes, errors.New("unexpected end of input (arg line)")
		}
		totalBytes += int64(len(argLine))

		arg := strings.TrimSpace(argLine)
		args = append(args, arg)
	}

	return args, totalBytes, nil
}
