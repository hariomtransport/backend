package utils

import (
	"fmt"
	"math"
	"strings"
)

var ones = []string{
	"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine",
	"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen",
	"Sixteen", "Seventeen", "Eighteen", "Nineteen",
}

var tens = []string{
	"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety",
}

func NumberToWords(num int) string {
	switch {
	case num == 0:
		return ""
	case num < 20:
		return ones[num]
	case num < 100:
		return strings.TrimSpace(tens[num/10] + " " + ones[num%10])
	case num < 1000:
		remainder := num % 100
		if remainder == 0 {
			return ones[num/100] + " Hundred"
		}
		return ones[num/100] + " Hundred " + NumberToWords(remainder)
	case num < 100000:
		remainder := num % 1000
		if remainder == 0 {
			return NumberToWords(num/1000) + " Thousand"
		}
		return NumberToWords(num/1000) + " Thousand " + NumberToWords(remainder)
	case num < 10000000:
		remainder := num % 100000
		if remainder == 0 {
			return NumberToWords(num/100000) + " Lakh"
		}
		return NumberToWords(num/100000) + " Lakh " + NumberToWords(remainder)
	default:
		remainder := num % 10000000
		if remainder == 0 {
			return NumberToWords(num/10000000) + " Crore"
		}
		return NumberToWords(num/10000000) + " Crore " + NumberToWords(remainder)
	}
}

// ✅ Main helper for Rupees + Paise
func NumberToCurrencyWords(amount float64) string {
	rupees := int(math.Floor(amount))
	paise := int(math.Round((amount - float64(rupees)) * 100))

	var parts []string

	if rupees > 0 {
		parts = append(parts, fmt.Sprintf("%s Rupees", strings.TrimSpace(NumberToWords(rupees))))
	}
	if paise > 0 {
		parts = append(parts, fmt.Sprintf("%s Paise", strings.TrimSpace(NumberToWords(paise))))
	}

	if len(parts) == 0 {
		return "Zero Rupees Only"
	}

	return strings.Join(parts, " and ") + " Only"
}
