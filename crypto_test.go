package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNewEncryptionKey(t *testing.T) {
	for i := 0; i < 10; i++ {

		key, err := newEncryptionKey()
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		fmt.Printf("key is %s", key)
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
	key, err := newEncryptionKey()
	if err != nil {
		t.Error(err)
	}
	_, err = copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(len(dst.String()))
	//  decrypt
	out := new(bytes.Buffer)
	// decrypt the key in destination, and copy to the buffer
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)

	}
	// iv + payload, because we prepend it
	if nw != 16+len(payload) {
		t.Fail()
	}
	fmt.Println(len(payload))
	if out.String() != payload {
		t.Errorf("decryption failed!!!")
	}
}
