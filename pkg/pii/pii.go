package pii

type Category string

const (
	CategoryUnknown Category = "Unknown"

	CategoryPerson Category = "Person"

	CategoryNRP     Category = "NRP"
	CategoryAddress Category = "Address"

	CategoryEmail Category = "Email"
	CategoryPhone Category = "Phone"

	CategoryIBAN       Category = "IBAN"
	CategoryCreditCard Category = "CreditCard"
	CategoryCrypto     Category = "Crypto"

	CategoryURL       Category = "URL"
	CategoryIPAddress Category = "IPAddress"
)

type Analysis struct {
	Category Category
	Score    float64
}
