package wallet

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// KeyStore ..
type KeyStore interface {
	// 密钥持久化 ..
	Store(name string, key []byte) error
	// 密钥加载
	Load(name string) ([]byte, error)
}

// FileKeyStore ..
type FileKeyStore struct {
	password string
	baseDir  string
}

// NewFilKeyStore 生成 FileKeyStore 实例
func NewFilKeyStore(baseDir string, password string) *FileKeyStore {
	return &FileKeyStore{
		password: password,
		baseDir:  baseDir,
	}
}

// Store 密钥持久化 ..
func (fk *FileKeyStore) Store(name string, key []byte) error {
	ad, err := filepath.Abs(fk.baseDir)
	if err != nil {
		return err
	}
	fileName := filepath.Join(ad, name+".sec")
	data, err := fk.encrypt(key, fk.password)
	if err != nil {
		return err
	}

	err = os.MkdirAll(ad, 0766)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (fk *FileKeyStore) encrypt(plain []byte, password string) ([]byte, error) {
	if len(password) == 0 {
		return plain, nil
	}
	secData, err := AESEncrypt(plain, []byte(fk.password))
	if err != nil {
		return nil, err
	}
	return secData, nil
}

func (fk *FileKeyStore) decrypt(data []byte, password string) ([]byte, error) {
	if len(password) == 0 {
		return data, nil
	}
	plain, err := AESDecrypt(data, []byte(fk.password))
	if err != nil {
		return nil, err
	}
	return plain, nil
}

// Load 密钥加载
func (fk *FileKeyStore) Load(name string) ([]byte, error) {
	fileName := filepath.Join(fk.baseDir, name+".sec")
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	ori, err := fk.decrypt(data, fk.password)
	if err != nil {
		return nil, err
	}
	return ori, nil
}
