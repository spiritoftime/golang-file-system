package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNewEncryptionKey(t *testing.T) {
	for i := 0; i < 10; i++ {

		key := newEncryptionKey()
		fmt.Println(key)
		for i := 0; i < len(key); i++ {
			if key[i] == 0x0 {
				t.Errorf("0 bytes")
			}
		}
	}
}

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo not Bar"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(dst.String())
	//  decrypt
	out := new(bytes.Buffer)
	// decrypt the key in destination, and copy to the buffer
	if _, err := copyDecrypt(key, dst, out); err != nil {
		t.Error(err)

	}
	if out.String() != payload {
		t.Errorf("decryption failed!!!")
	}
}
