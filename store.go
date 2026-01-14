package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

// content-addressable storage (CAS) path
// Why do this?
// This prevents having too many files in a single directory. For example, instead of storing a million files in one folder, they're distributed across nested subdirectories based on their hash, improving filesystem performance and organization. The hash ensures the same key always maps to the same path.
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key)) // returns [20] byte => []byte => [:] to converts 20 byte hash to 40 char hexadecimal string (each byte becomes 2 hex chars)
	hashStr := hex.EncodeToString(hash[:])
	// Divides the 40-character hash into 8 blocks of 5 characters each.
	blocksize := 5
	sliceLen := len(hashStr) / blocksize

	paths := make([]string, sliceLen)
	// Extracts each 5-character segment and stores it in the paths slice.

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"), //Combines the segments into a path like 2aae6/c35c9/4fcfb/415db/e95f4/08b78/3ff89/18d27.,
		Filename: hashStr,
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 { // not a directory
		return ""
	}
	return paths[0]
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
	// Root is the folder name of the root, containing all the folders/files of the system.
	Root string
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 { // no root specified
		opts.Root = defaultRootFolderName
	}

	return &Store{StoreOpts: opts}
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)

	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.Filename)
	}()
	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())
	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader) (int64, error) {
	// filesize and error
	f, err := s.openFileForWriting(key)
	if err != nil {
		return 0, err
	}
	n, err := copyDecrypt(encKey, r, f)

	return int64(n), err

}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(key)
	if err != nil {
		return 0, err
	}
	defer f.Close() // --------- ADDED

	return io.Copy(f, r)
	// filesize and error

}

func (s *Store) Read(key string) (int64, io.Reader, error) {

	return s.readStream(key)

}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	file, err := os.Open(pathKeyWithRoot)
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), file, nil
}

func (s *Store) openFileForWriting(key string) (*os.File, error) {
	// anything that calls this should defer file.close!

	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullpathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Create(fullpathWithRoot)

}
