package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/elbars/copyright_checker/internal/conf"
	"github.com/elbars/copyright_checker/internal/logformatter"
	"github.com/elbars/copyright_checker/internal/matcher"
	"github.com/elbars/copyright_checker/internal/templating"
	"github.com/spf13/cobra"
)

func wrapHelpFunctionWithExit(cmd *cobra.Command) {
	helpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		helpFunc(cmd, args)
		os.Exit(0)
	})
}

var errorLogMessages []string

func main() {
	log.SetFlags(0)

	rootCmd := &cobra.Command{
		Use:   "go run main.go",
		Short: "Path to config",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	wrapHelpFunctionWithExit(rootCmd)

	defaultConfigPath := "config/config.yaml"

	var configPath, fileGlobPattern, fileGlobExcludePattern, errorMessageText, filePatternPath, templateDataJSON string
	rootCmd.Flags().StringVarP(&configPath, "config", "c", defaultConfigPath, "Сonfig path")
	rootCmd.Flags().StringVar(&fileGlobPattern, "filesglob", "**", "Files glob to check")
	rootCmd.Flags().StringVar(&fileGlobExcludePattern, "filesglobexclude", "", "Files glob to exclude")
	rootCmd.Flags().StringVar(&errorMessageText, "message", "", "Message to print")
	rootCmd.Flags().StringVar(&filePatternPath, "filespattern", "", "Files pattern to check")
	rootCmd.Flags().StringVar(&templateDataJSON, "templatedata", "", "Key value pairs for template")

	if err := rootCmd.Execute(); err != nil {
		errorLogMessages = append(errorLogMessages, err.Error())
		logformatter.LogError(errorLogMessages)
	}

	var configuration *conf.Conf
	var filePattern *conf.Pattern
	var err error
	var patternFileBytes []byte

	if fileGlobPattern != "**" || fileGlobExcludePattern != "" {
		fileGlobPatterns := []string{}

		fileGlobPatterns = append(fileGlobPatterns, strings.Split(fileGlobPattern, " ")...)

		if filePatternPath != "" {
			patternFileBytes, err = os.ReadFile(filePatternPath)
		} else {
			errorLogMessages = append(errorLogMessages, "A variable 'filePatternPath' with arguments must be declared")
			logformatter.LogError(errorLogMessages)
		}

		if err != nil {
			errorLogMessages = append(errorLogMessages, err.Error())
			logformatter.LogError(errorLogMessages)
		}

		var templateData map[string]interface{}
		if templateDataJSON != "" {
			err = json.Unmarshal([]byte(templateDataJSON), &templateData)
			if err != nil {
				errorLogMessages = append(errorLogMessages, err.Error())
				logformatter.LogError(errorLogMessages)
			}
		}

		fileGlobExcludePatterns := []string{}
		fileGlobExcludePatterns = append(fileGlobExcludePatterns, strings.Split(fileGlobExcludePattern, " ")...)
		patternFileStr := string(patternFileBytes)

		filePattern = &conf.Pattern{
			Literal: patternFileStr,
		}

		configuration = &conf.Conf{
			{
				ID:      "example_id",
				Message: errorMessageText,
				Glob:    fileGlobPatterns,
				Exclude: fileGlobExcludePatterns,
				Pattern: *filePattern,
				Vars:    templateData,
			},
		}
	} else {
		configuration, err = conf.GetConf(configPath)
	}

	if err != nil {
		errorLogMessages = append(errorLogMessages, err.Error())
		logformatter.LogError(errorLogMessages)
	}

	err = conf.CheckConfFields(*configuration)
	if err != nil {
		errorLogMessages = append(errorLogMessages, err.Error())
		logformatter.LogError(errorLogMessages)
	}

	for _, configItem := range *configuration {
		fileGlobPatterns := configItem.Glob
		fileGlobExcludePatterns := configItem.Exclude
		errorMessageText := configItem.Message
		templateData := configItem.Vars

		if errorMessageText == "" {
			errorMessageText = "Copyright text not matching"
		}
		filePatternLiteral := configItem.Pattern.Literal

		expectedText, err := templating.PrepareTemplatedText(filePatternLiteral, templateData)
		if err != nil {
			errorLogMessages = append(errorLogMessages, err.Error())
			logformatter.LogError(errorLogMessages)
		}

		err = matcher.MatchLicenseText(fileGlobPatterns, fileGlobExcludePatterns, expectedText, errorMessageText)
		if err != nil {
			errorLogMessages = append(errorLogMessages, err.Error())
		}

	}

	if len(errorLogMessages) > 0 {
		logformatter.LogError(errorLogMessages)
	}

	log.Println("All files successfully matched with the license")

}
