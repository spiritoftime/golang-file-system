package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	expectedFileName := "6804429f74181a63c50c3d81d733a12f14a353ff"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.Filename != expectedFileName {
		t.Errorf("have %s want %s", pathKey.Filename, expectedFileName)
	}

}

// func TestStoreDeleteKey(t *testing.T) {
// 	opts := StoreOpts{
// 		PathTransformFunc: CASPathTransformFunc,
// 	}
// 	s := NewStore(opts)
// 	key := "momsspecials"
// 	data := []byte("some jpg bytes")

// 	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
// 		t.Error(err)
// 	}
// 	if err := s.Delete(key); err != nil {
// 		t.Error(err)
// 	}
// }

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)
	//  u need loop because by default the test will cache the result. so if u tear down with a hardcoded key, it passes first time, if use 2nd time with same key the result got cached so not a good test
	for i := 0; i < 50; i++ {

		key := fmt.Sprintf("foo_%d", i)

		data := []byte("some jpg bytes")

		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s", key)
		}
		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, _ := io.ReadAll(r)

		fmt.Println(string(b))
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}
		fmt.Println(string(b))
		if err := s.Delete(key); err != nil {
			t.Error(err)
		}
		if ok := s.Has(key); ok {
			t.Errorf("expected to NOT have key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
