package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"errors"
	"strings"
)

func getDefaultRootFolder() (string, error) {
	defaultRootFolderName := "Storage"
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", dir, defaultRootFolderName), nil
}

func CASPathTransformFunc(key string) PathKey{
	// here we will be implementing our key ing
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

	return PathKey{
		PathName:  strings.Join(paths, "/"),
        FileName:  hashString,
	}
	
}
type PathKey struct {
	PathName string
	FileName string

}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

type PathTransformFunc func(key string) PathKey


func(p PathKey) FullPath() string { 
	return	fmt.Sprintf("%s/%s",p.PathName, p.FileName)
}

type StoreOpts struct {
	Root string // root is the name of the folder directory containing all the files and folders in the CAS
	PathTransformFunc PathTransformFunc
}


var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}
type Store struct {
	StoreOpts
}

func NewStore(storeOpts StoreOpts) *Store {
	if storeOpts.PathTransformFunc == nil {
		storeOpts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(storeOpts.Root) == 0 {
		var err error
		storeOpts.Root, err = getDefaultRootFolder()
		if err != nil {
			// Handle the error appropriately, e.g., log it or return an error
			return nil
		}
	}

	return &Store{
		StoreOpts: storeOpts,
	}
}

func (s *Store) HasKey(key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPath := pathKey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, fullPath)
	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}



func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
	
}

func (s *Store) Delete(key string) error {
	path := s.PathTransformFunc(key)
	fullPath := path.FullPath()
	firstPath := path.FirstPathName()
	firstPathWithRoot := fmt.Sprintf("%s/%s", s.Root, firstPath)
	defer func(){
		fmt.Printf("deleting: %s", fullPath)
	}()
	return os.RemoveAll(firstPathWithRoot)
}


func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	pathAndFilename := pathKey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathAndFilename)
	fi ,err := os.Stat(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}
	file, err := os.Open(fullPathWithRoot)
	if err!= nil {
        return 0, nil, err
    }
	return fi.Size(), file, nil
	
}



func (s *Store) Write(key string, r io.Reader) (int64, error) {// exposing the write steam
	
	return s.writeStream(key, r)
}

func (s *Store) writeStream(key string, r io.Reader)  (int64, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, err
		}

	pathAndFilename := pathKey.FullPath()
	fullPathAndFilenameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathAndFilename)
	f, err := os.Create(fullPathAndFilenameWithRoot)
	if err!= nil {
        return  0, err
    }

	n , err := io.Copy(f, r)
	if err!= nil {
        return  0, err
    }

	fmt.Printf("written (%d) bytes to disk: %s", n, fullPathAndFilenameWithRoot)


	return  n, err

}