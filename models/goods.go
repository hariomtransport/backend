package models

type Goods struct {
	ID          int64    `json:"id" db:"id"`
	BiltyID     int64    `json:"bilty_id" db:"bilty_id"`
	Particulars string   `json:"particulars" db:"particulars"`
	NumOfPkts   int      `json:"num_of_pkts" db:"num_of_pkts"`
	WeightKG    *float64 `json:"weight_kg,omitempty" db:"weight_kg"`
	Rate        *float64 `json:"rate,omitempty" db:"rate"`
	Per         *string  `json:"per,omitempty" db:"per"`
	Amount      *float64 `json:"amount,omitempty" db:"amount"`
}
