package server

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinexchain/trade-server/core"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func TestHandlerOK(t *testing.T) {

	db := dbm.NewMemDB()
	subMan := &core.MocSubscribeManager{}
	hub := core.NewHub(db, subMan, 99999, 0, 0, 0, "", 0)
	wsManager := core.NewWebSocketManager()
	// store height
	storeHeightInfo(db, 3, 1990)
	storeHeightInfo(db, 4, 1991)
	storeHeightInfo(db, 5, 1992)

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/misc/height", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler, _ := registerHandler(hub, wsManager, false, "", "", nil)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	require.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what we expect.
	expected := `{"height":5,"timestamp":1992}`
	require.Equal(t, expected, rr.Body.String())
}

func TestHandlerNotFound(t *testing.T) {

	db := dbm.NewMemDB()
	subMan := &core.MocSubscribeManager{}
	hub := core.NewHub(db, subMan, 99999, 0, 0, 0, "", 0)
	wsManager := core.NewWebSocketManager()
	// store height
	storeHeightInfo(db, 3, 1990)
	storeHeightInfo(db, 4, 1991)
	storeHeightInfo(db, 5, 1992)

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/misc/heighttt", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler, _ := registerHandler(hub, wsManager, false, "", "", nil)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	require.Equal(t, http.StatusNotFound, rr.Code)
}

func storeHeightInfo(db dbm.DB, height, timestamp uint64) {
	heightBytes := core.Int64ToBigEndianBytes(int64(height))
	key := append([]byte{core.BlockHeightByte}, heightBytes...)
	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val[:], timestamp)
	db.Set(key, val)
}
