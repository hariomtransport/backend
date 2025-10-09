package models

type BiltyPDFData struct {
	Company    *InitialSetup // Company / Initial setup
	Bilty      *Bilty        // Bilty details
	Contacts   string        // formatted mobile numbers
	Date       string        // formatted date
	Total      float64       // total amount including charges
	TotalWords string
	CopyTitle  string
	GoodsCount int
}
