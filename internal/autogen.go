/*
Copyright © 2013 the InMAP authors.
This file is part of InMAP.

InMAP is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

InMAP is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with InMAP.  If not, see <http://www.gnu.org/licenses/>.
*/

//+build ignore

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/spatialmodel/inmap"
	"github.com/spatialmodel/inmap/inmaputil"
	"github.com/spatialmodel/inmap/science/chem/simplechem"
	"github.com/spf13/cobra/doc"
)

func main() {

	// Generate documentation for the available commands.
	doc.GenMarkdownTree(inmaputil.Root, "./inmap/doc/")

	writeOutputOptions()
}

type config struct {
	InMAPData        string
	VariableGridData string
	VarGrid          inmap.VarGridConfig
}

func loadConfig(file string) (*config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("problem reading configuration file: %v", err)
	}

	cfg := new(config)
	_, err = toml.Decode(string(bytes), cfg)
	if err != nil {
		return nil, fmt.Errorf(
			"there has been an error parsing the configuration file: %v\n", err)
	}

	cfg.InMAPData = os.ExpandEnv(cfg.InMAPData)
	cfg.VariableGridData = os.ExpandEnv(cfg.VariableGridData)
	cfg.VarGrid.CensusFile = os.ExpandEnv(cfg.VarGrid.CensusFile)
	cfg.VarGrid.MortalityRateFile = os.ExpandEnv(cfg.VarGrid.MortalityRateFile)

	return cfg, err
}

// writeOutputOptions creates a list of output options in markdown format.
func writeOutputOptions() {

	cfg, err := loadConfig("inmap/configExample.toml")
	if err != nil {
		log.Fatal(err)
	}

	r, err := os.Open(cfg.VariableGridData)
	if err != nil {
		log.Fatal(err)
	}

	var m simplechem.Mechanism
	d := &inmap.InMAP{
		InitFuncs: []inmap.DomainManipulator{
			inmap.Load(r, &cfg.VarGrid, inmap.NewEmissions(), m),
		},
	}
	if err = d.Init(); err != nil {
		log.Fatal(err)
	}
	names, descriptions, units := d.OutputOptions(m)

	f, err := os.Create("doc/OutputOptions.md")
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(([]byte)("# InMAP output options\n\nThis is a list of InMAP output options that can be" +
		" included in the `OutputOptions` configuration variable.\n\n" +
		"This file is automatically generated; do not edit.\n\n"))
	if err != nil {
		log.Fatal(err)
	}

	for i, n := range names {
		s := fmt.Sprintf("* `%s`: %s [%s]\n", n, descriptions[i], units[i])
		_, err = f.Write(([]byte)(s))
		if err != nil {
			log.Fatal(err)
		}
	}

	f.Close()
}
