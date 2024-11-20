package main

import (
	"database/sql"
	"errors"
)

type DBScanner interface {
	Scan(...any) error
}

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) readData(dataScanner DBScanner) Parcel {
	var parcel = Parcel{}

	dataScanner.Scan(
		&parcel.Number,
		&parcel.Client,
		&parcel.Status,
		&parcel.Address,
		&parcel.CreatedAt,
	)

	return parcel
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	var result, insertError = s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt,
	)

	if insertError != nil {
		return 0, insertError
	}

	var parcelID, idError = result.LastInsertId()

	if idError != nil {
		return 0, idError
	}

	return int(parcelID), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	var parcelData = s.db.QueryRow("SELECT * FROM parcel WHERE number = ?", number)

	var selectError = parcelData.Err()

	if selectError != nil {
		return Parcel{}, selectError
	}

	parcel := Parcel{}

	parcel = s.readData(parcelData)

	return parcel, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var parcelsData, selectError = s.db.Query("SELECT * FROM parcel WHERE client = ?", client)

	defer parcelsData.Close()

	if selectError != nil {
		return nil, selectError
	}

	var clientParcels []Parcel

	for parcelsData.Next() {
		clientParcels = append(clientParcels, s.readData(parcelsData))
	}

	return clientParcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	var _, updateError = s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)

	if updateError != nil {
		return updateError
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	var parcel, selectError = s.Get(number)

	if selectError != nil {
		return selectError
	}

	if parcel.Status != ParcelStatusRegistered {
		return errors.New("you can change the address only if the status value is registered")
	}

	var _, updateError = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)

	if updateError != nil {
		return updateError
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	var parcel, selectError = s.Get(number)

	if selectError != nil {
		return selectError
	}

	if parcel.Status != ParcelStatusRegistered {
		return errors.New("you can delete a row only if the status value is registered")
	}

	var _, deleteError = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)

	if deleteError != nil {
		return deleteError
	}

	return nil
}
