package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gtuk/discordwebhook"
	"github.com/leekchan/accounting"
)

type Config struct {
	ErrorWebhook  string
	NotifyWebhook string
}

var seen map[int]Item = map[int]Item{}

// TODO: set to false after dev
var initialized bool = true

type SalaryEntries struct {
	Items []Item `json:"items"`
}

type Item struct {
	ID                 int     `json:"id"`
	RoleTitle          string  `json:"role_title"`
	RoleFocus          string  `json:"role_focus"`
	Grade              string  `json:"grade"`
	AnnualCompensation int     `json:"annual_compensation"`
	AnnualSalary       int     `json:"annual_salary"`
	AnnualBonus        int     `json:"annual_bonus"`
	AnnualStock        int     `json:"annual_stock"`
	SigningBonusTotal  int     `json:"signing_bonus_total"`
	YearsOfExperience  int     `json:"years_of_experience"`
	YearsAtCompany     int     `json:"years_at_company"`
	Company            Company `json:"company"`
}

func (i *Item) URL() string {
	return fmt.Sprintf("https://opensalary.jp/en/single-salary/%d", i.ID)
}

type Company struct {
	Name string `json:"name_en"`
	Slug string `json:"slug"`
}

func (c *Company) URL() string {
	return fmt.Sprintf("https://opensalary.jp/en/companies/%s", c.Slug)
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

func updateSalaryEntries(newSalaries chan Item) {
	entries, err := fetchSalaryEntries()
	if err != nil {
		panic(err)
	}

	for _, item := range entries.Items {
		if _, ok := seen[item.ID]; !ok {
			seen[item.ID] = item

			// Temporarily add logic to limit to 1 entries to avoid spamming during dev
			if len(seen) == 2 {
				close(newSalaries)
			} else if len(seen) <= 1 {
				newSalaries <- item
			}
		}
	}
}

func sendNotification(url, username, content string, embeds []discordwebhook.Embed) {
	message := discordwebhook.Message{
		Username: &username,
		Content:  &content,
		Embeds:   &embeds,
	}

	err := discordwebhook.SendMessage(url, message)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var config = &Config{
		NotifyWebhook: "",
		ErrorWebhook:  "",
	}

	ac := accounting.Accounting{Symbol: "¥", Precision: 0}
	newSalaries := make(chan Item)
	go func() {
		for {
			updateSalaryEntries(newSalaries)
			initialized = true
			time.Sleep(10 * time.Second)
		}
	}()

	for item := range newSalaries {
		if !initialized {
			return
		}

		content := "New salary submitted"
		companyURL := item.Company.URL()
		itemURL := item.URL()
		title := fmt.Sprintf("%s @ %s", ac.FormatMoney(item.AnnualCompensation), item.Company.Name)
		labelRole := "Role"
		labelFocus := "Focus"
		labelGrade := "Grade"
		labelSalary := "Salary"
		labelBonus := "Bonus"
		labelStock := "Stock"
		labelSigningBonus := "Signing bonus"
		labelYOE := "Years of experience"
		labelYAC := "Years at company"
		salary := ac.FormatMoney(item.AnnualSalary)
		bonus := ac.FormatMoney(item.AnnualBonus)
		stock := ac.FormatMoney(item.AnnualStock)
		signingBonus := ac.FormatMoney(item.SigningBonusTotal)
		yoe := fmt.Sprintf("%d", item.YearsOfExperience)
		yac := fmt.Sprintf("%d", item.YearsAtCompany)
		t := true
		color := "15258703"
		embeds := []discordwebhook.Embed{
			{
				Author: &discordwebhook.Author{Name: &item.Company.Name, Url: &companyURL},
				Title:  &title,
				Url:    &itemURL,
				Color:  &color,
				Fields: &[]discordwebhook.Field{
					{Name: &labelRole, Value: &item.RoleTitle, Inline: &t},
					{Name: &labelFocus, Value: &item.RoleFocus, Inline: &t},
					{Name: &labelGrade, Value: &item.Grade, Inline: &t},
					{Name: &labelSalary, Value: &salary, Inline: &t},
					{Name: &labelBonus, Value: &bonus, Inline: &t},
					{Name: &labelYOE, Value: &yoe, Inline: &t},
					{Name: &labelStock, Value: &stock, Inline: &t},
					{Name: &labelSigningBonus, Value: &signingBonus, Inline: &t},
					{Name: &labelYAC, Value: &yac, Inline: &t},
				},
			},
		}
		sendNotification(config.NotifyWebhook, "OpenSalary", content, embeds)
	}
}
