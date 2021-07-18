package transports

import (
	"errors"
	"io/ioutil"
	"os"
)

func NewPrivateKeyCredential(user string, pemBytes []byte, password string) *Credential {
	return &Credential{
		Type:     CredentialType_private_key,
		User:     user,
		Secret:   pemBytes,
		Password: password,
	}
}

func NewPrivateKeyCredentialFromPath(user string, path string, password string) (*Credential, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("private key does not exist " + path)
	}

	pemBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewPrivateKeyCredential(user, pemBytes, password), nil
}

func NewPasswordCredential(user string, password string) *Credential {
	return &Credential{
		Type:   CredentialType_password,
		User:   user,
		Secret: []byte(password),
	}
}

// GetPassword returns the first password in the list
func GetPassword(list []*Credential) (*Credential, error) {
	for i := range list {
		credential := list[i]
		if credential.Type == CredentialType_password {
			return credential, nil
		}
	}
	return nil, errors.New("no password found")
}