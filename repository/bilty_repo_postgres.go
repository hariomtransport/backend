package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hariomtransport/backend/models"
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

	if c.GSTIN != nil {
		var existingID int64
		err := tx.QueryRow(`SELECT id FROM company WHERE gstin = $1 LIMIT 1`, c.GSTIN).Scan(&existingID)
		if err == nil {
			c.ID = existingID
			return existingID, nil
		}
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

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

// Find or insert company address if not exists
func (r *PostgresBiltyRepo) findOrInsertCompanyAddress(tx *sql.Tx, addr *models.CompanyAddress) (int64, error) {
	if addr.CreatedAt.IsZero() {
		addr.CreatedAt = time.Now().UTC()
	}

	var existingID int64
	err := tx.QueryRow(`
		SELECT id FROM company_address
		WHERE company_id=$1 AND address_line=$2 AND city=$3 AND state=$4 AND pincode=$5
		LIMIT 1
	`, addr.CompanyID, addr.AddressLine, addr.City, addr.State, addr.Pincode).Scan(&existingID)

	if err == nil {
		return existingID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	var newID int64
	err = tx.QueryRow(`
		INSERT INTO company_address(company_id,address_line,city,state,pincode,is_default,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`, addr.CompanyID, addr.AddressLine, addr.City, addr.State, addr.Pincode, addr.IsDefault, addr.CreatedAt).Scan(&newID)

	return newID, err
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

// Insert new bilty
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
		RETURNING id,bilty_no
	`,
		bilty.ConsignorCompanyID, bilty.ConsigneeCompanyID, bilty.ConsignorAddressID, bilty.ConsigneeAddressID,
		bilty.FromLocation, bilty.ToLocation, bilty.Date, bilty.ToPay, bilty.GSTIN, bilty.InvNo,
		bilty.PVTMarks, bilty.PermitNo, bilty.ValueRupees, bilty.Remarks, bilty.Hamali,
		bilty.DDCharges, bilty.OtherCharges, bilty.FOV, bilty.Statistical, bilty.CreatedBy,
		bilty.CreatedAt, bilty.Status,
	).Scan(&bilty.ID, &bilty.BiltyNo)
}

// ------------------------ Handle Bilty Address ------------------------

func (r *PostgresBiltyRepo) handleBiltyAddress(
	tx *sql.Tx,
	companyID *int64,
	addrSnap *models.BiltyAddress,
	oldAddrID *int64,
) (*int64, error) {
	if addrSnap == nil {
		return nil, nil
	}

	if oldAddrID != nil {
		var existing models.BiltyAddress
		err := tx.QueryRow(`
		SELECT address_line, city, state, pincode
		FROM bilty_address
		WHERE id=$1
		`, *oldAddrID).Scan(&existing.AddressLine, &existing.City, &existing.State, &existing.Pincode)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, err
			}
		} else {
			if existing.AddressLine == addrSnap.AddressLine &&
				existing.City == addrSnap.City &&
				existing.State == addrSnap.State &&
				existing.Pincode == addrSnap.Pincode {
				return oldAddrID, nil // no change
			}
		}
	}

	var newID int64
	err := tx.QueryRow(`
		INSERT INTO bilty_address(company_id,address_line,city,state,pincode,created_at)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING id
	`, *companyID, addrSnap.AddressLine, addrSnap.City, addrSnap.State, addrSnap.Pincode, time.Now().UTC()).Scan(&newID)
	if err != nil {
		return nil, err
	}

	if oldAddrID != nil {
		fmt.Printf("[INFO] Updating bilty records from old address ID %d to new ID %d\n", *oldAddrID, newID)

		// Update consignor_address_id and consignee_address_id in a single query
		if _, err := tx.Exec(`
        UPDATE bilty
        SET consignor_address_id = CASE WHEN consignor_address_id = $1 THEN $2 ELSE consignor_address_id END,
            consignee_address_id = CASE WHEN consignee_address_id = $1 THEN $2 ELSE consignee_address_id END
        WHERE consignor_address_id = $1 OR consignee_address_id = $1
    	`, *oldAddrID, newID); err != nil {
			return nil, err
		}

		// Delete old address if unused
		var count int
		err = tx.QueryRow(` SELECT COUNT(*) FROM bilty_address WHERE id=$1 `, *oldAddrID).Scan(&count)
		if err != nil {
			return nil, err
		}

		if count != 0 {
			fmt.Printf("[INFO] Deleting unused bilty_address ID %d\n", *oldAddrID)
			if _, err := tx.Exec(`DELETE FROM bilty_address WHERE id=$1`, *oldAddrID); err != nil {
				return nil, err
			}
		}
	}

	return &newID, nil
}

// ------------------------ Create / Update Bilty ------------------------

func (r *PostgresBiltyRepo) CreateBiltyWithParties(bilty *models.Bilty) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if bilty.CreatedBy == 0 && bilty.CreatedByUser != nil {
		bilty.CreatedBy = bilty.CreatedByUser.ID
	}
	if bilty.CreatedBy == 0 {
		return errors.New("created_by cannot be empty")
	}

	if bilty.CreatedByUser != nil {
		if err := r.upsertUser(tx, bilty.CreatedByUser); err != nil {
			return err
		}
	}

	// Upsert companies
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

	// Upsert company addresses (not bilty addresses)
	if bilty.ConsignorAddressSnap != nil {
		_, err := r.findOrInsertCompanyAddress(tx, &models.CompanyAddress{
			CompanyID:   *bilty.ConsignorCompanyID,
			AddressLine: bilty.ConsignorAddressSnap.AddressLine,
			City:        bilty.ConsignorAddressSnap.City,
			State:       bilty.ConsignorAddressSnap.State,
			Pincode:     bilty.ConsignorAddressSnap.Pincode,
		})
		if err != nil {
			return err
		}
	}

	if bilty.ConsigneeAddressSnap != nil {
		_, err := r.findOrInsertCompanyAddress(tx, &models.CompanyAddress{
			CompanyID:   *bilty.ConsigneeCompanyID,
			AddressLine: bilty.ConsigneeAddressSnap.AddressLine,
			City:        bilty.ConsigneeAddressSnap.City,
			State:       bilty.ConsigneeAddressSnap.State,
			Pincode:     bilty.ConsigneeAddressSnap.Pincode,
		})
		if err != nil {
			return err
		}
	}

	// Insert or update main bilty
	if bilty.ID == 0 {
		if bilty.ConsignorAddressSnap != nil {
			bilty.ConsignorAddressID, err = r.handleBiltyAddress(tx, bilty.ConsignorCompanyID, bilty.ConsignorAddressSnap, bilty.ConsignorAddressID)
			if err != nil {
				return err
			}
		}
		if bilty.ConsigneeAddressSnap != nil {
			bilty.ConsigneeAddressID, err = r.handleBiltyAddress(tx, bilty.ConsigneeCompanyID, bilty.ConsigneeAddressSnap, bilty.ConsigneeAddressID)
			if err != nil {
				return err
			}
		}
		if err := r.insertBiltyMain(tx, bilty); err != nil {
			return err
		}
	} else {
		// -------------------- Handle Company Updates --------------------
		var existingConsignor, existingConsignee models.Company

		// Fetch existing company details for comparison
		err := tx.QueryRow(`SELECT name, gstin FROM company WHERE id=$1`, bilty.ConsignorCompanyID).
			Scan(&existingConsignor.Name, &existingConsignor.GSTIN)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		err = tx.QueryRow(`SELECT name, gstin FROM company WHERE id=$1`, bilty.ConsigneeCompanyID).
			Scan(&existingConsignee.Name, &existingConsignee.GSTIN)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// If consignor company changed, insert a new one
		if bilty.ConsignorCompany != nil {
			nameChanged := existingConsignor.Name != bilty.ConsignorCompany.Name
			gstinChanged := (existingConsignor.GSTIN == nil && bilty.ConsignorCompany.GSTIN != nil) ||
				(existingConsignor.GSTIN != nil && bilty.ConsignorCompany.GSTIN == nil) ||
				(existingConsignor.GSTIN != nil && bilty.ConsignorCompany.GSTIN != nil &&
					*existingConsignor.GSTIN != *bilty.ConsignorCompany.GSTIN)

			if nameChanged || gstinChanged {
				newConsignorID, err := r.upsertCompany(tx, bilty.ConsignorCompany)
				if err != nil {
					return err
				}
				bilty.ConsignorCompanyID = &newConsignorID
			}
		}

		// If consignee company changed, insert a new one
		if bilty.ConsigneeCompany != nil {
			nameChanged := existingConsignee.Name != bilty.ConsigneeCompany.Name
			gstinChanged := (existingConsignee.GSTIN == nil && bilty.ConsigneeCompany.GSTIN != nil) ||
				(existingConsignee.GSTIN != nil && bilty.ConsigneeCompany.GSTIN == nil) ||
				(existingConsignee.GSTIN != nil && bilty.ConsigneeCompany.GSTIN != nil &&
					*existingConsignee.GSTIN != *bilty.ConsigneeCompany.GSTIN)

			if nameChanged || gstinChanged {
				newConsigneeID, err := r.upsertCompany(tx, bilty.ConsigneeCompany)
				if err != nil {
					return err
				}
				bilty.ConsigneeCompanyID = &newConsigneeID
			}
		}

		// -------------------- Address Change Detection --------------------
		var hasConsignorAddressChanged bool
		var consignorErr error

		if bilty.ConsignorAddressID != nil {
			hasConsignorAddressChanged, consignorErr = r.hasAddressChanged(tx, *bilty.ConsignorAddressID, bilty.ConsignorAddressSnap)
		} else {
			hasConsignorAddressChanged = true
		}
		if consignorErr != nil {
			return consignorErr
		}
		if hasConsignorAddressChanged && bilty.ConsignorAddressSnap != nil {
			bilty.ConsignorAddressID, err = r.handleBiltyAddress(tx, bilty.ConsignorCompanyID, bilty.ConsignorAddressSnap, bilty.ConsignorAddressID)
			if err != nil {
				return err
			}
		}

		var hasConsigneeAddressChanged bool
		var consigneeErr error

		if bilty.ConsigneeAddressID != nil {
			hasConsigneeAddressChanged, consigneeErr = r.hasAddressChanged(tx, *bilty.ConsigneeAddressID, bilty.ConsigneeAddressSnap)
		} else {
			hasConsigneeAddressChanged = true
		}
		if consigneeErr != nil {
			return consigneeErr
		}
		if hasConsigneeAddressChanged && bilty.ConsigneeAddressSnap != nil {
			bilty.ConsigneeAddressID, err = r.handleBiltyAddress(tx, bilty.ConsigneeCompanyID, bilty.ConsigneeAddressSnap, bilty.ConsigneeAddressID)
			if err != nil {
				return err
			}
		}

		// -------------------- Update Main Bilty --------------------
		_, err = tx.Exec(`
		UPDATE bilty SET
			consignor_company_id=$1,
			consignee_company_id=$2,
			from_location=$3,
			to_location=$4,
			date=$5,
			to_pay=$6,
			gstin=$7,
			inv_no=$8,
			pvt_marks=$9,
			permit_no=$10,
			value_rupees=$11,
			remarks=$12,
			hamali=$13,
			dd_charges=$14,
			other_charges=$15,
			fov=$16,
			statistical=$17,
			status=$18,
			updated_at=$19,
			consignor_address_id=$20,
			consignee_address_id=$21
		WHERE id=$22
	`,
			bilty.ConsignorCompanyID, bilty.ConsigneeCompanyID,
			bilty.FromLocation, bilty.ToLocation, bilty.Date, bilty.ToPay, bilty.GSTIN,
			bilty.InvNo, bilty.PVTMarks, bilty.PermitNo, bilty.ValueRupees, bilty.Remarks,
			bilty.Hamali, bilty.DDCharges, bilty.OtherCharges, bilty.FOV, bilty.Statistical,
			bilty.Status, time.Now().UTC(), bilty.ConsignorAddressID, bilty.ConsigneeAddressID, bilty.ID,
		)
		if err != nil {
			return err
		}

		// Refresh goods
		if _, err := tx.Exec(`DELETE FROM goods WHERE bilty_id=$1`, bilty.ID); err != nil {
			return err
		}
	}

	// Insert goods
	if err := r.insertGoods(tx, bilty.ID, bilty.Goods); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresBiltyRepo) hasAddressChanged(tx *sql.Tx, existingID int64, newAddr *models.BiltyAddress) (bool, error) {
	var existing models.BiltyAddress
	err := tx.QueryRow(`
		SELECT address_line, city, state, pincode
		FROM bilty_address
		WHERE id=$1
	`, existingID).Scan(&existing.AddressLine, &existing.City, &existing.State, &existing.Pincode)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, err
	}

	if existing.AddressLine != newAddr.AddressLine ||
		existing.City != newAddr.City ||
		existing.State != newAddr.State ||
		existing.Pincode != newAddr.Pincode {
		return true, nil
	}
	return false, nil
}

// ------------------------ GetBilty ------------------------

func (r *PostgresBiltyRepo) GetBilty(filters map[string]interface{}, single bool) ([]*models.Bilty, error) {
	query := `
		SELECT 
			b.id, b.bilty_no, b.consignor_company_id, b.consignee_company_id,
			b.consignor_address_id, b.consignee_address_id,
			b.from_location, b.to_location, b.date, b.to_pay, b.gstin, b.inv_no, b.pvt_marks, b.permit_no,
			b.value_rupees, b.remarks, b.hamali, b.dd_charges, b.other_charges, b.fov, b.statistical,
			b.created_by, b.created_at, b.status,

			-- Consignor company
			cc1.id, cc1.name, cc1.gstin, cc1.created_at,
			-- Consignee company
			cc2.id, cc2.name, cc2.gstin, cc2.created_at,

			-- Consignor address
			ca1.id, ca1.company_id, ca1.address_line, ca1.city, ca1.state, ca1.pincode, ca1.created_at,
			-- Consignee address
			ca2.id, ca2.company_id, ca2.address_line, ca2.city, ca2.state, ca2.pincode, ca2.created_at,

			-- Created by user
			u.id, u.name, u.email, u.role, u.created_at
		FROM bilty b
		LEFT JOIN company cc1 ON b.consignor_company_id = cc1.id
		LEFT JOIN company cc2 ON b.consignee_company_id = cc2.id
		LEFT JOIN bilty_address ca1 ON b.consignor_address_id = ca1.id
		LEFT JOIN bilty_address ca2 ON b.consignee_address_id = ca2.id
		LEFT JOIN app_user u ON b.created_by = u.id
	`

	args := []interface{}{}
	where := []string{}
	i := 1
	for k, v := range filters {
		where = append(where, fmt.Sprintf("b.%s = $%d", k, i))
		args = append(args, v)
		i++
	}
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	if !single {
		query += " ORDER BY b.created_at DESC"
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Bilty
	for rows.Next() {
		var b models.Bilty
		var consignorC, consigneeC models.Company
		var consignorA, consigneeA models.BiltyAddress
		var user models.AppUser

		err := rows.Scan(
			&b.ID, &b.BiltyNo, &b.ConsignorCompanyID, &b.ConsigneeCompanyID,
			&b.ConsignorAddressID, &b.ConsigneeAddressID,
			&b.FromLocation, &b.ToLocation, &b.Date, &b.ToPay, &b.GSTIN, &b.InvNo,
			&b.PVTMarks, &b.PermitNo, &b.ValueRupees, &b.Remarks,
			&b.Hamali, &b.DDCharges, &b.OtherCharges, &b.FOV, &b.Statistical,
			&b.CreatedBy, &b.CreatedAt, &b.Status,

			&consignorC.ID, &consignorC.Name, &consignorC.GSTIN, &consignorC.CreatedAt,
			&consigneeC.ID, &consigneeC.Name, &consigneeC.GSTIN, &consigneeC.CreatedAt,

			&consignorA.ID, &consignorA.CompanyID, &consignorA.AddressLine, &consignorA.City, &consignorA.State, &consignorA.Pincode, &consignorA.CreatedAt,
			&consigneeA.ID, &consigneeA.CompanyID, &consigneeA.AddressLine, &consigneeA.City, &consigneeA.State, &consigneeA.Pincode, &consigneeA.CreatedAt,

			&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if consignorC.ID != 0 {
			b.ConsignorCompany = &consignorC
		}
		if consigneeC.ID != 0 {
			b.ConsigneeCompany = &consigneeC
		}
		if consignorA.ID != 0 {
			b.ConsignorAddressSnap = &consignorA
		}
		if consigneeA.ID != 0 {
			b.ConsigneeAddressSnap = &consigneeA
		}
		if user.ID != 0 {
			b.CreatedByUser = &user
		}

		result = append(result, &b)
	}

	// Load all goods in one go (to avoid N+1)
	if len(result) > 0 {
		ids := make([]interface{}, len(result))
		idStrs := make([]string, len(result))
		for i, b := range result {
			ids[i] = b.ID
			idStrs[i] = fmt.Sprintf("$%d", i+1)
		}
		goodsQuery := fmt.Sprintf(`
			SELECT id, bilty_id, particulars, num_of_pkts, weight_kg, rate, per, amount
			FROM goods
			WHERE bilty_id IN (%s)
		`, strings.Join(idStrs, ","))
		goodsRows, _ := r.DB.Query(goodsQuery, ids...)
		defer goodsRows.Close()

		goodsMap := make(map[int64][]models.Goods)
		for goodsRows.Next() {
			var g models.Goods
			_ = goodsRows.Scan(&g.ID, &g.BiltyID, &g.Particulars, &g.NumOfPkts, &g.WeightKG, &g.Rate, &g.Per, &g.Amount)
			goodsMap[g.BiltyID] = append(goodsMap[g.BiltyID], g)
		}

		for _, b := range result {
			if g, ok := goodsMap[b.ID]; ok {
				b.Goods = g
			}
		}
	}

	if single {
		if len(result) > 0 {
			return []*models.Bilty{result[0]}, nil
		}
		return nil, nil
	}
	return result, nil
}

// ------------------------ PDF Helpers ------------------------

func (r *PostgresBiltyRepo) UpdatePDFCreatedAt(biltyID int64, t time.Time) error {
	_, err := r.DB.Exec("UPDATE bilty SET pdf_created_at = $1 WHERE id = $2", t, biltyID)
	return err
}

func (r *PostgresBiltyRepo) UpdatePDFInfo(id int64, path string, createdAt time.Time) error {
	query := `
		UPDATE bilty
		SET pdf_path = $1, pdf_created_at = $2
		WHERE id = $3
	`
	_, err := r.DB.Exec(query, path, createdAt, id)
	return err
}

// ------------------------ Delete Bilty ------------------------

func (r *PostgresBiltyRepo) DeleteBilty(biltyID int64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Fetch bilty addresses and companies
	var consignorAddrID, consigneeAddrID *int64
	var consignorCompanyID, consigneeCompanyID *int64
	err = tx.QueryRow(`
		SELECT consignor_address_id, consignee_address_id,
		       consignor_company_id, consignee_company_id
		FROM bilty WHERE id=$1
	`, biltyID).Scan(&consignorAddrID, &consigneeAddrID, &consignorCompanyID, &consigneeCompanyID)
	if err != nil {
		return err
	}

	// Delete goods linked to bilty
	if _, err := tx.Exec(`DELETE FROM goods WHERE bilty_id=$1`, biltyID); err != nil {
		return err
	}

	// Delete the bilty itself
	if _, err := tx.Exec(`DELETE FROM bilty WHERE id=$1`, biltyID); err != nil {
		return err
	}

	// Delete bilty addresses if unused
	deleteBiltyAddressIfUnused := func(addrID *int64) error {
		if addrID == nil {
			return nil
		}
		var count int
		err := tx.QueryRow(`
			SELECT COUNT(*) FROM bilty WHERE consignor_address_id=$1 OR consignee_address_id=$1
		`, *addrID).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err := tx.Exec(`DELETE FROM bilty_address WHERE id=$1`, *addrID)
			return err
		}
		return nil
	}
	if err := deleteBiltyAddressIfUnused(consignorAddrID); err != nil {
		return err
	}
	if err := deleteBiltyAddressIfUnused(consigneeAddrID); err != nil {
		return err
	}

	return tx.Commit()
}

// ------------------------ Get Bilty By ID ------------------------

func (r *PostgresBiltyRepo) GetBiltyByID(id int64) (*models.Bilty, error) {
	query := `
		SELECT id, updated_at, pdf_created_at, pdf_path
		FROM bilty
		WHERE id = $1
	`
	row := r.DB.QueryRow(query, id)

	var bilty models.Bilty
	err := row.Scan(&bilty.ID, &bilty.UpdatedAt, &bilty.PdfCreatedAt, &bilty.PdfPath)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &bilty, nil
}
