package database

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/giornetta/watchazon"
)

type Database struct {
	db *badger.DB
}

type Record struct {
	*watchazon.Product
	Users []int64
}

func (r *Record) Encode() ([]byte, error) {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(&r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecodeProduct(b []byte) (*Record, error) {
	reader := bytes.NewReader(b)

	var r Record
	if err := gob.NewDecoder(reader).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

func Open(path string) (*Database, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}

func (db *Database) Close() {
	_ = db.db.Close()
}

func (db *Database) Get(link string) (*Record, error) {
	key := []byte(link)
	var r *Record
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			r, err = DecodeProduct(val)
			return err
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (db *Database) GetAll() ([]*Record, error) {
	records := make([]*Record, 0)
	err := db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var rec *Record
			err := item.Value(func(val []byte) error {
				r, err := DecodeProduct(val)
				if err != nil {
					return err
				}

				rec = r
				return nil
			})
			if err != nil {
				return nil
			}
			records = append(records, rec)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (db *Database) Update(product *watchazon.Product, userID int64) error {
	key := []byte(product.Link)

	return db.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		var r *Record
		err = item.Value(func(val []byte) error {
			r, err = DecodeProduct(val)
			return err
		})
		if err != nil {
			return err
		}

		r.Product = product
		if userID != 0 && !contains(r.Users, userID) {
			r.Users = append(r.Users, userID)
		}

		b, err := r.Encode()
		if err != nil {
			return err
		}

		return txn.Set(key, b)
	})
}

func (db *Database) Insert(product *watchazon.Product, userID int64) error {
	key := []byte(product.Link)

	return db.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil {
			return fmt.Errorf("already exists")
		}

		p := &Record{
			Product: product,
			Users:   []int64{userID},
		}
		b, err := p.Encode()
		if err != nil {
			return err
		}

		return txn.Set(key, b)
	})
}

func (db *Database) GetUserWatchList(userID int64) ([]*Record, error) {
	records := make([]*Record, 0)
	err := db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var rec *Record
			err := item.Value(func(val []byte) error {
				r, err := DecodeProduct(val)
				if err != nil {
					return err
				}
				rec = r
				return nil
			})
			if err != nil {
				continue
			}
			if contains(rec.Users, userID) {
				records = append(records, rec)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (db *Database) RemoveFromWatchList(link string, userID int64) error {
	key := []byte(link)

	return db.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		var r *Record
		err = item.Value(func(val []byte) error {
			r, err = DecodeProduct(val)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		for i, u := range r.Users {
			if u == userID {
				r.Users = append(r.Users[:i], r.Users[i+1:]...)
			}
		}

		if len(r.Users) == 0 {
			return txn.Delete(key)
		}

		b, err := r.Encode()
		if err != nil {
			return err
		}

		return txn.Set(key, b)
	})
}

func contains(slice []int64, val int64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
