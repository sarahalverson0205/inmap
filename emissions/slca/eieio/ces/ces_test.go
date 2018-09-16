/*
Copyright © 2017 the InMAP authors.
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
along with InMAP.  If not, see <http://www.gnu.org/licenses/>.*/

package ces_test

import (
	"context"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gonum/floats"
	"github.com/spatialmodel/inmap/emissions/slca/eieio"
	"github.com/spatialmodel/inmap/emissions/slca/eieio/ces"
	"github.com/spatialmodel/inmap/emissions/slca/eieio/eieiorpc"
	"github.com/spatialmodel/inmap/epi"
)

func TestCES(t *testing.T) {
	f, err := os.Open("../data/test_config.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var cfg eieio.ServerConfig
	_, err = toml.DecodeReader(f, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Config.Years = []eieio.Year{2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2013, 2014, 2015}

	s, err := eieio.NewServer(&cfg, epi.NasariACS)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	c, err := ces.NewCES(s)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("demand", func(t *testing.T) {
		d, err := c.DemographicConsumption(context.Background(), &eieiorpc.DemographicConsumptionInput{
			Demograph: eieiorpc.Demograph_WhiteOther,
			Year:      2014,
		})
		if err != nil {
			t.Fatal(err)
		}
		haveWhiteOther := floats.Sum(d.Data)
		want := 1.1494028913920795e+13
		if haveWhiteOther != want {
			t.Errorf("white/other = %g; want %g", haveWhiteOther, want)
		}

		d, err = c.DemographicConsumption(context.Background(), &eieiorpc.DemographicConsumptionInput{
			Demograph: eieiorpc.Demograph_Black,
			Year:      2014,
		})
		if err != nil {
			t.Fatal(err)
		}
		haveBlack := floats.Sum(d.Data)
		want = 1.2121719653466628e+12
		if haveBlack != want {
			t.Errorf("black = %g; want %g", haveBlack, want)
		}

		d, err = c.DemographicConsumption(context.Background(), &eieiorpc.DemographicConsumptionInput{
			Demograph: eieiorpc.Demograph_Hispanic,
			Year:      2014,
		})
		if err != nil {
			t.Fatal(err)
		}
		haveHispanic := floats.Sum(d.Data)
		want = 1.5296507939512983e+12
		if haveHispanic != want {
			t.Errorf("hispanic = %g; want %g", haveHispanic, want)
		}

		d, err = c.DemographicConsumption(context.Background(), &eieiorpc.DemographicConsumptionInput{
			Demograph: eieiorpc.Demograph_All,
			Year:      2014,
		})
		if err != nil {
			t.Fatal(err)
		}
		haveAll := floats.Sum(d.Data)
		allSum := haveBlack + haveHispanic + haveWhiteOther

		if !floats.EqualWithinAbsOrRel(haveAll, allSum, 1e-1, 1e-1) {
			t.Errorf("total demographic: %g != %g", haveAll, allSum)
		}

		var overallTotal float64
		for _, dt := range []eieiorpc.FinalDemandType{
			eieiorpc.FinalDemandType_PersonalConsumption,
			eieiorpc.FinalDemandType_PrivateResidential,
			eieiorpc.FinalDemandType_PrivateStructures,
			eieiorpc.FinalDemandType_PrivateEquipment,
			eieiorpc.FinalDemandType_PrivateIP,
			eieiorpc.FinalDemandType_InventoryChange} {

			d, err := s.FinalDemand(context.TODO(), &eieiorpc.FinalDemandInput{
				FinalDemandType: dt,
				Year:            2014,
				Location:        eieiorpc.Location_Domestic,
			})
			if err != nil {
				t.Fatal(err)
			}
			overallTotal += floats.Sum(d.Data)
		}

		if !floats.EqualWithinAbsOrRel(haveAll, overallTotal, 1e-8, 1e-8) {
			t.Errorf("overall total: %g != %g", haveAll, overallTotal)
		}

		for _, year := range cfg.Config.Years {
			_, err = c.DemographicConsumption(context.Background(), &eieiorpc.DemographicConsumptionInput{
				Demograph: eieiorpc.Demograph_Hispanic,
				Year:      int32(year),
			})
			if err != nil {
				t.Error(err)
			}
		}
	})
}
