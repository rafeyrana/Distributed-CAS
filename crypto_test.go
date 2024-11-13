package main

import (
	"fmt"
	"bytes"
	"testing"
)


func TestNewEncryptionKey(t *testing.T) {
	key := newEncryptionKey()
	if len(key) != 32 {
		t.Error("expected 32 bytes")
	}
	
}
func TestCopyEncryptDecrypt(t *testing.T) {
	src := bytes.NewReader([]byte("some bytes"))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()

	_ , err := copyEncrypt(key, src, dst)
	if err != nil{
		t.Error(err)
	}

	fmt.Println(dst.String())
	out := new(bytes.Buffer)
	_, err = copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(out.String())

	


}