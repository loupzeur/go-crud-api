package utils

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

//JSON return a json type
type JSON []byte

//To return JSON to type must be transfered via pointer &val
func (j JSON) To(ret interface{}) error {
	if j.IsNull() {
		return nil
	}
	err := json.Unmarshal(j, ret)
	return err
}

//ToInterface to return an string interface
func (j JSON) ToInterface() map[string]interface{} {
	if j.IsNull() {
		return nil
	}
	var ret map[string]interface{}
	json.Unmarshal(j, &ret)
	return ret
}

//Value to return the current value in string
func (j JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

//New to create from scratch bject
func (j *JSON) New(value interface{}) error {
	jsn, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return j.Scan(jsn)
}

//Scan to scan from sql value
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source")
	}
	*j = append((*j)[0:0], s...)
	return nil
}

//MarshalJSON return byte
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

//UnmarshalJSON return JSON
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("null point exception")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

//IsNull ...
func (j JSON) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

//Equals ...
func (j JSON) Equals(j1 JSON) bool {
	return bytes.Equal([]byte(j), []byte(j1))
}
