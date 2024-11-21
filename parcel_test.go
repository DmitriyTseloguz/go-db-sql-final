package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, openDBError := sql.Open("sqlite", "tracker.db")

	require.NoError(t, openDBError)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	var id, addError = store.Add(parcel)

	require.NoError(t, addError)
	assert.NotEmpty(t, id)

	parcel.Number = id

	var addedParcel, selectError = store.Get(parcel.Number)

	require.NoError(t, selectError)
	require.Equal(t, parcel, addedParcel)

	var deleteError = store.Delete(parcel.Number)

	assert.NoError(t, deleteError)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, openDBError := sql.Open("sqlite", "tracker.db")

	require.NoError(t, openDBError)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	var id, addError = store.Add(parcel)

	require.NoError(t, addError)
	assert.NotEmpty(t, id)

	parcel.Number = id

	newAddress := "new test address"
	var changeAddressError = store.SetAddress(parcel.Number, newAddress)

	require.NoError(t, changeAddressError)

	var storedParcel, selectError = store.Get(parcel.Number)

	require.NoError(t, selectError)
	assert.Equal(t, newAddress, storedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, openDBError := sql.Open("sqlite", "tracker.db")

	require.NoError(t, openDBError)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	var id, addError = store.Add(parcel)

	require.NoError(t, addError)
	assert.NotEmpty(t, id)

	parcel.Number = id

	var statusCahgeError = store.SetStatus(parcel.Number, ParcelStatusSent)

	require.NoError(t, statusCahgeError)

	var addedParcel, selectError = store.Get(parcel.Number)

	require.NoError(t, selectError)

	assert.Equal(t, ParcelStatusSent, addedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, openDBError := sql.Open("sqlite", "tracker.db")

	require.NoError(t, openDBError)

	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])

		require.NoError(t, err)
		require.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	// check
	for _, storedParcel := range storedParcels {

		var mappedParcel, isExist = parcelMap[storedParcel.Number]

		require.True(t, isExist)

		assert.Equal(t, mappedParcel, storedParcel)
	}
}
