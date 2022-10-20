// filmdetect
// Copyright (C) 2021 Honza Pokorny <honza@pokorny.ca>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package filmdetect

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/olekukonko/tablewriter"
)

// The number of fields in Recipe
const FullScore = 16

type Recipe struct {
	Name                 string `json:"name"`
	Author               string
	Url                  string
	FilmSimulation       string `json:"film_simulation"`
	GrainEffectSize      string `json:"grain_effect_size"`
	GrainEffectRoughness string `json:"grain_effect_roughness"`
	ColorChromeEffect    string `json:"color_chrome_effect"`
	ColorChromeFXBlue    string `json:"color_chrome_fx_blue"`
	WhiteBalanceMode     string `json:"white_balance_mode"`
	WhiteBalanceRed      int    `json:"white_balance_r"`
	WhiteBalanceBlue     int    `json:"white_balance_b"`
	DynamicRange         string `json:"dynamic_range"`
	Highlights           int    `json:"tone_curve_highlights"`
	Shadows              int    `json:"tone_curve_shadows"`
	Color                int
	Sharpness            int
	NoiseReduction       int `json:"noise_reduction"`
	Clarity              int
}

func (r Recipe) String() string {
	return fmt.Sprintf(`Name: %s
  FilmSimulation: %s
  GrainEffectSize: %s
  GrainEffectRoughness: %s
  ColorChromeEffect: %s
  ColorChromeFXBlue: %s
  WhiteBalanceMode: %s
  WhiteBalanceRed: %d
  WhiteBalanceBlue: %d
  DynamicRange: %s
  Highlights: %d
  Shadows: %d
  Color: %d
  Sharpness: %d
  NoiseReduction: %d
  Clarity: %d
`,
		r.Name,
		r.FilmSimulation,
		r.GrainEffectSize,
		r.GrainEffectRoughness,
		r.ColorChromeEffect,
		r.ColorChromeFXBlue,
		r.WhiteBalanceMode,
		r.WhiteBalanceRed,
		r.WhiteBalanceBlue,
		r.DynamicRange,
		r.Highlights,
		r.Shadows,
		r.Color,
		r.Sharpness,
		r.NoiseReduction,
		r.Clarity)
}

func GetFiles(path string) ([]string, error) {
	var files []string

	fs, err := ioutil.ReadDir(path)

	if err != nil {
		return files, err
	}

	for _, file := range fs {
		abs := filepath.Join(
			path,
			file.Name(),
		)
		files = append(files, abs)
	}

	sort.Strings(files)

	return files, nil
}

func ParseRecipeFile(filename string) (Recipe, error) {
	var recipe Recipe
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return recipe, err
	}
	err = json.Unmarshal(contents, &recipe)

	if err != nil {
		return recipe, err
	}

	return recipe, nil
}

func GetRecipes(simulationDir string) ([]Recipe, error) {
	var recipes []Recipe
	files, err := GetFiles(simulationDir)

	if err != nil {
		return recipes, err
	}

	for _, file := range files {
		recipe, err := ParseRecipeFile(file)

		if err != nil {
			return recipes, err
		}

		recipes = append(recipes, recipe)

	}

	return recipes, nil

}

func GetRecipeFromFile(filename string) (Recipe, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		fmt.Printf("Error when intializing: %v", err)
		return Recipe{}, err
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(filename)

	recipe := Recipe{
		DynamicRange: "Auto",
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			fmt.Printf("Error concerning %v: %v", fileInfo.File, fileInfo.Err)
			continue
		}

		for k, v := range fileInfo.Fields {
			if k == "Subject" {
				continue
			}
			stringValue := ""
			floatValue := 0.0

			switch value := v.(type) {
			case string:
				stringValue = value
			case float64:
				floatValue = value
			default:
				return Recipe{}, errors.New("Field value isn't string of float.")
			}

			if k == "FilmMode" {
				recipe.FilmSimulation = stringValue
			}

			if k == "GrainEffectRoughness" {
				recipe.GrainEffectRoughness = stringValue
			}

			if k == "ColorChromeEffect" {
				recipe.ColorChromeEffect = stringValue
			}

			if k == "ColorChromeFXBlue" {
				recipe.ColorChromeFXBlue = stringValue
			}

			if k == "WhiteBalance" {
				recipe.WhiteBalanceMode = stringValue
			}

			if k == "WhiteBalanceFineTune" {

				p := regexp.MustCompile(`Red ([\-+][0-9]+), Blue ([\-+][0-9]+)`)
				matches := p.FindStringSubmatch(stringValue)

				redMatch := matches[1]
				blueMatch := matches[2]

				red, _ := strconv.Atoi(redMatch)
				blue, _ := strconv.Atoi(blueMatch)

				red = red / 20
				blue = blue / 20

				recipe.WhiteBalanceRed = red
				recipe.WhiteBalanceBlue = blue
			}

			if k == "DevelopmentDynamicRange" {
				dyn := strconv.FormatFloat(floatValue, 'f', 0, 64)
				recipe.DynamicRange = dyn
			}

			if k == "HighlightTone" {
				p := regexp.MustCompile(`([\-+]?[0-9]+)`)
				matches := p.FindStringSubmatch(stringValue)
				if len(matches) < 2 {
					return Recipe{}, errors.New("Unexpected highlight value")
				}
				highlightMatch := matches[1]
				highlightValue, _ := strconv.Atoi(highlightMatch)
				recipe.Highlights = highlightValue
			}

			if k == "ShadowTone" {
				p := regexp.MustCompile(`([\-+]?[0-9]+)`)
				matches := p.FindStringSubmatch(stringValue)
				if len(matches) < 2 {
					return Recipe{}, errors.New("Unexpected shadow value")
				}
				shadowMatch := matches[1]
				shadowValue, _ := strconv.Atoi(shadowMatch)
				recipe.Shadows = shadowValue
			}

			if k == "Saturation" {
				p := regexp.MustCompile(`([\-+]?[0-9]+)`)
				matches := p.FindStringSubmatch(stringValue)
				if len(matches) < 2 {
					return Recipe{}, errors.New("Unexpected saturation value")
				}
				colorMatch := matches[1]
				colorValue, _ := strconv.Atoi(colorMatch)
				recipe.Color = colorValue
			}

			if k == "Sharpness" {
				switch stringValue {
				case "Softest":
					recipe.Sharpness = -4
				case "Very Soft":
					recipe.Sharpness = -3
				case "Soft":
					recipe.Sharpness = -2
				case "Medium Soft":
					recipe.Sharpness = -1
				case "Normal":
					recipe.Sharpness = 0
				case "Medium Hard":
					recipe.Sharpness = 1
				case "Hard":
					recipe.Sharpness = 2
				case "Very Hard":
					recipe.Sharpness = 3
				case "Hardest":
					recipe.Sharpness = 4
				}
			}

			if k == "NoiseReduction" {
				p := regexp.MustCompile(`([\-+]?[0-9]+)`)
				matches := p.FindStringSubmatch(stringValue)
				if len(matches) < 2 {
					fmt.Println(stringValue, matches)
					return Recipe{}, errors.New("Unexpected noise reduction value")
				}
				noiseMatch := matches[1]
				noiseValue, _ := strconv.Atoi(noiseMatch)
				recipe.NoiseReduction = noiseValue
			}

			if k == "Clarity" {
				recipe.Clarity = int(floatValue)
			}

			if k == "GrainEffectSize" {
				recipe.GrainEffectSize = stringValue
			}

		}
	}

	return recipe, nil

}

