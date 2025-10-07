package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"hariomtransport/models"
)

type PostgresBiltyRepo struct {
	DB *sql.DB
}

func NewPostgresBiltyRepo(db *sql.DB) *PostgresBiltyRepo {
	return &PostgresBiltyRepo{DB: db}
}

// ------------------------ Helper Functions ------------------------

// Upsert AppUser
func (r *PostgresBiltyRepo) upsertUser(tx *sql.Tx, u *models.AppUser) error {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	log.Println("Upserting user:", u.ID)
	_, err := tx.Exec(`
		INSERT INTO app_user(id,name,email,role,created_at)
		VALUES($1,$2,$3,$4,$5)
		ON CONFLICT(id) DO NOTHING
	`, u.ID, u.Name, u.Email, u.Role, u.CreatedAt)
	return err
}

// Upsert Company, returns ID
func (r *PostgresBiltyRepo) upsertCompany(tx *sql.Tx, c *models.Company) (int64, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	if c.ID == 0 {
		var id int64
		err := tx.QueryRow(`
			INSERT INTO company(name,gstin,created_at)
			VALUES($1,$2,$3)
			RETURNING id
		`, c.Name, c.GSTIN, c.CreatedAt).Scan(&id)
		if err != nil {
			return 0, err
		}
		c.ID = id
		return id, nil
	}
	_, err := tx.Exec(`
		INSERT INTO company(id,name,gstin,created_at)
		VALUES($1,$2,$3,$4)
		ON CONFLICT(id) DO NOTHING
	`, c.ID, c.Name, c.GSTIN, c.CreatedAt)
	return c.ID, err
}

// Insert company address
func (r *PostgresBiltyRepo) insertCompanyAddress(tx *sql.Tx, addr *models.CompanyAddress) (int64, error) {
	if addr.CreatedAt.IsZero() {
		addr.CreatedAt = time.Now().UTC()
	}
	var id int64
	err := tx.QueryRow(`
		INSERT INTO company_address(company_id,address_line,city,state,pincode,is_default,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`, addr.CompanyID, addr.AddressLine, addr.City, addr.State, addr.Pincode, addr.IsDefault, addr.CreatedAt).Scan(&id)
	return id, err
}

// Insert bilty address snapshot
func (r *PostgresBiltyRepo) insertBiltyAddress(tx *sql.Tx, addr *models.CompanyAddress) (int64, error) {
	var id int64
	err := tx.QueryRow(`
		INSERT INTO bilty_address(company_id,address_line,city,state,pincode,created_at)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING id
	`, addr.CompanyID, addr.AddressLine, addr.City, addr.State, addr.Pincode, time.Now().UTC()).Scan(&id)
	return id, err
}

