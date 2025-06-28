package tasks

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	backend "popmart/src/backend"
	helpers "popmart/src/middleware/helpers"
)

// ----------------------- UTILITY FUNCS ----------------------- \\
func ParseInt(val string, def int) int {
	if parsed, err := strconv.Atoi(val); err == nil {
		return parsed
	}
	return def
}

func Normalize(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "ra", "rand", "random":
		if rand.Intn(2) == 0 {
			return "Single box"
		}
		return "Whole set"
	case "single box":
		return "Single box"
	case "whole set":
		return "Whole set"
	default:
		return input
	}
}

// ----------------------- TASK HELPERS ----------------------- \\
func LoadJson[T any](path string, target *T) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func LoadCsv(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return csv.NewReader(file).ReadAll()
}

func FindProxy(groups []backend.ProxyGroup, name string) *backend.ProxyGroup {
	for _, g := range groups {
		if g.Name == name {
			return &g
		}
	}
	return nil
}

func FindAccount(groups []backend.AccountGroup, groupName, email string) string {
	for _, g := range groups {
		if g.Name == groupName {
			for _, acc := range g.Accounts {
				parts := strings.SplitN(acc, ":", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], email) {
					return acc
				}
			}
		}
	}
	return ""
}

func BuildProfile(records [][]string) (map[string][]helpers.Profile, error) {
	if len(records) < 2 {
		return nil, fmt.Errorf("profiles.csv is empty or missing headers")
	}

	headers := records[0]
	index := -1
	for i, h := range headers {
		if h == "Profile Group Name" {
			index = i
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("missing 'Profile Group Name' column in profiles.csv")
	}

	grouped := make(map[string][]helpers.Profile)
	for _, row := range records[1:] {
		if len(row) < 15 {
			continue
		}
		group := row[index]
		profile := helpers.Profile{
			ProfileName: row[1], Email: row[2], Name: row[3], Phone: row[4],
			Address1: row[5], Address2: row[6], City: row[7], PostCode: row[8],
			Country: row[9], State: row[10], CardNumber: row[11], ExpMonth: row[12],
			ExpYear: row[13], CVV: row[14],
		}
		grouped[group] = append(grouped[group], profile)
	}
	return grouped, nil
}
