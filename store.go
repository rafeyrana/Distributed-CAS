package main

import (
	"crypto/sha1"
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"strings"
	"os"
	"bytes"
)


func CASPathTransformFunc(key string) string{
	// here we will be implementing our key hashing
	// decisions to be made here : SHA1, MD5 or SHA256 
	// we are using SHA1
	hash  := sha1.Sum([]byte(key))
	hashString := hex.EncodeToString(hash[:]) // trick to convert to slice [:]
	blockSize := 5 // depth for block
	sliceLength := len(hashString) / blockSize

	paths := make([]string, sliceLength)
	for i := 0; i < sliceLength; i++ {
		from, to := i*blockSize, (i+1)*blockSize
        paths[i] = hashString[from:to]
    }
	return strings.Join(paths, "/")
}

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



	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	filenameBytes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(filenameBytes[:])

	pathAndFilename := pathName + "/" + filename
	f, err := os.Create(pathAndFilename)
	if err!= nil {
        return err
    }

	n , err := io.Copy(f, buf)
	if err!= nil {
        return err
    }

	log.Printf("written (%d) bytes to disk: %s", n, pathAndFilename)


	


	return nil

}