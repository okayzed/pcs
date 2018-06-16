package sybil

import "compress/gzip"
import "fmt"
import "path"

import "io/ioutil"
import "math/rand"
import "os"
import "strconv"
import "testing"
import "time"
import "strings"

func TestOpenCompressedInfoDB(t *testing.T) {
	tableName := getTestTableName(t)
	deleteTestDb(tableName)
	defer deleteTestDb(tableName)

	blockCount := 3
	created := addRecords(tableName, func(r *Record, index int) {
		r.AddIntField("id", int64(index))
		age := int64(rand.Intn(20)) + 10
		r.AddIntField("age", age)
		r.AddStrField("age_str", strconv.FormatInt(int64(age), 10))
		r.AddIntField("time", int64(time.Now().Unix()))
		r.AddStrField("name", fmt.Sprint("user", index))
	}, blockCount)

	nt := saveAndReloadTable(t, tableName, blockCount)

	if nt.Name != tableName {
		t.Error("TEST TABLE NAME INCORRECT")
	}

	filename := fmt.Sprintf("db/%s/info.db", tableName)
	dat, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("ERR", err)
		t.Error("Test table did not create info.db")
	}

	// NOW WE COMPRESS INFO.DB.GZ
	zfilename := fmt.Sprintf("db/%s/info.db.gz", tableName)
	file, err := os.Create(zfilename)
	if err != nil {
		t.Error("COULDNT LOAD ZIPPED TABLE FILE FOR WRITING!")

	}
	zinfo := gzip.NewWriter(file)
	if _, err := zinfo.Write(dat); err != nil {
		t.Error(err)
	}
	if err := zinfo.Close(); err != nil {
		t.Error(err)
	}

	if err := os.RemoveAll(filename); err != nil {
		t.Error(err)
	}
	// END ZIPPING INFO.DB.GZ

	loadSpec := nt.NewLoadSpec()
	loadSpec.LoadAllColumns = true

	if err := nt.LoadTableInfo(); err != nil {
		t.Error("COULDNT LOAD ZIPPED TABLE INFO!", err)
	}

	if _, err := nt.LoadRecords(&loadSpec); err != nil {
		t.Error(err)
	}

	var records = make([]*Record, 0)
	for _, b := range nt.BlockList {
		records = append(records, b.RecordList...)
	}

	if len(records) != len(created) {
		t.Error("More records were created than expected", len(records))
	}

}

func TestOpenCompressedColumn(t *testing.T) {
	tableName := getTestTableName(t)
	deleteTestDb(tableName)
	defer deleteTestDb(tableName)

	blockCount := 3
	created := addRecords(tableName, func(r *Record, index int) {
		r.AddIntField("id", int64(index))
		age := int64(rand.Intn(20)) + 10
		r.AddIntField("age", age)
		r.AddStrField("age_str", strconv.FormatInt(int64(age), 10))
		r.AddIntField("time", int64(time.Now().Unix()))
		r.AddStrField("name", fmt.Sprint("user", index))
	}, blockCount)

	nt := saveAndReloadTable(t, tableName, blockCount)
	if err := nt.DigestRecords(); err != nil {
		t.Error(err)
	}
	if _, err := nt.LoadRecords(nil); err != nil {
		t.Error(err)
	}

	blocks := nt.BlockList

	if nt.Name != tableName {
		t.Error("TEST TABLE NAME INCORRECT")
	}

	// NOW WE COMPRESS ALL THE BLOCK FILES BY ITERATING THROUGH THE DIR AND
	// DOING SO
	for blockname := range blocks {
		files, _ := ioutil.ReadDir(blockname)
		Debug("READING BLOCKNAME", blockname)
		for _, f := range files {
			filename := path.Join(blockname, f.Name())
			if !strings.HasSuffix(filename, ".db") {
				continue
			}
			dat, _ := ioutil.ReadFile(filename)

			zfilename := fmt.Sprintf("%s.gz", filename)
			file, err := os.Create(zfilename)
			if err != nil {
				t.Error("COULDNT LOAD ZIPPED TABLE FILE FOR WRITING!")

			}
			zinfo := gzip.NewWriter(file)
			if _, err := zinfo.Write(dat); err != nil {
				t.Error(err)
			}
			if err := zinfo.Close(); err != nil {
				t.Error(err)
			}
			Debug("CREATED GZIP FILE", zfilename)

			err = os.RemoveAll(filename)
			Debug("REMOVED", filename, err)

		}
	}

	// END COMPRESSING BLOCK FILES

	bt := saveAndReloadTable(t, tableName, blockCount)

	loadSpec := bt.NewLoadSpec()
	loadSpec.LoadAllColumns = true

	if err := bt.LoadTableInfo(); err != nil {
		t.Error("COULDNT LOAD ZIPPED TABLE INFO!", err)
	}

	if _, err := bt.LoadRecords(&loadSpec); err != nil {
		t.Error(err)
	}

	var records = make([]*Record, 0)
	for _, b := range bt.BlockList {
		records = append(records, b.RecordList...)
	}

	if len(records) != len(created) {
		t.Error("More records were created than expected", len(records))
	}

}
