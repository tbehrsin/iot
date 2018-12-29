package net

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type EUI64 [8]byte

var NullEUI64 = EUI64{}

func (e EUI64) String() string {
	return strings.ToUpper(hex.EncodeToString(e[:]))
}

func (e EUI64) MarshalV8() interface{} {
	return e.String()
}

func (e *EUI64) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("0x%s", e.String())
	return json.Marshal(s)
}

func (e *EUI64) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if n, err := NewEUI64(strings.TrimPrefix(s, "0x")); err != nil {
		return err
	} else {
		copy(e[:], n[:])
		return nil
	}
}

func NewEUI64(s string) (EUI64, error) {
	if e, err := hex.DecodeString(s); err != nil {
		return NullEUI64, err
	} else if len(e) != 8 {
		return NullEUI64, fmt.Errorf("eui64 not valid")
	} else {
		return EUI64{e[0], e[1], e[2], e[3], e[4], e[5], e[6], e[7]}, nil
	}
}
