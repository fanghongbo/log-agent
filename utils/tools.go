package utils

import "os"

func CheckPath(path string) (bool, error) {
	var err error

	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
