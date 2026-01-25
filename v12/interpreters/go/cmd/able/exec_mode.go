package main

import (
	"fmt"
	"strings"
)

type interpreterMode string

const (
	interpreterTreewalker interpreterMode = "treewalker"
	interpreterBytecode   interpreterMode = "bytecode"
)

func parseExecMode(args []string) (interpreterMode, []string, error) {
	mode := interpreterTreewalker
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			remaining = append(remaining, args[i:]...)
			break
		}
		switch {
		case arg == "--exec-mode":
			if i+1 >= len(args) {
				return mode, nil, fmt.Errorf("--exec-mode expects a value")
			}
			next := args[i+1]
			parsed, err := parseExecModeValue(next)
			if err != nil {
				return mode, nil, err
			}
			mode = parsed
			i++
		case strings.HasPrefix(arg, "--exec-mode="):
			value := strings.TrimPrefix(arg, "--exec-mode=")
			parsed, err := parseExecModeValue(value)
			if err != nil {
				return mode, nil, err
			}
			mode = parsed
		default:
			remaining = append(remaining, arg)
		}
	}
	return mode, remaining, nil
}

func parseExecModeValue(value string) (interpreterMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return interpreterTreewalker, fmt.Errorf("--exec-mode expects a value")
	case string(interpreterTreewalker):
		return interpreterTreewalker, nil
	case string(interpreterBytecode):
		return interpreterBytecode, nil
	default:
		return interpreterTreewalker, fmt.Errorf("unknown --exec-mode value '%s' (expected treewalker or bytecode)", value)
	}
}
