package matcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v2"
)

func MatchLicenseText(filesGlob []string, filesGlobExclude []string, copyrightString string, errMessage string) error {
	var filesToCheck []string

	for _, fileGlob := range filesGlob {
		filesMatched, err := doublestar.Glob(fileGlob)
		if err != nil {
			return formatError(err)
		}
		filesToCheck = append(filesToCheck, filesMatched...)
	}

	var filesToExclude []string

	for _, fileGlobExclude := range filesGlobExclude {
		filesExcludeMatched, err := doublestar.Glob(fileGlobExclude)
		if err != nil {
			return formatError(err)
		}
		filesToExclude = append(filesToExclude, filesExcludeMatched...)
	}

	var filesNoMatchedPattern []string

	for _, file := range filesToCheck {
		if f, err := os.Stat(file); err != nil {
			return formatError(err)
		} else {
			if f.IsDir() {
				continue
			}
		}

		if slices.Contains(filesToExclude, file) {
			continue
		}

		fileText, err := os.ReadFile(file)
		if err != nil {
			return formatError(err)
		}

		if strings.Contains(string(fileText), copyrightString) {
			err = nil
		} else {
			filesNoMatchedPattern = append(filesNoMatchedPattern, file)
		}
	}

	if len(filesNoMatchedPattern) > 0 {
		err := errors.New("\n" + errMessage + " (in " + strconv.Itoa(len(filesNoMatchedPattern)) + " files): \n[" + strings.Join(filesNoMatchedPattern, " ") + "]\n")
		return formatError(err)
	}

	return nil
}

func formatError(err error) error {
	if err != nil {
		// Получаем информацию о файле и строке, где произошла ошибка
		_, file, line, _ := runtime.Caller(1)

		// Получаем текущий рабочий каталог
		dir, dirErr := os.Getwd()
		if dirErr != nil {
			return fmt.Errorf("%s:%d:\n%w", file, line, err)
		}

		// Получаем относительный путь файла
		relFile, relErr := filepath.Rel(dir, file)
		if relErr != nil {
			return fmt.Errorf("%s:%d:\n%w", file, line, err)
		}

		return fmt.Errorf("%s:%d: %w", relFile, line, err)
	}
	return err
}
