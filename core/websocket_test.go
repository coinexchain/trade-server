package core

import (
	"encoding/json"
	"testing"
	
	"github.com/stretchr/testify/require"
)

func TestGroupOfDataPacket(t *testing.T)  {
	bz := groupOfDataPacket("test", nil)
	require.EqualValues(t, "{\"type\":\"test\", \"payload\":[]}", bz)
	
	data := make([]json.RawMessage,  2)
	data[0] = json.RawMessage("hello")
	data[1] = json.RawMessage("world")
	bz = groupOfDataPacket("test", data)
	require.EqualValues(t, "{\"type\":\"test\", \"payload\":[hello,world]}", bz)
	
}
