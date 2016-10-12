package samfs

import (
	"os"

	"github.com/golang/glog"
)

type DB struct {
	filePath string
	fd       *os.File
}

func NewDB(path string) (*DB, error) {
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		glog.Errorf("failed to open db file at path %s :: %v\n", path, err)
		return nil, err
	}

	db := &DB{
		filePath: path,
		fd:       fd,
	}
	return db, nil
}

//Lookup returns -1 if path does not exist in the database.
func (db *DB) Lookup(path string) (int64, error) {
	//TODO (arman)
	return 0, nil
}

//Increment returns after flushing data to disk.
//If path does not exist in database, it adds a new entry with value 0.
func (db *DB) Increment(path string) (int64, error) {
	//TODO (arman)
	return 0, nil
}

func (db *DB) Close() error {
	err := db.fd.Close()
	if err != nil {
		glog.Errorf("failed to close db file :: %v\n", err)
		return err
	}
	return nil
}
