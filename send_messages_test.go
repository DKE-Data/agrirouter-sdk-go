package agrirouter_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMessagesStreamedBodySendsContentLength(t *testing.T) {
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusAccepted})

	payload := []byte("zipped shapefile bytes")
	// An *os.File is a body http.NewRequest cannot size on its own.
	f, err := os.CreateTemp(t.TempDir(), "payload-*.zip")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	_, err = f.Write(payload)
	require.NoError(t, err)
	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)

	err = client.SendMessages(context.Background(), &agrirouter.SendMessagesParams{
		ContentLength:            int64(len(payload)),
		XAgrirouterIsPublish:     true,
		XAgrirouterSentTimestamp: time.Now(),
		XAgrirouterEndpointId:    endpointID,
		XAgrirouterTenantId:      tenantID,
		XAgrirouterMessageType:   "shp:shape:zip",
	}, f)
	require.NoError(t, err)

	assert.Equal(t, "22", rec.headers.Get("Content-Length"))
	assert.Empty(t, rec.headers.Get("Transfer-Encoding"))
	assert.Equal(t, string(payload), rec.body)
}

func TestSendMessagesKeepsDerivedContentLengthWhenParamUnset(t *testing.T) {
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusAccepted})

	payload := []byte("in-memory payload")
	err := client.SendMessages(context.Background(), &agrirouter.SendMessagesParams{
		XAgrirouterIsPublish:     true,
		XAgrirouterSentTimestamp: time.Now(),
		XAgrirouterEndpointId:    endpointID,
		XAgrirouterTenantId:      tenantID,
		XAgrirouterMessageType:   "shp:shape:zip",
	}, bytes.NewReader(payload))
	require.NoError(t, err)

	assert.Equal(t, "17", rec.headers.Get("Content-Length"))
	assert.Empty(t, rec.headers.Get("Transfer-Encoding"))
	assert.Equal(t, string(payload), rec.body)
}
