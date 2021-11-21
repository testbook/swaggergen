package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"

	"swaggergen/codescan"
)

type list []string

func (i *list) String() string {
	return ""
}

func (i *list) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	workDir     = flag.String("w", ".", "base path")
	buildTags   = flag.String("t", "", "build tags")
	scanModels  = flag.Bool("m", false, "includes models that were annotated with 'swagger:model'")
	compact     = flag.Bool("c", false, "don't prettify output json")
	outputFile  = flag.String("o", "", "output file")
	inputFile   = flag.String("i", "", "an input swagger file with which to merge")
	excludeDeps = flag.Bool("ed", false, "exclude all dependencies of project")
	include     list
	exclude     list
	includeTags list
	excludeTags list
)

func init() {
	flag.Var(&include, "ip", "include packages matching pattern")
	flag.Var(&exclude, "ep", "exclude packages matching pattern")
	flag.Var(&includeTags, "it", "include routes having specified tags")
	flag.Var(&excludeTags, "et", "exclude routes having specified tags")
	flag.Parse()
}

func main() {
	// by default consider all the paths under the working directory
	packages := []string{"./..."}
	if len(flag.Args()) > 1 {
		packages = flag.Args()
	}

	input, err := loadSpec(*inputFile)
	if err != nil {
		panic(err)
	}

	opts := &codescan.Options{
		Packages:    packages,
		WorkDir:     *workDir,
		InputSpec:   input,
		ScanModels:  *scanModels,
		BuildTags:   *buildTags,
		ExcludeDeps: *excludeDeps,
		Include:     include,
		Exclude:     exclude,
		IncludeTags: includeTags,
		ExcludeTags: excludeTags,
	}
	spec, err := codescan.Run(opts)
	if err != nil {
		panic(err)
	}

	err = write(spec, *compact, *outputFile)
	if err != nil {
		panic(err)
	}
}

func loadSpec(input string) (s *spec.Swagger, err error) {
	if input == "" {
		return nil, nil
	}
	fi, err := os.Stat(input)
	if err != nil {
		return
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("expected %q to be a file not a directory", input)
	}
	sp, err := loads.Spec(input)
	if err != nil {
		return nil, err
	}
	return sp.Spec(), nil
}

func write(spec *spec.Swagger, compact bool, output string) (err error) {
	f := os.Stdout
	if output != "" {
		f, err = os.Create(output)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	enc := json.NewEncoder(f)
	if !compact {
		enc.SetIndent("", "	")
	}
	return enc.Encode(spec)
}
