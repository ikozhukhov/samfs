package samfs

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

//TODO (arman): need a lock per entry?
type DB struct {
	filePath string
	entries  map[string]int64
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		filePath: path,
	}

	err := db.readIntoMemory()
	if err != nil {
		glog.Errorf("failed to read database file at path %s :: %v\n", path, err)
		return nil, err
	}

	return db, nil
}

func (db *DB) readIntoMemory() error {
	fd, err := os.OpenFile(db.filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		glog.Errorf("failed to open db file at path %s :: %v\n", db.filePath, err)
		return err
	}

	db.entries = make(map[string]int64) //reset in memory data structure

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("line: %s\n", line)
		tokens := strings.Split(line, "|")
		fmt.Printf("tokens: %v\n", tokens)

		num, err := strconv.ParseInt(tokens[1], 10, 64)
		if err != nil {
			glog.Errorf("failed to parse %s as int64 :: %v\n", tokens[1], err)
			db.entries = nil //reset in memory data structure
			return err
		}

		db.entries[tokens[0]] = num
	}

	return nil
}

//flushes to disk first, and then
func (db *DB) writeToDisk() error {
	rollback := false

	//use a tmp file
	tmpFilePath := db.filePath + ".tmp"
	fd, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		0666)
	if err != nil {
		glog.Errorf("failed to open tmp db file at path %s :: %v\n", tmpFilePath,
			err)
		return err
	}
	defer func() {
		fd.Close()
		if rollback {
			err := os.Remove(tmpFilePath)
			if err != nil {
				glog.Errorf("failed to delete tmp db file :: %v\n", err)
			}
		}
	}()

	//write in-memory data structure to tmp file
	for path, num := range db.entries {
		line := fmt.Sprintf("%s|%d\n", path, num)
		_, err := fd.Write([]byte(line))
		if err != nil {
			glog.Errorf("could not write to tmp db file :: %v\n", err)
			rollback = true
			return err
		}
	}

	//flush tmp file to disk
	err = fd.Sync()
	if err != nil {
		glog.Errorf("could not fsync tmp db file :: %v\n", err)
		rollback = true
		return err
	}

	//overwrite db file with tmp file
	err = os.Rename(tmpFilePath, db.filePath)
	if err != nil {
		glog.Errorf("could not move tmp db file to db file :: %v\n", err)
		rollback = true
		return err
	}

	//flush db file just in case
	err = flush(db.filePath)
	if err != nil {
		return err
	}

	//flush db file's directory to make sure metadata is persisted
	err = flush(path.Dir(db.filePath))
	if err != nil {
		return err
	}

	return nil
}

func flush(path string) error {
	fd, err := os.Open(path)
	if err != nil {
		glog.Errorf("failed to open file/dir at path %s :: %v\n", path, err)
		return err
	}
	defer fd.Close()

	err = fd.Sync()
	if err != nil {
		glog.Errorf("could not fsync file/dir at path %s :: %v\n", path, err)
		return err
	}

	return nil
}

//Lookup returns -1 if path does not exist in the database.
func (db *DB) Lookup(path string) int64 {
	path = strings.TrimSpace(path)
	num, ok := db.entries[path]
	if !ok {
		return -1
	}
	return num
}

//Increment returns after flushing data to disk.
//If path does not exist in database, it adds a new entry with value 0.
func (db *DB) Increment(path string) (int64, error) {
	path = strings.TrimSpace(path)
	num, ok := db.entries[path]
	if !ok {
		num = -1
	}

	num += 1
	db.entries[path] = num
	err := db.writeToDisk()
	if err != nil {
		glog.Errorf("failed to persist incrementing %s to disk :: %v\n", path, err)
		return -1, err
	}

	return num, nil
}
