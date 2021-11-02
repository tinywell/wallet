package keystore

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"bewallet/pkg/utils"
)

// keystore const
const (
	Tag     = "keystore"
	TagFile = ".tag"
)

// 错误
var (
	ErrPassword error = errors.New("口令错误")
)

// KeyStore ..
type KeyStore interface {
	// 密钥持久化 ..
	Store(opt StoreOpts) error
	// 密钥加载
	Load(opt LoadOpts) ([]byte, error)
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

func (fk *FileKeyStore) encrypt(plain []byte) ([]byte, error) {
	if len(fk.password) == 0 {
		return plain, nil
	}
	secData, err := utils.AESEncrypt(plain, []byte(fk.password))
	if err != nil {
		return nil, err
	}
	return secData, nil
}

func (fk *FileKeyStore) decrypt(data []byte) ([]byte, error) {
	if len(fk.password) == 0 {
		return data, nil
	}
	plain, err := utils.AESDecrypt(data, []byte(fk.password))
	if err != nil {
		return nil, err
	}
	return plain, nil
}

// Store 密钥持久化 ..
func (fk *FileKeyStore) Store(opt StoreOpts) error {
	ad, err := filepath.Abs(fk.baseDir)
	if err != nil {
		return err
	}
	dir := filepath.Join(ad, opt.Identity())
	err = os.MkdirAll(dir, 0766)
	if err != nil {
		return err
	}
	filename := getFileName(opt.StoreType(), opt.Identity())

	filePath := filepath.Join(dir, filename)
	data, err := fk.encrypt(opt.Data())
	if err != nil {
		return err
	}

	err = os.MkdirAll(ad, 0766)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

// Load 密钥加载
func (fk *FileKeyStore) Load(opt LoadOpts) ([]byte, error) {

	fileName := getFileName(opt.LoadType(), opt.Identity())
	filePath := filepath.Join(fk.baseDir, opt.Identity(), fileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if err == os.ErrNotExist && opt.LoadType() == KeyTypeNetwork {
			return nil, nil
		}
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

// Reset 重置钱包存储数据
func Reset(baseDir string) {
	os.RemoveAll(baseDir)
}

func getFileName(keyType, name string) string {
	var fileName string
	switch keyType {
	case KeyTypeSecret:
		fileName = name + ".sec"
	case KeyTypeNetwork:
		fileName = name + ".net"
	}
	return fileName
}
