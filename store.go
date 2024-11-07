package main

import (
	"io"
	"os"
	"log"
)

type PathTransformFunc func(key string) string

type StoreOpts struct {

	PathTransformFunc PathTransformFunc
}




var DefaultPathTransformFunc = func(key string) string {
	return key
}
type Store struct {
	StoreOpts
}

func NewStore(storeOpts StoreOpts) *Store{

	return &Store{
		StoreOpts: storeOpts,
	}
}

func (s *Store) writeStream(key string, r io.Reader)  error {
	pathName := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
		}

	filename := "someFilename"

	pathAndFilename := pathName + "/" + filename
	f, err := os.Create(pathAndFilename)
	if err!= nil {
        return err
    }

	n , err := io.Copy(f, r)
	if err!= nil {
        return err
    }

	log.Printf("written (%d) bytes to disk: %s", n, pathAndFilename)


	


	return nil

}