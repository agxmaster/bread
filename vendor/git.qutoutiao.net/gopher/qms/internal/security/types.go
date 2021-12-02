// Package security created on 2017/6/22
package security

//Cipher interface declares two function for encryption and decryption
type Cipher interface {
	Encrypt(src string) (string, error)

	Decrypt(src string) (string, error)
}
