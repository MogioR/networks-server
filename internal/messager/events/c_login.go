package events

import "bytes"

type LoginEvent struct {
	Login    string
	Password string
}

func (e *LoginEvent) Deserialize(buf *bytes.Buffer) error {
	loginLength, err := buf.ReadByte()
	if err != nil {
		return err
	}

	loginBytes := make([]byte, loginLength)
	if _, err := buf.Read(loginBytes); err != nil {
		return err
	}
	e.Login = string(loginBytes)

	passwordLength, err := buf.ReadByte()
	if err != nil {
		return err
	}

	passwordBytes := make([]byte, passwordLength)
	if _, err := buf.Read(passwordBytes); err != nil {
		return err
	}
	e.Password = string(passwordBytes)

	return nil
}
