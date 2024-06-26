package toy_store

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type ToyStoreRecord struct {
	ID        uint
	Key       string
	Value     string
	ExpiredAt time.Time
}

type ToyStore struct {
	db *sql.DB
}

func NewToyStore(db *sql.DB) *ToyStore {
	return &ToyStore{db: db}
}

func (toystore *ToyStore) Set(key string, value string, expiredAt time.Time) (bool, error) {

	sqlStatement := `
	INSERT INTO public.toy_dynamo (key, value, expired_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (key) DO UPDATE
	SET value = EXCLUDED.value, expired_at = EXCLUDED.expired_at;
	`
	_, err := toystore.db.Exec(sqlStatement, key, value, expiredAt)
	if err != nil {
		log.Printf("Failed to insert or update record: %v", err)
		return false, err
	}
	return true, nil

}

func (toystore *ToyStore) Get(key string) (ToyStoreRecord, error) {

	sqlStatement := `SELECT id, key, value, expired_at FROM toy_dynamo 
	WHERE key = $1 AND expired_at > NOW()`

	var record ToyStoreRecord
	err := toystore.db.QueryRow(sqlStatement, key).Scan(&record.ID, &record.Key, &record.Value, &record.ExpiredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No record found for key: %v", key)
			return ToyStoreRecord{}, nil
		}
		log.Printf("Failed to retrieve record: %v", err)
		return ToyStoreRecord{}, err
	}
	return record, nil
}

func (toystore *ToyStore) Delete(key string) (bool, error) {

	sqlStatement := `
	UPDATE toy_dynamo set expired_at = TIMESTAMP '1970-01-01 00:00:00' where key = $1 
	and expired_at > NOW();
	`
	_, err := toystore.db.Exec(sqlStatement, key)
	if err != nil {
		log.Printf("Failed to delete record for key:%s due to error: %v", key, err)
		return false, err
	}
	return true, nil

}
