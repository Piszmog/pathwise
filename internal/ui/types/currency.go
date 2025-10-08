package types

// CurrencyInfo holds the code and symbol for a currency.
type CurrencyInfo struct {
	Code   string
	Symbol string
}

// AvailableCurrencies is a list of common currencies for the dropdown.
var AvailableCurrencies = []CurrencyInfo{
	{Code: "USD", Symbol: "$"},
	{Code: "EUR", Symbol: "€"},
	{Code: "GBP", Symbol: "£"},
	{Code: "CAD", Symbol: "CA$"},
	{Code: "AUD", Symbol: "A$"},
	{Code: "JPY", Symbol: "¥"},
	{Code: "CHF", Symbol: "CHF"},
}
