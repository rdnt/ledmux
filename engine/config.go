package engine

import (
	"encoding/json"
	"os"
)

// SaveConfig updates the saved configuration file with in-memory values
func (amb *Engine) SaveConfig() error {
	f, err := os.OpenFile("ambilight.conf", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc, err := json.MarshalIndent(amb, "", "  ")
	if err != nil {
		return err
	}
	err = f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.Write(enc)
	if err != nil {
		return err
	}
	return nil
}
