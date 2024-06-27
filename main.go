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

func applyExitOnHelp(cmd *cobra.Command) {
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
	applyExitOnHelp(rootCmd)

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

	var config *conf.Conf
	var pattern *conf.Pattern
	var err error
	var patternFile []byte

	if fileGlobPattern != "**" || fileGlobExcludePattern != "" {
		filesGlob := []string{}

		filesGlob = append(filesGlob, strings.Split(fileGlobPattern, " ")...)

		if filePatternPath != "" {
			patternFile, err = os.ReadFile(filePatternPath)
		} else {
			errorLogMessages = append(errorLogMessages, "A variable 'filePatternPath' with arguments must be declared")
			logformatter.LogError(errorLogMessages)
		}

		if err != nil {
			errorLogMessages = append(errorLogMessages, err.Error())
			logformatter.LogError(errorLogMessages)
		}

		var data map[string]interface{}
		if templateDataJSON != "" {
			err = json.Unmarshal([]byte(templateDataJSON), &data)
			if err != nil {
				errorLogMessages = append(errorLogMessages, err.Error())
				logformatter.LogError(errorLogMessages)
			}
		}

		filesGlobExclude := []string{}
		filesGlobExclude = append(filesGlobExclude, strings.Split(fileGlobExcludePattern, " ")...)
		patternFileStr := string(patternFile)

		pattern = &conf.Pattern{
			Literal: patternFileStr,
		}

		config = &conf.Conf{
			{
				ID:      "example_id",
				Message: errorMessageText,
				Glob:    filesGlob,
				Exclude: filesGlobExclude,
				Pattern: *pattern,
				Vars:    data,
			},
		}
	} else {
		config, err = conf.GetConf(configPath)
	}

	if err != nil {
		errorLogMessages = append(errorLogMessages, err.Error())
		logformatter.LogError(errorLogMessages)
	}

	err = conf.CheckConfFields(*config)
	if err != nil {
		errorLogMessages = append(errorLogMessages, err.Error())
		logformatter.LogError(errorLogMessages)
	}

	for _, c := range *config {
		filesGlob := c.Glob
		filesGlobExclude := c.Exclude
		errMessage := c.Message
		varsTemplate := c.Vars

		if errMessage == "" {
			errMessage = "Copyright text not matching"
		}
		pattern := c.Pattern.Literal


		expectedText, err := templating.PrepareTemplatedText(pattern, varsTemplate)
		if err != nil {
			errorLogMessages = append(errorLogMessages, err.Error())
			logformatter.LogError(errorLogMessages)
		}

		err = matcher.MatchLicenseText(filesGlob, filesGlobExclude, expectedText, errMessage)
		if err != nil {
			errorLogMessages = append(errorLogMessages, err.Error())
		}

	}

	if len(errorLogMessages) > 0 {
		logformatter.LogError(errorLogMessages)
	}

	log.Println("All files successfully matched with the license")

}
