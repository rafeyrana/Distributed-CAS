package main

import (
	"bytes"
	"io/ioutil"	
	"fmt"
	"testing"
	"os"
)

func TestPathTransformFuncs(t *testing.T) {
	key := "momsbestpicture"

	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPath := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathKey.FileName != expectedOriginalKey || pathKey.PathName != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, pathKey.PathName)
	}

}




func (s *Store) HasKey(key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPath := pathKey.FullPath()
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}


func TestDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "myspecialpictures"
	data := []byte("some jpeg bytes")
	err := s.writeStream(key, bytes.NewReader(data))
	if err != nil {
		t.Error(err)
	}
	err = s.Delete(key)
	if err != nil {
		t.Error(err)
	}
}
func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "myspecialpictures"
	data := []byte("some jpeg bytes")
	err := s.writeStream(key, bytes.NewReader(data))
	if err != nil {
		t.Error(err)
	}
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := ioutil.ReadAll(r)
	fmt.Println("this is the data:", string(b))
	if string(b) != string(data) {
		t.Errorf("expected data : %s, got data from reading: %s", data, b)
	}
	
}