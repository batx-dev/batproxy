package ssh

import "fmt"

type key struct {
	User string
	Host string
}

func (k *key) String() string {
	return fmt.Sprintf("%s@%s", k.User, k.Host)
}
