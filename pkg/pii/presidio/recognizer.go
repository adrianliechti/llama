package presidio

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/pii"
)

type Recognizer struct {
	url string

	client *http.Client
}

type Option func(*Recognizer)

func NewRecognizer(url string, options ...Option) (*Recognizer, error) {
	if url == "" {
		return nil, errors.New("url is required")
	}

	r := &Recognizer{
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(r)
	}

	return r, nil
}

func (r *Recognizer) Recognize(text string) ([]pii.Analysis, error) {
	body := &AnalyzeRequest{
		Text:     text,
		Language: "en",
	}

	u, _ := url.JoinPath(r.url, "/analyze")
	resp, err := r.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to recognize")
	}

	var results []AnalyzeResponse

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	var result []pii.Analysis

	for _, r := range results {
		score := r.Score
		category := convertCategory(r.EntityType)

		if category == pii.CategoryUnknown {
			continue
		}

		result = append(result, pii.Analysis{
			Category: category,
			Score:    score,
		})
	}

	return result, nil
}

func convertCategory(val EntityType) pii.Category {
	switch val {
	case EntityTypePerson:
		return pii.CategoryPerson

	case EntityTypeNRP:
		return pii.CategoryNRP

	case EntityTypeAddress:
		return pii.CategoryAddress

	case EntityTypeEmail:
		return pii.CategoryEmail

	case EntityTypePhone:
		return pii.CategoryPhone

	case EntityTypeIBAN:
		return pii.CategoryIBAN

	case EntityTypeCreditCard:
		return pii.CategoryCreditCard

	case EntityTypeCrypto:
		return pii.CategoryCrypto

	case EntityTypeURL:
		return pii.CategoryURL

	case EntityTypeIPAddress:
		return pii.CategoryIPAddress
	}

	return pii.CategoryUnknown
}

type AnalyzeRequest struct {
	Text     string `json:"text,omitempty"`
	Language string `json:"language,omitempty"`
}

type EntityType string

// https://github.com/microsoft/presidio/blob/main/docs/supported_entities.md
const (
	EntityTypePerson EntityType = "PERSON"

	EntityTypeNRP     EntityType = "NRP" // Nationality, Religious or Political Group
	EntityTypeAddress EntityType = "LOCATION"

	EntityTypeEmail EntityType = "EMAIL_ADDRESS"
	EntityTypePhone EntityType = "PHONE_NUMBER"

	EntityTypeIBAN       EntityType = "IBAN_CODE"
	EntityTypeCreditCard EntityType = "CREDIT_CARD"
	EntityTypeCrypto     EntityType = "CRYPTO"

	// EntityTypeDate           EntityType = "DATE_TIME"
	//
	// EntityTypeMedicalLicense EntityType = "MEDICAL_LICENSE"

	EntityTypeURL       EntityType = "URL"
	EntityTypeIPAddress EntityType = "IP_ADDRESS"
)

type AnalyzeResponse struct {
	EntityType EntityType `json:"entity_type,omitempty"`

	Score float64 `json:"score,omitempty"`

	Start int `json:"start,omitempty"`
	End   int `json:"end,omitempty"`
}
