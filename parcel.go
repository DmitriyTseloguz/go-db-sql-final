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

func (s ParcelStore) readData(dataScanner DBScanner) (Parcel, error) {
	var parcel = Parcel{}

	var err = dataScanner.Scan(
		&parcel.Number,
		&parcel.Client,
		&parcel.Status,
		&parcel.Address,
		&parcel.CreatedAt,
	)

	return parcel, err
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	var result, err = s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt,
	)

	if err != nil {
		return 0, err
	}

	parcelID, err := result.LastInsertId()

	return int(parcelID), err
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	var err error

	var parcelData = s.db.QueryRow("SELECT * FROM parcel WHERE number = ?", number)

	err = parcelData.Err()

	if err != nil {
		return Parcel{}, err
	}

	var parcel = Parcel{}

	parcel, err = s.readData(parcelData)

	return parcel, err
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var parcelsData, err = s.db.Query("SELECT * FROM parcel WHERE client = ?", client)

	defer parcelsData.Close()

	if err != nil {
		return nil, err
	}

	var clientParcels []Parcel

	for parcelsData.Next() {
		parcel, err := s.readData(parcelsData)

		if err != nil {
			return clientParcels, err
		}

		clientParcels = append(clientParcels, parcel)
	}

	return clientParcels, parcelsData.Err()
}

func (s ParcelStore) SetStatus(number int, status string) error {
	var _, err = s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)

	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	var parcel, err = s.Get(number)

	if err != nil {
		return err
	}

	if parcel.Status != ParcelStatusRegistered {
		return errors.New("you can change the address only if the status value is registered")
	}

	_, err = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)

	return err
}

func (s ParcelStore) Delete(number int) error {
	var parcel, err = s.Get(number)

	if err != nil {
		return err
	}

	if parcel.Status != ParcelStatusRegistered {
		return errors.New("you can delete a row only if the status value is registered")
	}

	_, err = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)

	return err
}
