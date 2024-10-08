package ft

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func PassCmd(args []string) ([]string, error) {
	cmd := args[1]

	// all commands are expected to lead with "-"
	// the only command that doesn't is a change directory
	// commands that are symbols have this leader added to avoid logic
	// meant for a directory path
	switch cmd {
	case "]", "[", "..", "-":
		cmd = fmt.Sprintf("-%s", cmd)
	default:
		break
	}

	// keys and key evals are given a leader of "_"
	// to help identify the appropriate function
	// in the map
	if string(cmd[0]) != "-" {
		cmd = "_"
		parsedCmd := make([]string, 3)
		parsedCmd[0] = args[0]
		parsedCmd[1] = cmd
		parsedCmd[2] = args[1]
		args = parsedCmd
	}

	_, ok := AvailCmds[cmd]
	if !ok {
		return nil, errors.New(fmt.Sprintf("%v is not a valid command, use 'ft -h' or 'ft -help' for a list of valid commands", cmd))
	}

	// verify user provided correct minimum number of arguments
	// too many args will work, any args beyond expected number are simply ignored
	switch cmd {
	case "-ls", "-]", "-[", "-..", "--":
		return []string{cmd}, nil
	// providing help for a specific command may be needed in the future
	case "-help", "-h", "-version", "-v", "-is":
		break
	case "-rn":
		if len(args) <= 3 {
			return nil, errors.New(fmt.Sprintf("Insufficient args provided %v, usage: ft rn <key> <newKey>", args[1:]))
		}
	default:
		if len(args) <= 2 {
			return nil, errors.New(fmt.Sprintf("Insufficient args provided %v, usage: ft <command> <path/key>", args[1:]))
		}
	}

	// return args without 'ft'
	return args[1:], nil
}

func changeDirectory(data *CmdArgs) error {
	if len(data.allPaths) == 0 {
		fmt.Print(NoLocationsSetMsg)
		return nil
	}

	var key string
	provided_string := data.cmd[1]

	if strings.Contains(provided_string, "/") {

		path_array := strings.Split(provided_string, "/")
		eval_array := make([]string, len(path_array))

		// key evaluation if the first string before "/" delimeter is key
		key = path_array[0]
		if p, ok := data.allPaths[key]; ok {
			eval_array[0] = p
		} else {
			eval_array[0] = key
		}
		for i, str := range path_array {
			if i != 0 {
				eval_array[i] = str
			}
		}

		// handles evaluated path and relative paths
		path := strings.Join(eval_array, "/")
		dir, err := os.Stat(path)
		if err == nil {
			if dir.IsDir() {
				fmt.Println(path)
				return nil
			}
		}

		fmt.Printf(InvalidDirectoryMsg, provided_string, path)
		return nil

	} else {

		key = provided_string
		p, ok := data.allPaths[key]
		if !ok {
			fmt.Printf(UnrecognizedKeyMsg, key, key, key)
			return nil
		}

		fmt.Println(p)
		return nil

	}
}

func setDirectoryVar(data *CmdArgs) error {
	key := data.cmd[1]
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	for k, v := range data.allPaths {
		if path == v {
			fmt.Printf(PathAlreadyExistsMsg, path, k, k)
			var res string
			_, err := fmt.Fscan(data.rdr, &res)
			if err != nil {
				return err
			}
			if overwrite, err := verifyInput(res); !overwrite {
				if err != nil {
					return err
				}
				fmt.Printf(AbortedOverwriteKeyMsg, key)
				return nil
			} else {
				if err != nil {
					return err
				}
			}
			delete(data.allPaths, k)
			data.allPaths[key] = path
			dataUpdate(data.allPaths, data.file)
			fmt.Printf(RenamedKeyMsg, k, key, path)
			return nil
		}
	}

	val, ok := data.allPaths[key]
	if ok {
		// capture user response and act accordingly
		fmt.Printf(KeyAlreadyExistsMsg, key, val, key)
		var res string
		_, err := fmt.Fscan(data.rdr, &res)
		if err != nil {
			return err
		}
		if overwrite, err := verifyInput(res); !overwrite {
			if err != nil {
				return err
			}
			fmt.Printf(AbortedOverwriteKeyMsg, key)
			return nil
		} else {
			if err != nil {
				return err
			}
		}

	}

	// key doesn't exist yet or user wants to overwrite
	data.allPaths[key] = path
	dataUpdate(data.allPaths, data.file)
	fmt.Printf(AddKeyMsg, key, path)
	return nil
}

func displayAllPaths(data *CmdArgs) error {
	printMap(data.allPaths)
	return nil
}

func removeKey(data *CmdArgs) error {
	var res string
	key := data.cmd[1]

	_, ok := data.allPaths[key]
	if !ok {
		fmt.Printf(KeyDoesNotExistMsg, key)
		return nil
	}
	fmt.Printf(VerifyRemoveMsg, key)
	_, err := fmt.Fscan(data.rdr, &res)
	if err != nil {
		return err
	}
	if rm, err := verifyInput(res); !rm {
		if err != nil {
			return err
		}
		fmt.Printf(AbortRemoveKeyMsg, key)
		return nil
	} else {
		if err != nil {
			return err
		}
	}
	delete(data.allPaths, key)
	dataUpdate(data.allPaths, data.file)
	fmt.Printf(RemoveKeyMsg, key)
	return nil
}

func renameKey(data *CmdArgs) error {
	originalKey := data.cmd[1]
	newKey := data.cmd[2]

	_, ok := data.allPaths[newKey]
	if ok {
		fmt.Printf(RenameKeyAlreadyExistsMsg, newKey)
		return nil
	}
	path, ok := data.allPaths[originalKey]
	if !ok {
		fmt.Printf(RenameKeyDoesNotExistMsg, originalKey)
		return nil
	}

	var res string

	fmt.Printf(VerifyRenameMsg, originalKey, newKey)
	_, err := fmt.Fscan(data.rdr, &res)
	if err != nil {
		return err
	}
	if rm, err := verifyInput(res); !rm {
		if err != nil {
			fmt.Println("Error: ", err)
		}
		fmt.Printf(AbortRenameKeyMsg, originalKey, newKey)
		return nil
	} else {
		if err != nil {
			return err
		}
	}
	delete(data.allPaths, originalKey)
	data.allPaths[newKey] = path

	dataUpdate(data.allPaths, data.file)
	fmt.Printf(RenamedKeyMsg, originalKey, newKey, path)
	return nil
}

func showHelp(data *CmdArgs) error {
	printMap(CmdDesc)
	return nil
}

func showVersion(data *CmdArgs) error {
	fmt.Print(Logo)
	fmt.Println("version:\t", Version)
	return nil
}

func showDirectoryVar(data *CmdArgs) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	pathMaps := data.allPaths

	for k := range pathMaps {
		v, _ := pathMaps[k]

		if strings.Compare(v, dir) == 0 {
			fmt.Printf(IsKeyMsg, dir, k)
			return nil
		}
	}

	fmt.Printf(IsNotKeyMsg, dir)
	return nil
}

// Used for commands that are simply handled by the shell function
func passToShell(data *CmdArgs) error {
	c := data.cmd[0]
	command := string(c[1:])

	switch command {
	case "]", "[", "..", "-":
		fmt.Println(command)
	default:
		return errors.New(fmt.Sprintf("Tried to pass command to shell, but '%s' is not a valid command for the shell function.", command))
	}

	return nil
}
