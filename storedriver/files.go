package storedriver

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/dgraph-io/badger"
)

type Badger struct {
	ValueDir string
	Db       *badger.DB
}

func (bd *Badger) Open() error {
	var err error
	checkPath(bd.ValueDir)
	bd.Db, err = badger.Open(badger.DefaultOptions(bd.ValueDir))
	return err
}

func (bd *Badger) Close() {
	bd.Db.Close()
}

func (bd *Badger) Add(key string, data map[string]interface{}) error {
	insert_data, err := json.Marshal(data)
	if err != nil {
		return err
	}
	addhandle := bd.Db.NewTransaction(true)
	defer addhandle.Discard()
	if err := addhandle.Set([]byte(key), insert_data); err == nil {
		err = addhandle.Commit()
		return err
	} else {
		return err
	}

}

func (bd *Badger) Delete(host string) error {
	get := bd.Db.NewTransaction(false)
	defer get.Discard()
	if _, err := get.Get([]byte(host)); err == nil {
		delTxn := bd.Db.NewTransaction(true)
		defer delTxn.Discard()
		err := delTxn.Delete([]byte(host))
		if err == nil {
			err = delTxn.Commit()

		}
		return err
	}
	return errors.New("cannot get")

}

func (bd *Badger) GetAll() []map[string]interface{} {
	var datas []map[string]interface{}
	txn := bd.Db.NewTransaction(false)
	defer txn.Discard()
	iter := badger.DefaultIteratorOptions
	it := txn.NewIterator(iter)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		data, _ := item.ValueCopy(nil)
		tem := make(map[string]interface{})
		err := json.Unmarshal(data, &tem)
		if err == nil {
			datas = append(datas, tem)
		}

	}
	return datas

}

func checkPath(path string) {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}

}
func init() {

}
