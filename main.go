package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var seen map[int]Item = map[int]Item{}

type SalaryEntries struct {
	Items []Item `json:"items"`
}

type Item struct {
	ID                 int     `json:"id"`
	RoleTitle          string  `json:"role_title"`
	RoleFocus          string  `json:"role_focus"`
	AnnualCompensation int     `json:"annual_compensation"`
	AnnualSalary       int     `json:"annual_salary"`
	AnnualBonus        int     `json:"annual_bonus"`
	AnnualStock        int     `json:"annual_stock"`
	SigningBonusTotal  int     `json:"signing_bonus_total"`
	YearsOfExperience  int     `json:"years_of_experience"`
	YearsAtCompany     int     `json:"years_at_company"`
	Company            Company `json:"company"`
}

type Company struct {
	Name string `json:"name_en"`
}

func fetchSalaryEntries() (*SalaryEntries, error) {
	resp, err := http.Get("https://api.opensalary.jp/api/salary-entries?page=1&job_role=software-engineer&locale=en")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to fetch salary entries: status code %v", resp.StatusCode)
	}

	var entries *SalaryEntries
	body, _ := io.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func main() {
	entries, err := fetchSalaryEntries()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", entries)
}
