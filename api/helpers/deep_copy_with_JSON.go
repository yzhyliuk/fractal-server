package helpers

import "encoding/json"

func DeepCopyJSON(source interface{}, destination interface{}) error {

	bytes, err := json.Marshal(source)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, destination)
	if err != nil {
		return err
	}
	return nil
}

