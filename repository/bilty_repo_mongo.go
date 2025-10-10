package repository

import (
	"context"
	"errors"
	"time"

	"hariomtransport/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoBiltyRepo struct {
	DB *mongo.Client
}

func NewMongoBiltyRepo(db *mongo.Client) *MongoBiltyRepo {
	return &MongoBiltyRepo{DB: db}
}

// CreateBiltyWithParties inserts a bilty document with nested companies, addresses, and goods
func (r *MongoBiltyRepo) CreateBiltyWithParties(bilty *models.Bilty) error {
	ctx := context.Background()
	db := r.DB.Database("hariomtransport")

	if bilty.CreatedAt.IsZero() {
		bilty.CreatedAt = time.Now().UTC()
	}

	// Upsert app_user if provided
	if bilty.CreatedByUser != nil {
		if bilty.CreatedByUser.ID == 0 {
			bilty.CreatedByUser.ID = bilty.CreatedBy
		}
		_, _ = db.Collection("app_user").
			UpdateOne(ctx,
				bson.M{"_id": bilty.CreatedByUser.ID},
				bson.M{"$setOnInsert": bilty.CreatedByUser},
				options.Update().SetUpsert(true),
			)
		if bilty.CreatedBy == 0 {
			bilty.CreatedBy = bilty.CreatedByUser.ID
		}
	}

	// Upsert companies
	upsertCompany := func(comp *models.Company, idPtr *int64) {
		if comp == nil {
			return
		}
		if comp.ID == 0 && idPtr != nil {
			comp.ID = *idPtr
		}
		_, _ = db.Collection("company").
			UpdateOne(ctx,
				bson.M{"_id": comp.ID},
				bson.M{"$setOnInsert": comp},
				options.Update().SetUpsert(true),
			)
		if idPtr != nil {
			*idPtr = comp.ID
		}
	}

	upsertCompany(bilty.ConsignorCompany, bilty.ConsignorCompanyID)
	upsertCompany(bilty.ConsigneeCompany, bilty.ConsigneeCompanyID)

	// Insert bilty_address snapshots
	insertAddress := func(addr *models.BiltyAddress, companyID *int64, idPtr *int64) error {
		if addr == nil {
			return nil
		}
		if addr.ID == 0 && idPtr != nil {
			addr.ID = *idPtr
		}
		addr.CompanyID = companyID
		_, err := db.Collection("bilty_address").InsertOne(ctx, addr)
		if err != nil {
			return err
		}
		if idPtr != nil {
			*idPtr = addr.ID
		}
		return nil
	}

	if err := insertAddress(bilty.ConsignorAddressSnap, bilty.ConsignorCompanyID, bilty.ConsignorAddressID); err != nil {
		return err
	}
	if err := insertAddress(bilty.ConsigneeAddressSnap, bilty.ConsigneeCompanyID, bilty.ConsigneeAddressID); err != nil {
		return err
	}

	// Insert main bilty
	_, err := db.Collection("bilty").InsertOne(ctx, bilty)
	if err != nil {
		return err
	}

	// Insert goods
	for _, g := range bilty.Goods {
		g.BiltyID = bilty.ID
		_, err := db.Collection("goods").InsertOne(ctx, g)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetBilty fetches bilties from MongoDB; single=true fetches one record
func (r *MongoBiltyRepo) GetBilty(filters map[string]interface{}, single bool) ([]*models.Bilty, error) {
	ctx := context.Background()
	db := r.DB.Database("hariomtransport")

	bsonFilter := bson.M{}
	if filters != nil {
		for k, v := range filters {
			bsonFilter[k] = v
		}
	}

	var cur *mongo.Cursor
	var err error

	if single {
		var b models.Bilty
		err = db.Collection("bilty").FindOne(ctx, bsonFilter).Decode(&b)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return []*models.Bilty{}, nil
			}
			return nil, err
		}
		cur = &mongo.Cursor{} // dummy cursor to reuse logic
		return []*models.Bilty{r.populateNested(&b, ctx, db)}, nil
	} else {
		cur, err = db.Collection("bilty").Find(ctx, bsonFilter)
		if err != nil {
			return nil, err
		}
		defer cur.Close(ctx)
	}

	var out []*models.Bilty
	for cur.Next(ctx) {
		var b models.Bilty
		if err := cur.Decode(&b); err != nil {
			return nil, err
		}
		out = append(out, r.populateNested(&b, ctx, db))
	}

	return out, nil
}

