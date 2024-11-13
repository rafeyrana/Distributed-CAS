package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// getDefaultRootFolder returns the default storage directory.
func getDefaultRootFolder() (string, error) {
	defaultRootFolderName := "Storage"
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return fmt.Sprintf("%s/%s", dir, defaultRootFolderName), nil
}

// CASPathTransformFunc transforms a key into a PathKey using SHA1 hashing.
func CASPathTransformFunc(key string) PathKey {
	// Implementing key transformation using SHA1.
	hash := sha1.Sum([]byte(key))
	hashString := hex.EncodeToString(hash[:]) // Convert to slice [:]
	blockSize := 5                            // Depth for block
	sliceLength := len(hashString) / blockSize

	paths := make([]string, sliceLength)
	for i := 0; i < sliceLength; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashString[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashString,
	}
}

// PathKey represents the transformed path and file name.
type PathKey struct {
	PathName string
	FileName string
}

// FirstPathName returns the first segment of the path name.
func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// PathTransformFunc defines a function type for transforming keys into PathKeys.
type PathTransformFunc func(key string) PathKey

// FullPath constructs the full path for the file.
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}

// StoreOpts contains options for creating a Store.
type StoreOpts struct {
	Root             string           // Root directory for storing files.
	PathTransformFunc PathTransformFunc // Function to transform keys into PathKeys.
}

// DefaultPathTransformFunc is the default transformation function.
var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

// Store is responsible for managing file storage.
type Store struct {
	StoreOpts
}

// NewStore creates a new Store instance with the provided options.
func NewStore(storeOpts StoreOpts) *Store {
	if storeOpts.PathTransformFunc == nil {
		storeOpts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(storeOpts.Root) == 0 {
		var err error
		storeOpts.Root, err = getDefaultRootFolder()
		if err != nil {
			return nil // Handle error appropriately
		}
	}
	return &Store{
		StoreOpts: storeOpts,
	}
}

// HasKey checks if a key exists in the store.
func (s *Store) HasKey(key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPath := pathKey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, fullPath)
	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

// Clear removes all files in the store's root directory.
func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

// Delete removes the file associated with the given key.
func (s *Store) Delete(key string) error {
	path := s.PathTransformFunc(key)
	fullPath := path.FullPath()
	firstPath := path.FirstPathName()
	firstPathWithRoot := fmt.Sprintf("%s/%s", s.Root, firstPath)

	defer func() {
		fmt.Printf("Deleting: %s\n", fullPath)
	}()
	return os.RemoveAll(firstPathWithRoot)
}

// Read retrieves the file associated with the given key.
func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
}

// readStream opens a stream for reading the file associated with the key.
func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	pathAndFilename := pathKey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathAndFilename)

	fi, err := os.Stat(fullPathWithRoot)
	if err != nil {
		return 0, nil, fmt.Errorf("error stating file %s: %w", fullPathWithRoot, err)
	}

	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, fmt.Errorf("error opening file %s: %w", fullPathWithRoot, err)
	}
	return fi.Size(), file, nil
}

// Write stores the content from the provided reader associated with the key.
func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

// WriteDecrypt stores the content from the provided reader associated with the key after decryption.
func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader) (int64, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, fmt.Errorf("error creating directory %s: %w", pathNameWithRoot, err)
	}

	pathAndFilename := pathKey.FullPath()
	fullPathAndFilenameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathAndFilename)
	f, err := os.Create(fullPathAndFilenameWithRoot)
	if err != nil {
		return 0, fmt.Errorf("error creating file %s: %w", fullPathAndFilenameWithRoot, err)
	}

	n, err := copyDecrypt(encKey, r, f)
	if err != nil {
		return 0, fmt.Errorf("error copying decrypted content to file: %w", err)
	}

	fmt.Printf("Written (%d) bytes to disk: %s\n", n, fullPathAndFilenameWithRoot)
	return int64(n), nil
}

// writeStream writes the content from the provided reader associated with the key.
func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, fmt.Errorf("error creating directory %s: %w", pathNameWithRoot, err)
	}

	pathAndFilename := pathKey.FullPath()
	fullPathAndFilenameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathAndFilename)
	f, err := os.Create(fullPathAndFilenameWithRoot)
	if err != nil {
		return 0, fmt.Errorf("error creating file %s: %w", fullPathAndFilenameWithRoot, err)
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return 0, fmt.Errorf("error writing to file %s: %w", fullPathAndFilenameWithRoot, err)
	}

	return n, nil
}