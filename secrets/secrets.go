package secrets

import (
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/zalando/go-keyring"
	"os"
	"strings"
)

type SecretService struct {
	serviceName string
}

func New(serviceName string) *SecretService {
	return &SecretService{serviceName}
}

func (s *SecretService) LoadCredentialItem(item string, ptr *string, description string, resetCredentials bool) (err error) {
	if resetCredentials {
		return s.setKeyringItem(item, ptr, description)
	}
	// e.g. "Confluence Username" -> "CONFLUENCE_USERNAME"
	envVar := os.Getenv(strings.Replace(strings.ToUpper(item), " ", "_", -1))
	if envVar != "" {
		*ptr = envVar
		return
	}

	if *ptr != "" && !strings.HasPrefix(*ptr, "<optional-") {
		return nil
	}

	return s.loadKeyringItem(item, ptr, description)
}

func (s *SecretService) loadKeyringItem(item string, ptr *string, description string) (err error) {
	if pw, err := keyring.Get(s.serviceName, item); err == nil {
		*ptr = pw
		return nil
	} else if err == keyring.ErrNotFound {
		return s.setKeyringItem(item, ptr, description)
	} else {
		return err
	}
}

func (s *SecretService) setKeyringItem(item string, ptr *string, description string) (err error) {
	if description != "" {
		_, err := color.New().Add(color.Bold).Fprintln(os.Stderr, description)
		if err != nil {
			return err
		}
	}
	prompt := promptui.Prompt{
		Label: "Enter " + item,
	}
	if strings.HasSuffix(strings.ToLower(item), "password") {
		prompt.Mask = '*'
	}
	result, err := prompt.Run()
	if err != nil {
		return err
	}
	err = keyring.Set(s.serviceName, item, string(result))
	if err != nil {
		return err
	}
	return s.loadKeyringItem(item, ptr, description)
}
