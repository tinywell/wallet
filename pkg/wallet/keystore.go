package wallet

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// keystore const
const (
	Tag     = "keystore"
	TagFile = ".tag"
)

var (
	ErrPassword error = errors.New("口令错误")
)

// KeyStore ..
type KeyStore interface {
	// 密钥持久化 ..
	Store(name string, key []byte) error
	// 密钥加载
	Load(name string) ([]byte, error)
	// 列出密钥列表
	List() ([]string, error)
}

// FileKeyStore ..
type FileKeyStore struct {
	password string
	baseDir  string
}

// NewFilKeyStore 生成 FileKeyStore 实例
func NewFilKeyStore(baseDir string, password string) (*FileKeyStore, error) {
	fk := &FileKeyStore{
		password: password,
		baseDir:  baseDir,
	}
	if fk.isNew() {
		err := fk.tag()
		if err != nil {
			return nil, err
		}
		return fk, nil
	}
	err := fk.verify()
	if err != nil {
		return nil, err
	}
	return fk, nil
}

func (fk *FileKeyStore) isNew() bool {
	file := filepath.Join(fk.baseDir, TagFile)
	_, err := os.Stat(file)
	if err != nil {
		return true
	}
	return false
}

func (fk *FileKeyStore) tag() error {
	err := os.MkdirAll(fk.baseDir, 0766)
	if err != nil {
		return err
	}
	file := filepath.Join(fk.baseDir, TagFile)
	enctag, err := fk.encrypt([]byte(Tag))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, enctag, 0666)
}

func (fk *FileKeyStore) verify() error {
	file := filepath.Join(fk.baseDir, TagFile)
	enctag, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	tag, err := fk.decrypt(enctag)
	if err != nil {
		return ErrPassword
	}
	if string(tag) == Tag {
		return nil
	}
	return ErrPassword
}

// Store 密钥持久化 ..
func (fk *FileKeyStore) Store(name string, key []byte) error {
	ad, err := filepath.Abs(fk.baseDir)
	if err != nil {
		return err
	}
	dir := filepath.Join(ad, name)
	err = os.MkdirAll(dir, 0766)
	if err != nil {
		return err
	}
	fileName := filepath.Join(dir, name+".sec")
	data, err := fk.encrypt(key)
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

func (fk *FileKeyStore) encrypt(plain []byte) ([]byte, error) {
	if len(fk.password) == 0 {
		return plain, nil
	}
	secData, err := AESEncrypt(plain, []byte(fk.password))
	if err != nil {
		return nil, err
	}
	return secData, nil
}

func (fk *FileKeyStore) decrypt(data []byte) ([]byte, error) {
	if len(fk.password) == 0 {
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
	fileName := filepath.Join(fk.baseDir, name, name+".sec")
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	ori, err := fk.decrypt(data)
	if err != nil {
		return nil, err
	}
	return ori, nil
}

// List 返回已存储的密钥列表
func (fk *FileKeyStore) List() ([]string, error) {
	list := []string{}
	f, err := os.Open(fk.baseDir)
	if err != nil {
		return nil, err
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		list = append(list, f.Name())
	}
	return list, nil
}
