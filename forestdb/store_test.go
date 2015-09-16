//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package forestdb

import (
	"os"
	"reflect"
	"testing"

	"github.com/blevesearch/bleve/index/store"
)

func TestForestDBStore(t *testing.T) {
	defer func() {
		err := os.RemoveAll("testdir")
		if err != nil {
			t.Fatal(err)
		}
	}()

	err := os.MkdirAll("testdir", 0700)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New("testdir/test", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := s.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	CommonTestKVStore(t, s)
}

func TestReaderIsolation(t *testing.T) {
	defer func() {
		err := os.RemoveAll("testdir")
		if err != nil {
			t.Fatal(err)
		}
	}()

	err := os.MkdirAll("testdir", 0700)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New("testdir/test", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := s.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	CommonTestReaderIsolation(t, s)
}

// TestRollbackSameHandle tries to rollback a handle
// and ensure that subsequent reads from it also
// reflect the rollback
func TestRollbackSameHandle(t *testing.T) {
	defer func() {
		err := os.RemoveAll("testdir")
		if err != nil {
			t.Fatal(err)
		}
	}()

	err := os.MkdirAll("testdir", 0700)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New("testdir/test", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := s.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	writer, err := s.Writer()
	if err != nil {
		t.Fatal(err)
	}

	// create 2 docs, a and b
	err = writer.Set([]byte("a"), []byte("val-a"))
	if err != nil {
		t.Error(err)
	}

	err = writer.Set([]byte("b"), []byte("val-b"))
	if err != nil {
		t.Error(err)
	}

	// get the rollback id
	rollbackId, err := s.GetRollbackID()
	if err != nil {
		t.Error(err)
	}

	// create a 3rd doc c
	err = writer.Set([]byte("c"), []byte("val-c"))
	if err != nil {
		t.Error(err)
	}

	err = writer.Close()
	if err != nil {
		t.Error(err)
	}

	// make sure c is there
	reader, err := s.Reader()
	if err != nil {
		t.Error(err)
	}
	val, err := reader.Get([]byte("c"))
	if err != nil {
		t.Error(err)
	}
	if string(val) != "val-c" {
		t.Errorf("expected value 'val-c' got '%s'", val)
	}
	err = reader.Close()
	if err != nil {
		t.Fatal(err)
	}

	// now rollback
	err = s.RollbackTo(rollbackId)
	if err != nil {
		t.Fatal(err)
	}

	// now make sure c is not there
	reader, err = s.Reader()
	if err != nil {
		t.Error(err)
	}
	val, err = reader.Get([]byte("c"))
	if err != nil {
		t.Error(err)
	}
	if val != nil {
		t.Errorf("expected missing, got '%s'", val)
	}
	err = reader.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// TestRollbackNewHandle tries to rollback the
// database, then opens a new handle, and ensures
// that the rollback is reflected there as well
func TestRollbackNewHandle(t *testing.T) {
	defer func() {
		err := os.RemoveAll("testdir")
		if err != nil {
			t.Fatal(err)
		}
	}()

	err := os.MkdirAll("testdir", 0700)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New("testdir/test", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := s.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	writer, err := s.Writer()
	if err != nil {
		t.Fatal(err)
	}

	// create 2 docs, a and b
	err = writer.Set([]byte("a"), []byte("val-a"))
	if err != nil {
		t.Error(err)
	}

	err = writer.Set([]byte("b"), []byte("val-b"))
	if err != nil {
		t.Error(err)
	}

	// get the rollback id
	rollbackId, err := s.GetRollbackID()
	if err != nil {
		t.Error(err)
	}

	// create a 3rd doc c
	err = writer.Set([]byte("c"), []byte("val-c"))
	if err != nil {
		t.Error(err)
	}

	err = writer.Close()
	if err != nil {
		t.Error(err)
	}

	// make sure c is there
	reader, err := s.Reader()
	if err != nil {
		t.Error(err)
	}
	val, err := reader.Get([]byte("c"))
	if err != nil {
		t.Error(err)
	}
	if string(val) != "val-c" {
		t.Errorf("expected value 'val-c' got '%s'", val)
	}
	err = reader.Close()
	if err != nil {
		t.Fatal(err)
	}

	// now rollback
	err = s.RollbackTo(rollbackId)
	if err != nil {
		t.Fatal(err)
	}

	// now lets open another handle
	s2, err := New("testdir/test", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s2.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	// now make sure c is not there
	reader2, err := s2.Reader()
	if err != nil {
		t.Error(err)
	}
	val, err = reader2.Get([]byte("c"))
	if err != nil {
		t.Error(err)
	}
	if val != nil {
		t.Errorf("expected missing, got '%s'", val)
	}
	err = reader2.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// TestRollbackOtherHandle tries to create 2 handles
// at the beginning, then rollback one of them
// and ensure it affects the other
func TestRollbackOtherHandle(t *testing.T) {
	defer func() {
		err := os.RemoveAll("testdir")
		if err != nil {
			t.Fatal(err)
		}
	}()

	err := os.MkdirAll("testdir", 0700)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New("testdir/test", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := s.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// open another handle at the same time
	s2, err := New("testdir/test", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s2.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	writer, err := s.Writer()
	if err != nil {
		t.Fatal(err)
	}

	// create 2 docs, a and b
	err = writer.Set([]byte("a"), []byte("val-a"))
	if err != nil {
		t.Error(err)
	}

	err = writer.Set([]byte("b"), []byte("val-b"))
	if err != nil {
		t.Error(err)
	}

	// get the rollback id
	rollbackId, err := s.GetRollbackID()
	if err != nil {
		t.Error(err)
	}

	// create a 3rd doc c
	err = writer.Set([]byte("c"), []byte("val-c"))
	if err != nil {
		t.Error(err)
	}

	err = writer.Close()
	if err != nil {
		t.Error(err)
	}

	// make sure c is there
	reader, err := s.Reader()
	if err != nil {
		t.Error(err)
	}
	val, err := reader.Get([]byte("c"))
	if err != nil {
		t.Error(err)
	}
	if string(val) != "val-c" {
		t.Errorf("expected value 'val-c' got '%s'", val)
	}
	err = reader.Close()
	if err != nil {
		t.Fatal(err)
	}

	// now rollback
	err = s.RollbackTo(rollbackId)
	if err != nil {
		t.Fatal(err)
	}

	// now make sure c is not on the other handle
	reader2, err := s2.Reader()
	if err != nil {
		t.Error(err)
	}
	val, err = reader2.Get([]byte("c"))
	if err != nil {
		t.Error(err)
	}
	if val != nil {
		t.Errorf("expected missing, got '%s'", val)
	}
	err = reader2.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func CommonTestKVStore(t *testing.T, s store.KVStore) {

	writer, err := s.Writer()
	if err != nil {
		t.Error(err)
	}
	err = writer.Set([]byte("a"), []byte("val-a"))
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Set([]byte("z"), []byte("val-z"))
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Delete([]byte("z"))
	if err != nil {
		t.Fatal(err)
	}

	batch := writer.NewBatch()
	batch.Set([]byte("b"), []byte("val-b"))
	batch.Set([]byte("c"), []byte("val-c"))
	batch.Set([]byte("d"), []byte("val-d"))
	batch.Set([]byte("e"), []byte("val-e"))
	batch.Set([]byte("f"), []byte("val-f"))
	batch.Set([]byte("g"), []byte("val-g"))
	batch.Set([]byte("h"), []byte("val-h"))
	batch.Set([]byte("i"), []byte("val-i"))
	batch.Set([]byte("j"), []byte("val-j"))

	err = batch.Execute()
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	reader, err := s.Reader()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	it := reader.Iterator([]byte("b"))
	key, val, valid := it.Current()
	if !valid {
		t.Fatalf("valid false, expected true")
	}
	if string(key) != "b" {
		t.Fatalf("expected key b, got %s", key)
	}
	if string(val) != "val-b" {
		t.Fatalf("expected value val-b, got %s", val)
	}

	it.Next()
	key, val, valid = it.Current()
	if !valid {
		t.Fatalf("valid false, expected true")
	}
	if string(key) != "c" {
		t.Fatalf("expected key c, got %s", key)
	}
	if string(val) != "val-c" {
		t.Fatalf("expected value val-c, got %s", val)
	}

	it.Seek([]byte("i"))
	key, val, valid = it.Current()
	if !valid {
		t.Fatalf("valid false, expected true")
	}
	if string(key) != "i" {
		t.Fatalf("expected key i, got %s", key)
	}
	if string(val) != "val-i" {
		t.Fatalf("expected value val-i, got %s", val)
	}

	err = it.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func CommonTestReaderIsolation(t *testing.T, s store.KVStore) {
	// insert a kv pair
	writer, err := s.Writer()
	if err != nil {
		t.Error(err)
	}
	err = writer.Set([]byte("a"), []byte("val-a"))
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	// create an isolated reader
	reader, err := s.Reader()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// verify we see the value already inserted
	val, err := reader.Get([]byte("a"))
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(val, []byte("val-a")) {
		t.Errorf("expected val-a, got nil")
	}

	// verify that an iterator sees it
	count := 0
	it := reader.Iterator([]byte{0})
	defer func() {
		err := it.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	for it.Valid() {
		it.Next()
		count++
	}
	if count != 1 {
		t.Errorf("expected iterator to see 1, saw %d", count)
	}

	// add something after the reader was created
	writer, err = s.Writer()
	if err != nil {
		t.Error(err)
	}
	err = writer.Set([]byte("b"), []byte("val-b"))
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	// ensure that a newer reader sees it
	newReader, err := s.Reader()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := newReader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	val, err = newReader.Get([]byte("b"))
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(val, []byte("val-b")) {
		t.Errorf("expected val-b, got nil")
	}

	// ensure that the director iterator sees it
	count = 0
	it2 := newReader.Iterator([]byte{0})
	defer func() {
		err := it2.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	for it2.Valid() {
		it2.Next()
		count++
	}
	if count != 2 {
		t.Errorf("expected iterator to see 2, saw %d", count)
	}

	// but that the isolated reader does not
	val, err = reader.Get([]byte("b"))
	if err != nil {
		t.Error(err)
	}
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}

	// and ensure that the iterator on the isolated reader also does not
	count = 0
	it3 := reader.Iterator([]byte{0})
	defer func() {
		err := it3.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	for it3.Valid() {
		it3.Next()
		count++
	}
	if count != 1 {
		t.Errorf("expected iterator to see 1, saw %d", count)
	}

}
