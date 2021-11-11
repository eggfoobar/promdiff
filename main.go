package main

import (
	"flag"
	"fmt"
	"os"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/eggfoobar/promdiff/pkg/config"
	"github.com/eggfoobar/promdiff/pkg/prom"
	"github.com/eggfoobar/promdiff/pkg/templates"
)

var (
	configFile   string
	outputFile   string
	outputFormat string
	templateFile string
	lineWidth    int
	leftPadding  int
)

func init() {
	flag.StringVar(&configFile, "c", "prom.yaml", "config file to use")
	flag.StringVar(&outputFormat, "o", "markdown", "output format (ex. markdown, jira, text)")
	flag.StringVar(&outputFile, "f", "", "output file to save data (default: stdout)")
	flag.StringVar(&templateFile, "t", "", "custom template file to use")
	flag.IntVar(&lineWidth, "w", 90, "line width for markdown print out (only used when printing to markdown stdout)")
	flag.IntVar(&leftPadding, "lp", 0, "left hand padding markdown print out (only used when printing to markdown stdout)")
	flag.Parse()
}

func main() {

	conf, err := config.NewConfig(configFile)
	if err != nil {
		panic(err)
	}

	results, err := prom.FetchData(conf)
	if err != nil {
		panic(fmt.Errorf("failed to get data for target (%s)", err))
	}

	var output templates.OutputTemplate
	if templateFile != "" {
		output, err = templates.ParseWithTemplateFile(templateFile, results)
	} else {
		output, err = templates.Parse(outputFormat, results)
	}
	if err != nil {
		panic(err)
	}

	printout(output, outputFile)
}

func printout(output templates.OutputTemplate, outputFile string) error {
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(output.ParsedTemplate), os.ModePerm)
		if err != nil {
			return err
		}
		return nil
	}
	printOut := output.ParsedTemplate
	if output.Type == templates.Markdown {
		result := markdown.Render(output.ParsedTemplate, lineWidth, leftPadding)
		printOut = string(result)
	}

	fmt.Println(printOut)

	return nil
}