// Insert goods
func (r *PostgresBiltyRepo) insertGoods(tx *sql.Tx, biltyID int64, goods []models.Goods) error {
	for i := range goods {
		g := &goods[i]
		_, err := tx.Exec(`
			INSERT INTO goods(bilty_id,particulars,num_of_pkts,weight_kg,rate,per,amount)
			VALUES($1,$2,$3,$4,$5,$6,$7)
		`, biltyID, g.Particulars, g.NumOfPkts, g.WeightKG, g.Rate, g.Per, g.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

// Insert main bilty record
func (r *PostgresBiltyRepo) insertBiltyMain(tx *sql.Tx, bilty *models.Bilty) error {
	if bilty.CreatedAt.IsZero() {
		bilty.CreatedAt = time.Now().UTC()
	}
	return tx.QueryRow(`
		INSERT INTO bilty(
			consignor_company_id,consignee_company_id,
			consignor_address_id,consignee_address_id,
			from_location,to_location,date,to_pay,gstin,inv_no,pvt_marks,permit_no,
			value_rupees,remarks,hamali,dd_charges,other_charges,fov,statistical,
			created_by,created_at,status
		)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING id, bilty_no
	`, bilty.ConsignorCompanyID, bilty.ConsigneeCompanyID, bilty.ConsignorAddressID, bilty.ConsigneeAddressID,
		bilty.FromLocation, bilty.ToLocation, bilty.Date, bilty.ToPay, bilty.GSTIN, bilty.InvNo,
		bilty.PVTMarks, bilty.PermitNo, bilty.ValueRupees, bilty.Remarks, bilty.Hamali,
		bilty.DDCharges, bilty.OtherCharges, bilty.FOV, bilty.Statistical, bilty.CreatedBy,
		bilty.CreatedAt, bilty.Status).Scan(&bilty.ID, &bilty.BiltyNo)
}

// ------------------------ Main Function ------------------------

// CreateBilty creates bilty with companies, addresses, and goods
func (r *PostgresBiltyRepo) CreateBiltyWithParties(bilty *models.Bilty) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure CreatedBy
	if bilty.CreatedBy == 0 && bilty.CreatedByUser != nil {
		bilty.CreatedBy = bilty.CreatedByUser.ID
	}
	if bilty.CreatedBy == 0 {
		return errors.New("created_by cannot be empty")
	}

	// Upsert CreatedByUser
	if bilty.CreatedByUser != nil {
		if err := r.upsertUser(tx, bilty.CreatedByUser); err != nil {
			return err
		}
	}

	// Upsert consignor and consignee companies
	if bilty.ConsignorCompanyID == nil && bilty.ConsignorCompany != nil {
		id, err := r.upsertCompany(tx, bilty.ConsignorCompany)
		if err != nil {
			return err
		}
		bilty.ConsignorCompanyID = &id
	}
	if bilty.ConsigneeCompanyID == nil && bilty.ConsigneeCompany != nil {
		id, err := r.upsertCompany(tx, bilty.ConsigneeCompany)
		if err != nil {
			return err
		}
		bilty.ConsigneeCompanyID = &id
	}

	// Insert consignor address
	if bilty.ConsignorAddressID == nil && bilty.ConsignorAddressSnap != nil {
		companyAddr := &models.CompanyAddress{
			CompanyID:   *bilty.ConsignorCompanyID,
			AddressLine: bilty.ConsignorAddressSnap.AddressLine,
			City:        bilty.ConsignorAddressSnap.City,
			State:       bilty.ConsignorAddressSnap.State,
			Pincode:     bilty.ConsignorAddressSnap.Pincode,
			IsDefault:   false,
		}

		_, _ = r.insertCompanyAddress(tx, companyAddr)
		id, err := r.insertBiltyAddress(tx, companyAddr)
		if err != nil {
			return err
		}
		bilty.ConsignorAddressID = &id
	}

	// Insert consignee address
	if bilty.ConsigneeAddressID == nil && bilty.ConsigneeAddressSnap != nil {
		companyAddr := &models.CompanyAddress{
			CompanyID:   *bilty.ConsigneeCompanyID,
			AddressLine: bilty.ConsigneeAddressSnap.AddressLine,
			City:        bilty.ConsigneeAddressSnap.City,
			State:       bilty.ConsigneeAddressSnap.State,
			Pincode:     bilty.ConsigneeAddressSnap.Pincode,
			IsDefault:   false,
		}
		_, _ = r.insertCompanyAddress(tx, companyAddr)
		id, err := r.insertBiltyAddress(tx, companyAddr)
		if err != nil {
			return err
		}
		bilty.ConsigneeAddressID = &id
	}

	// Insert bilty
	if err := r.insertBiltyMain(tx, bilty); err != nil {
		return err
	}

	// Insert goods
	if err := r.insertGoods(tx, bilty.ID, bilty.Goods); err != nil {
		return err
	}

	return tx.Commit()
}

// ------------------------ Combined Fetch Function ------------------------

// GetBilty fetches one or multiple bilties with nested companies, addresses, goods, and created_by
func (r *PostgresBiltyRepo) GetBilty(filters map[string]interface{}, single bool) ([]*models.Bilty, error) {
	query := `
		SELECT id,bilty_no,consignor_company_id,consignee_company_id,
		       consignor_address_id,consignee_address_id,
		       from_location,to_location,date,to_pay,gstin,inv_no,pvt_marks,permit_no,
		       value_rupees,remarks,hamali,dd_charges,other_charges,fov,statistical,
		       created_by,created_at,status
		FROM bilty
	`
	args := []interface{}{}
	whereClauses := []string{}
	i := 1

	// Build WHERE dynamically
	for k, v := range filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if !single {
		query += " ORDER BY created_at DESC"
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Bilty

	// Helper to load company
	loadCompany := func(companyID *int64) (*models.Company, error) {
		if companyID == nil {
			return nil, nil
		}
		var c models.Company
		err := r.DB.QueryRow(`SELECT id,name,gstin,created_at FROM company WHERE id=$1`, *companyID).
			Scan(&c.ID, &c.Name, &c.GSTIN, &c.CreatedAt)
		if err != nil {
			return nil, nil
		}
		return &c, nil
	}

	// Helper to load bilty address
	loadAddress := func(addressID *int64) (*models.BiltyAddress, error) {
		if addressID == nil {
			return nil, nil
		}
		var a models.BiltyAddress
		err := r.DB.QueryRow(`SELECT id,company_id,address_line,city,state,pincode,created_at FROM bilty_address WHERE id=$1`, *addressID).
			Scan(&a.ID, &a.CompanyID, &a.AddressLine, &a.City, &a.State, &a.Pincode, &a.CreatedAt)
		if err != nil {
			return nil, nil
		}
		return &a, nil
	}

	for rows.Next() {
		var b models.Bilty
		if err := rows.Scan(
			&b.ID, &b.BiltyNo, &b.ConsignorCompanyID, &b.ConsigneeCompanyID,
			&b.ConsignorAddressID, &b.ConsigneeAddressID,
			&b.FromLocation, &b.ToLocation, &b.Date, &b.ToPay, &b.GSTIN, &b.InvNo,
			&b.PVTMarks, &b.PermitNo, &b.ValueRupees, &b.Remarks,
			&b.Hamali, &b.DDCharges, &b.OtherCharges, &b.FOV, &b.Statistical,
			&b.CreatedBy, &b.CreatedAt, &b.Status,
		); err != nil {
			return nil, err
		}

		// Load nested entities
		b.ConsignorCompany, _ = loadCompany(b.ConsignorCompanyID)
		b.ConsigneeCompany, _ = loadCompany(b.ConsigneeCompanyID)
		b.ConsignorAddressSnap, _ = loadAddress(b.ConsignorAddressID)
		b.ConsigneeAddressSnap, _ = loadAddress(b.ConsigneeAddressID)

		// Load goods
		gRows, _ := r.DB.Query(`SELECT id,bilty_id,particulars,num_of_pkts,weight_kg,rate,per,amount FROM goods WHERE bilty_id=$1`, b.ID)
		for gRows.Next() {
			var g models.Goods
			if err := gRows.Scan(&g.ID, &g.BiltyID, &g.Particulars, &g.NumOfPkts, &g.WeightKG, &g.Rate, &g.Per, &g.Amount); err == nil {
				b.Goods = append(b.Goods, g)
			}
		}
		gRows.Close()

		// Load created_by user
		if b.CreatedBy != 0 {
			var u models.AppUser
			if err := r.DB.QueryRow(`SELECT id,name,email,role,created_at FROM app_user WHERE id=$1`, b.CreatedBy).
				Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err == nil {
				b.CreatedByUser = &u
			}
		}

		result = append(result, &b)
	}

	if single {
		if len(result) > 0 {
			return []*models.Bilty{result[0]}, nil
		}
		return nil, nil
	}

	return result, nil
}