type Difference struct {
	Input     Recipe
	Candidate Recipe
	Lines     [][]string
}

func DifferenceFromRecipes(input, candidate Recipe) Difference {
	d := Difference{Input: input, Candidate: candidate}
	d.Lines = d.GetLines()
	return d
}

func (d Difference) IsFullScore() bool {
	return len(d.Lines) == 0
}

func (d Difference) Score() int {
	return FullScore - len(d.Lines)
}

func (d Difference) AsList() []string {
	return []string{"White balance", "1", "2"}
}
func (d Difference) GetLines() [][]string {
	vInput := reflect.ValueOf(d.Input)
	vCandidate := reflect.ValueOf(d.Candidate)

	typeOfvInput := vInput.Type()
	// typeOfvCandidate := vCandidate.Type()

	result := [][]string{}
	for i := 0; i < vInput.NumField(); i++ {
		fieldName := typeOfvInput.Field(i).Name

		if strings.Contains("Name Author Url", fieldName) {
			continue
		}

		vInputValue := vInput.Field(i).Interface()
		vCandidateValue := vCandidate.Field(i).Interface()

		if vInputValue != vCandidateValue {
			result = append(result, []string{
				fieldName,
				fmt.Sprintf("%v", vInputValue),
				fmt.Sprintf("%v", vCandidateValue),
			})
		}

	}

	return result

}

func (d Difference) String() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{d.Candidate.Name, "Input", "Candidate"})
	table.AppendBulk(d.Lines)
	table.Render()
	return tableString.String()
}

func DetectFromRecipes(recipes []Recipe, filename string) ([]Difference, bool, error) {
	resultDifferences := []Difference{}

	recipe, err := GetRecipeFromFile(filename)
	if err != nil {
		return resultDifferences, false, err
	}

	differences := []Difference{}

	for _, candidate := range recipes {
		differences = append(differences, DifferenceFromRecipes(recipe, candidate))
	}

	sort.Slice(differences, func(i, j int) bool {
		return differences[i].Score() > differences[j].Score()
	})

	topScore := 0

	for _, diff := range differences {
		if diff.IsFullScore() {
			return []Difference{diff}, true, nil
		}

		if topScore != 0 {
			if topScore > diff.Score() {
				break
			}
			resultDifferences = append(resultDifferences, diff)
			continue
		}

		if topScore == 0 {
			topScore = diff.Score()
			resultDifferences = append(resultDifferences, diff)
		}

	}

	return resultDifferences, false, nil
}

// Detect is the main library function. It returns a list of differences, and
// the bool in the return means "were we able to find a perfect match?"
func Detect(simulationDir string, filename string) ([]Difference, bool, error) {
	allRecipes, err := GetRecipes(simulationDir)
	if err != nil {
		return []Difference{}, false, err
	}

	return DetectFromRecipes(allRecipes, filename)

}

// CLI
func Run(simulationDir string, filename string) {
	diffs, havePerfectMatch, err := Detect(simulationDir, filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	if havePerfectMatch {
		fmt.Println(diffs[0].Candidate.Name)
		return
	}

	fmt.Println("We were not able to find a perfect match.  These recipes are the closest:")

	for _, diff := range diffs {
		fmt.Println(diff)
	}
}
