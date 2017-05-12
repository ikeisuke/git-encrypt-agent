package install

import (
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/service/kms"

	"github.com/ikeisuke/git-encrypt-agent/config"
)

type DataKey struct {
	request *kms.GenerateDataKeyWithoutPlaintextInput
	kms     *kms.KMS
}

func LoadDataKey(file string) ([]byte, error) {
	data, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return data, err
}

func SaveDataKey(file string, data []byte) error {
	return ioutil.WriteFile(datakeyfile, data.CiphertextBlob, 0644)
}

func NewDataKeyWithConfig(c *config.Config) (*DataKey, error) {
	d := new(DataKey)
	session := c.AWSSession()
	d.kms = kms.New(session)
	return nil, d
}

func (d *DataKey) SetKeyId(keyId string) {
	d.request = kms.GenerateDataKeyWithoutPlaintextInput{}
	d.request.SetKeyId(keyId)
	d.request.SetKeySpec("AES_256")
}

func (d *DataKey) GenerateDataKey() ([]byte, error) {
	data, err := d.kms.GenerateDataKeyWithoutPlaintext(&d.request)
	if err != nil {
		return nil, err
	}
	return data.CiphertextBlob, nil
}
