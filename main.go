package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}

	parcel.Number = id

	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	return parcel, nil
}

func (s ParcelService) ClientsParcel(client int) error {
	list, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}

	fmt.Println()
	for _, item := range list {
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			item.Number, item.Address, item.Client, item.CreatedAt, item.Status)
	}
	fmt.Println()

	return nil
}

func (s ParcelService) NextStatus(number int) error {
	parcel, err := s.store.Get(number)
	if err != nil {
		return err
	}

	var nextStatus string
	switch parcel.Status {
	case ParcelStatusRegistered:
		nextStatus = ParcelStatusSent
	case ParcelStatusSent:
		nextStatus = ParcelStatusDelivered
	case ParcelStatusDelivered:
		return nil
	}

	fmt.Printf("У посылки %d новый статус: %s\n", number, nextStatus)

	return s.store.SetStatus(number, nextStatus)
}

func (s ParcelService) ChangeAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

func main() {
	// настройте подключение к БД

	store := // создайте объект ParcelStore функцией NewParcelStore
	service := NewParcelService(store)

	// регистрация посылки
	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		panic(err)
	}

	// изменение адреса
	newAddress := "Псков, д. Пушкина, ул. Колотушкина, д. 25"
	err = service.ChangeAddress(p.Number, newAddress)
	if err != nil {
		panic(err)
	}

	// изменение статуса
	err = service.NextStatus(p.Number)
	if err != nil {
		panic(err)
	}

	// вывод посылок клиента
	err = service.ClientsParcel(client)
	if err != nil {
		panic(err)
	}

	// попытка удаления отправленной посылки
	err = service.Delete(p.Number)
	if err == nil {
		errMsg := "произошли нежелательные изменения: удалилась посылка со статусом, отличным от «registered»"
		panic(errors.New(errMsg))
	}

	// регистрация новой посылки
	p, err = service.Register(client, address)
	if err != nil {
		panic(err)
	}

	// удаление новой посылки
	err = service.Delete(p.Number)
	if err != nil {
		panic(err)
	}
}
