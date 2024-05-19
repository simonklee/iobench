package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type FioResult struct {
	Jobs []struct {
		Read struct {
			Bw   float64 `json:"bw"`
			Iops float64 `json:"iops"`
		} `json:"read"`
		Write struct {
			Bw   float64 `json:"bw"`
			Iops float64 `json:"iops"`
		} `json:"write"`
	} `json:"jobs"`
}

func loadFioJson(filename string) (FioResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return FioResult{}, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return FioResult{}, err
	}

	var result FioResult
	if err := json.Unmarshal(bytes, &result); err != nil {
		return FioResult{}, err
	}

	return result, nil
}

func main() {
	inputDir := flag.String("input", "", "Input directory containing JSON files from fio")
	outputFile := flag.String("output", "", "Output PNG file for the plot")
	title := flag.String("title", "SSD Benchmark Results", "Title of the plot")
	xLabel := flag.String("xlabel", "Test Type", "Label for the x-axis")
	yLabel := flag.String("ylabel", "", "Label for the y-axis")
	valueType := flag.String("value-type", "bw", "Type of value to plot: 'bw' for bandwidth, 'iops' for IOPS")
	flag.Parse()

	if *inputDir == "" || *outputFile == "" || *yLabel == "" {
		log.Fatal("input, output, and ylabel must be specified")
	}

	var readValues, writeValues plotter.Values
	var testTypes []string

	err := filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".json" {
			result, err := loadFioJson(path)
			if err != nil {
				return fmt.Errorf("error loading fio JSON file %s: %v", path, err)
			}

			testName := filepath.Base(path[:len(path)-len(filepath.Ext(path))])
			testTypes = append(testTypes, testName)
			if *valueType == "iops" {
				readValues = append(readValues, result.Jobs[0].Read.Iops)
				writeValues = append(writeValues, result.Jobs[0].Write.Iops)
			} else {
				readValues = append(readValues, result.Jobs[0].Read.Bw/1024)    // Convert KB/s to MB/s
				writeValues = append(writeValues, result.Jobs[0].Write.Bw/1024) // Convert KB/s to MB/s
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking through input directory: %v", err)
	}

	if len(readValues) == 0 && len(writeValues) == 0 {
		log.Fatal("No data points found for plotting")
	}

	p := plot.New()
	p.Title.Text = *title
	p.X.Label.Text = *xLabel
	p.Y.Label.Text = *yLabel

	p.X.Tick.Label.Rotation = 1.0
	p.X.Tick.Label.Font.Size = vg.Points(10)
	p.Y.Tick.Label.Font.Size = vg.Points(10)

	p.Add(plotter.NewGrid())

	readBars, err := plotter.NewBarChart(readValues, vg.Points(15))
	if err != nil {
		log.Fatal(err)
	}
	readBars.LineStyle.Width = vg.Length(0)
	readBars.Color = plotutil.Color(0)
	readBars.Offset = -vg.Points(7.5)

	writeBars, err := plotter.NewBarChart(writeValues, vg.Points(15))
	if err != nil {
		log.Fatal(err)
	}
	writeBars.LineStyle.Width = vg.Length(0)
	writeBars.Color = plotutil.Color(1)
	writeBars.Offset = vg.Points(7.5)

	p.Add(readBars, writeBars)
	p.Legend.Add("Read", readBars)
	p.Legend.Add("Write", writeBars)
	p.Legend.Top = true

	p.NominalX(testTypes...)

	if err := p.Save(12*vg.Inch, 8*vg.Inch, *outputFile); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Plot saved as %s\n", *outputFile)
}
