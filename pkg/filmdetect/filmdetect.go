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

func GetRecipeFromJson(b []byte) (Recipe, error) {
	recipe := Recipe{}
	err := json.Unmarshal(b, &recipe)
	if err != nil {
		return recipe, err
	}

	return recipe, nil
}

func ParseWhiteBalanceOffset(input string) (int, int, error) {
	if input == "" {
		return 0, 0, nil
	}
	p := regexp.MustCompile(`Red ([\-+][0-9]+), Blue ([\-+][0-9]+)`)
	matches := p.FindStringSubmatch(input)

	redMatch := matches[1]
	blueMatch := matches[2]

	red, err := strconv.Atoi(redMatch)
	if err != nil {
		return 0, 0, err
	}
	blue, err := strconv.Atoi(blueMatch)
	if err != nil {
		return 0, 0, err
	}

	red = red / 20
	blue = blue / 20
	return red, blue, nil
}

func ParseHighlightShadow(input string) (int, error) {
	if input == "" || input == "Normal" {
		return 0, nil
	}
	p := regexp.MustCompile(`([\-+]?[0-9]+)`)
	matches := p.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0, fmt.Errorf("Parsing highlight/shadow value failed: Unexpected value: '%s'", input)
	}
	match := matches[1]
	value, err := strconv.Atoi(match)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ParseSharpness(input string) (int, error) {
	switch input {
	case "Softest":
		return -4, nil
	case "Very Soft":
		return -3, nil
	case "Soft":
		return -2, nil
	case "Medium Soft":
		return -1, nil
	case "Normal":
		return 0, nil
	case "Medium Hard":
		return 1, nil
	case "Hard":
		return 2, nil
	case "Very Hard":
		return 3, nil
	case "Hardest":
		return 4, nil
	}

	return 0, fmt.Errorf("wrong value for sharpness")
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
				red, blue, err := ParseWhiteBalanceOffset(stringValue)
				if err != nil {
					return recipe, err
				}

				recipe.WhiteBalanceRed = red
				recipe.WhiteBalanceBlue = blue
			}

			if k == "DevelopmentDynamicRange" {
				dyn := strconv.FormatFloat(floatValue, 'f', 0, 64)
				recipe.DynamicRange = dyn
			}

			if k == "HighlightTone" {
				high, err := ParseHighlightShadow(stringValue)
				if err != nil {
					return Recipe{}, err
				}

				recipe.Highlights = high
			}

			if k == "ShadowTone" {
				shadow, err := ParseHighlightShadow(stringValue)
				if err != nil {
					return Recipe{}, err
				}

				recipe.Shadows = shadow
			}

			if k == "Saturation" {
				if strings.Contains(stringValue, "Acros") {
					recipe.Color = 0
					recipe.FilmSimulation = stringValue
				} else {
					color, err := ParseHighlightShadow(stringValue)
					if err != nil {
						return Recipe{}, err
					}
					recipe.Color = color
				}
			}

			if k == "Sharpness" {

				sharpness, err := ParseSharpness(stringValue)
				if err != nil {
					return recipe, err
				}

				recipe.Sharpness = sharpness
			}

			if k == "NoiseReduction" {
				noise, err := ParseHighlightShadow(stringValue)
				if err != nil {
					return recipe, err
				}

				recipe.NoiseReduction = noise
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

func DetectFromRecipes(recipes []Recipe, recipe Recipe) ([]Difference, bool, error) {
	resultDifferences := []Difference{}

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

	recipe, err := GetRecipeFromFile(filename)
	if err != nil {
		return []Difference{}, false, err
	}

	return DetectFromRecipes(allRecipes, recipe)

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
