package parser

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
)

func ParseRequest(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		if "EOF" == err.Error() {
			return nil, nil // Returning nil in case of EOF TODO: Check what is the correct way to handle this
		}
		return nil, err
	}

	line = strings.TrimSpace(line)

	if len(line) == 0 || line[0] != '*' {
		return nil, errors.New("invalid request format")
	}

	argCount, err := strconv.Atoi(line[1:])
	if err != nil || argCount < 1 {
		return nil, errors.New("invalid argument count")
	}

	args := make([]string, 0, argCount)

	for i := 0; i < argCount; i++ {
		_, err := reader.ReadString('\n') // Ignore length line
		if err != nil {
			return nil, errors.New("unexpected end of input")
		}

		arg, err := reader.ReadString('\n')
		if err != nil {
			return nil, errors.New("unexpected end of input")
		}

		args = append(args, strings.TrimSpace(arg))
	}

	return args, nil
}
