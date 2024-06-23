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

func ApplyExitOnHelp(cmd *cobra.Command) {
	helpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		helpFunc(cmd, args)
		os.Exit(0)
	})
}

var errStrings []string

func main() {
	log.SetFlags(0)

	rootCmd := &cobra.Command{
		Use:   "go run main.go",
		Short: "Path to config",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	ApplyExitOnHelp(rootCmd)

	confPathDef := "config/config.yaml"

	var confPathArg, filesGlobArg, filesGlobExcludeArg, messageArg, filesPatternArg, templateData string
	rootCmd.Flags().StringVarP(&confPathArg, "config", "c", confPathDef, "Сonfig path")
	rootCmd.Flags().StringVar(&filesGlobArg, "filesglob", "**", "Files glob to check")
	rootCmd.Flags().StringVar(&filesGlobExcludeArg, "filesglobexclude", "", "Files glob to exclude")
	rootCmd.Flags().StringVar(&messageArg, "message", "", "Message to print")
	rootCmd.Flags().StringVar(&filesPatternArg, "filespattern", "", "Files pattern to check")
	rootCmd.Flags().StringVar(&templateData, "templatedata", "", "Key value pairs for template")

	if err := rootCmd.Execute(); err != nil {
		errStrings = append(errStrings, err.Error())
		logformatter.LogError(errStrings)
	}

	var config *conf.Conf
	var pattern *conf.Pattern
	var err error
	var patternFile []byte

	if filesGlobArg != "**" || filesGlobExcludeArg != "" {
		filesGlob := []string{}

		filesGlob = append(filesGlob, strings.Split(filesGlobArg, " ")...)

		if filesPatternArg != "" {
			patternFile, err = os.ReadFile(filesPatternArg)
		} else {
			errStrings = append(errStrings, "A variable 'filesPatternArg' with arguments must be declared")
			logformatter.LogError(errStrings)
		}

		if err != nil {
			errStrings = append(errStrings, err.Error())
			logformatter.LogError(errStrings)
		}

		var data map[string]interface{}
		if templateData != "" {
			err = json.Unmarshal([]byte(templateData), &data)
			if err != nil {
				errStrings = append(errStrings, err.Error())
				logformatter.LogError(errStrings)
			}
		}

		filesGlobExclude := []string{}
		filesGlobExclude = append(filesGlobExclude, strings.Split(filesGlobExcludeArg, " ")...)
		patternFileStr := string(patternFile)

		pattern = &conf.Pattern{
			Literal: patternFileStr,
		}

		config = &conf.Conf{
			{
				ID:      "example_id",
				Message: messageArg,
				Glob:    filesGlob,
				Exclude: filesGlobExclude,
				Pattern: *pattern,
				Vars:    data,
			},
		}
	} else {
		config, err = conf.GetConf(confPathArg)
	}

	if err != nil {
		errStrings = append(errStrings, err.Error())
		logformatter.LogError(errStrings)
	}

	err = conf.CheckConfFields(*config)
	if err != nil {
		errStrings = append(errStrings, err.Error())
		logformatter.LogError(errStrings)
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

		patternTemplate, err := templating.PrepareTeamplate(pattern, varsTemplate)
		if err != nil {
			errStrings = append(errStrings, err.Error())
			logformatter.LogError(errStrings)
		}

		err = matcher.MatchLicenseText(filesGlob, filesGlobExclude, patternTemplate, errMessage)
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}

	}

	if len(errStrings) > 0 {
		logformatter.LogError(errStrings)
	}

	log.Println("All files successfully matched with the license")

}