// populateNested loads nested companies, addresses, goods, and created_by user
func (r *MongoBiltyRepo) populateNested(b *models.Bilty, ctx context.Context, db *mongo.Database) *models.Bilty {
	if b.ConsignorCompanyID != nil && *b.ConsignorCompanyID != 0 {
		var c models.Company
		_ = db.Collection("company").FindOne(ctx, bson.M{"_id": *b.ConsignorCompanyID}).Decode(&c)
		b.ConsignorCompany = &c
	}
	if b.ConsigneeCompanyID != nil && *b.ConsigneeCompanyID != 0 {
		var c models.Company
		_ = db.Collection("company").FindOne(ctx, bson.M{"_id": *b.ConsigneeCompanyID}).Decode(&c)
		b.ConsigneeCompany = &c
	}
	if b.ConsignorAddressID != nil && *b.ConsignorAddressID != 0 {
		var a models.BiltyAddress
		_ = db.Collection("bilty_address").FindOne(ctx, bson.M{"_id": *b.ConsignorAddressID}).Decode(&a)
		b.ConsignorAddressSnap = &a
	}
	if b.ConsigneeAddressID != nil && *b.ConsigneeAddressID != 0 {
		var a models.BiltyAddress
		_ = db.Collection("bilty_address").FindOne(ctx, bson.M{"_id": *b.ConsigneeAddressID}).Decode(&a)
		b.ConsigneeAddressSnap = &a
	}
	if b.CreatedBy != 0 {
		var u models.AppUser
		_ = db.Collection("app_user").FindOne(ctx, bson.M{"_id": b.CreatedBy}).Decode(&u)
		b.CreatedByUser = &u
	}
	// Goods
	goodsCur, _ := db.Collection("goods").Find(ctx, bson.M{"bilty_id": b.ID})
	var goodsList []models.Goods
	for goodsCur.Next(ctx) {
		var g models.Goods
		_ = goodsCur.Decode(&g)
		goodsList = append(goodsList, g)
	}
	goodsCur.Close(ctx)
	b.Goods = goodsList

	return b
}
func (r *MongoBiltyRepo) GetBiltyByID(id int64) (*models.Bilty, error) {
	db := r.DB.Database("hariomtransport")
	collection := db.Collection("bilty")
	filter := bson.M{"id": id}

	var bilty models.Bilty
	err := collection.FindOne(context.Background(), filter).Decode(&bilty)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &bilty, nil
}

func (r *MongoBiltyRepo) UpdatePDFInfo(id int64, path string, createdAt time.Time) error {
	db := r.DB.Database("hariomtransport")
	collection := db.Collection("bilty")

	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
			"pdf_path":       path,
			"pdf_created_at": createdAt,
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *MongoBiltyRepo) DeleteBilty(biltyID int64) error {
	ctx := context.Background()
	db := r.DB.Database("hariomtransport")

	// 1. Fetch bilty
	var b models.Bilty
	err := db.Collection("bilty").FindOne(ctx, bson.M{"_id": biltyID}).Decode(&b)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}

	// 2. Delete goods
	_, _ = db.Collection("goods").DeleteMany(ctx, bson.M{"bilty_id": biltyID})

	// 3. Delete bilty itself
	_, _ = db.Collection("bilty").DeleteOne(ctx, bson.M{"_id": biltyID})

	// Helper: delete bilty_address if not used in any other bilty
	checkAndDeleteBiltyAddress := func(addrID *int64) error {
		if addrID == nil {
			return nil
		}
		count, err := db.Collection("bilty").CountDocuments(ctx, bson.M{
			"$or": []bson.M{
				{"consignor_address_id": *addrID},
				{"consignee_address_id": *addrID},
			},
		})
		if err != nil {
			return err
		}
		if count == 0 {
			var addr models.BiltyAddress
			if err := db.Collection("bilty_address").FindOne(ctx, bson.M{"_id": *addrID}).Decode(&addr); err == nil {
				_, _ = db.Collection("bilty_address").DeleteOne(ctx, bson.M{"_id": *addrID})

				// Delete company_address if not used anywhere
				if addr.CompanyID != nil {
					count2, _ := db.Collection("bilty_address").CountDocuments(ctx, bson.M{"company_id": addr.CompanyID})
					if count2 == 0 {
						_, _ = db.Collection("company_address").DeleteOne(ctx, bson.M{"_id": addr.CompanyID})
					}
				}
			}
		}
		return nil
	}

	// Delete bilty addresses and associated company addresses if safe
	if err := checkAndDeleteBiltyAddress(b.ConsignorAddressID); err != nil {
		return err
	}
	if err := checkAndDeleteBiltyAddress(b.ConsigneeAddressID); err != nil {
		return err
	}

	return nil
}
