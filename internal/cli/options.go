package cli

import (
	"fmt"
	"strings"
)

type commandOptions struct {
	evalCode    string
	printCode   string
	hasEval     bool
	hasPrint    bool
	file        string
	hasFile     bool
	stdin       bool
	arguments   []string
	inputType   string
	check       bool
	interactive bool
	showHelp    bool
	showVersion bool
}

type usageError struct {
	message string
}

func (e *usageError) Error() string { return e.message }

func parseOptions(args []string) (commandOptions, error) {
	var options commandOptions
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if strings.HasPrefix(argument, "--input-type=") {
			value := strings.TrimPrefix(argument, "--input-type=")
			if err := setInputType(&options, value); err != nil {
				return commandOptions{}, err
			}
			continue
		}
		switch argument {
		case "-h", "--help":
			options.showHelp = true
			return options, nil
		case "-v", "--version":
			options.showVersion = true
			return options, nil
		case "-":
			if options.hasEval || options.hasPrint {
				return commandOptions{}, &usageError{message: "only one input source may be specified"}
			}
			options.stdin = true
			options.arguments = append(options.arguments, args[index+1:]...)
			return finishOptions(options)
		case "--":
			if options.hasEval || options.hasPrint {
				options.arguments = append(options.arguments, args[index+1:]...)
				return finishOptions(options)
			}
			if index+1 >= len(args) {
				return finishOptions(options)
			}
			entry := args[index+1]
			if entry == "-" {
				options.stdin = true
			} else {
				options.file = entry
				options.hasFile = true
			}
			options.arguments = append(options.arguments, args[index+2:]...)
			return finishOptions(options)
		case "--check":
			options.check = true
		case "-i", "--interactive":
			options.interactive = true
		case "-e", "--eval":
			if options.hasPrint {
				return commandOptions{}, &usageError{message: "-e/--eval and -p/--print cannot be combined"}
			}
			if index+1 >= len(args) {
				return commandOptions{}, &usageError{message: argument + " requires code"}
			}
			index++
			options.hasEval = true
			options.evalCode = args[index]
		case "--input-type":
			if index+1 >= len(args) {
				return commandOptions{}, &usageError{message: "--input-type requires a value"}
			}
			index++
			if err := setInputType(&options, args[index]); err != nil {
				return commandOptions{}, err
			}
		case "-p", "--print":
			if options.hasEval {
				return commandOptions{}, &usageError{message: "-e/--eval and -p/--print cannot be combined"}
			}
			if index+1 >= len(args) {
				return commandOptions{}, &usageError{message: argument + " requires code"}
			}
			index++
			options.hasPrint = true
			options.printCode = args[index]
		default:
			if (options.hasEval || options.hasPrint) && (argument == "" || argument[0] != '-') {
				options.arguments = append(options.arguments, args[index:]...)
				return finishOptions(options)
			}
			if argument == "" || argument[0] != '-' {
				options.file = argument
				options.hasFile = true
				options.arguments = append(options.arguments, args[index+1:]...)
				return finishOptions(options)
			}
			return commandOptions{}, &usageError{message: fmt.Sprintf("unknown option %q", argument)}
		}
	}
	return finishOptions(options)
}

func finishOptions(options commandOptions) (commandOptions, error) {
	if err := validateOptions(options); err != nil {
		return commandOptions{}, err
	}
	return options, nil
}

func setInputType(options *commandOptions, value string) error {
	if value != "script" && value != "module" {
		return &usageError{message: `--input-type must be "script" or "module"`}
	}
	options.inputType = value
	return nil
}

func validateOptions(options commandOptions) error {
	if options.check && options.interactive {
		return &usageError{message: "--check and -i/--interactive cannot be combined"}
	}
	if options.check && options.hasPrint {
		return &usageError{message: "--check and -p/--print cannot be combined"}
	}
	if options.hasPrint && options.inputType == "module" {
		return &usageError{message: "-p/--print does not support module input"}
	}
	if options.stdin && options.interactive {
		return &usageError{message: "explicit stdin '-' and -i/--interactive cannot be combined"}
	}
	return nil
}
