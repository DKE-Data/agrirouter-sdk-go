package agrirouter_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The protocol behavior itself is tested in internal/agmasync. These tests
// cover the routing: that the public operations reach agrirouter over the
// client's own connection with the right endpoint, tenant, and body, and that
// results and errors come back out.

// fakeResponse is what the fake agrirouter answers every request with.
type fakeResponse struct {
	status      int
	body        string
	contentType string
}

// recorded is the last request the fake agrirouter received.
type recorded struct {
	method  string
	path    string
	headers http.Header
	body    string
}

func newMasterdataClient(t *testing.T, res fakeResponse, opts ...agrirouter.ClientOption) (*agrirouter.Client, *recorded) {
	t.Helper()
	rec := &recorded{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*rec = recorded{method: r.Method, path: r.URL.Path, headers: r.Header.Clone(), body: string(body)}
		contentType := res.contentType
		if contentType == "" {
			contentType = "application/json"
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(res.status)
		_, _ = w.Write([]byte(res.body))
	}))
	t.Cleanup(srv.Close)
	client, err := agrirouter.NewClient(srv.URL, opts...)
	require.NoError(t, err)
	return client, rec
}

var (
	endpointID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tenantID   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	objectID   = uuid.MustParse("33333333-3333-3333-3333-333333333333")
)

func TestPutFarmSendsMergePatchAsEndpoint(t *testing.T) {
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusOK,
		body: `{"type":"farm","agrirouter_id":"33333333-3333-3333-3333-333333333333","local_id":"farm-1",` +
			`"name":"Hof","revision":4}`})

	// A merge patch: owner removed by null, every other optional attribute left out.
	base := 3
	farm, err := client.PutFarm(context.Background(), "farm-1", endpointID, tenantID, &base,
		&agrirouter.Farm{Name: "Hof", Owner: nullable.NewNullNullable[agrirouter.EntityReference]()})
	require.NoError(t, err)
	assert.Equal(t, "Hof", farm.Name)
	require.NotNil(t, farm.Revision)
	assert.Equal(t, 4, *farm.Revision)

	assert.Equal(t, http.MethodPut, rec.method)
	assert.Equal(t, "/masterdata/farms/farm-1", rec.path)
	assert.JSONEq(t, `{"type":"farm","local_id":"farm-1","name":"Hof","owner":null}`, rec.body)
	assert.Equal(t, endpointID.String(), rec.headers.Get("X-Agrirouter-Endpoint-Id"))
	assert.Equal(t, tenantID.String(), rec.headers.Get("X-Agrirouter-Tenant-Id"))
	assert.Equal(t, "3", rec.headers.Get("X-Agrirouter-Base-Revision"))
}

func TestDeactivateEntityDecodesTheObject(t *testing.T) {
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusOK,
		body: `{"type":"field","name":"Acker","active":false}`})

	base := 2
	obj, err := client.DeactivateEntity(context.Background(), agrirouter.EntityTypeField, "field-1", endpointID, tenantID, &base)
	require.NoError(t, err)
	assert.Equal(t, agrirouter.EntityTypeField, obj.Envelope.Type)
	require.NotNil(t, obj.Field)
	require.NotNil(t, obj.Field.Active)
	assert.False(t, *obj.Field.Active)
	assert.Equal(t, "/masterdata/fields/field-1/deactivation", rec.path)
}

func TestInitialLoadAddressesEndpointByExternalID(t *testing.T) {
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusOK,
		body: `{"state":"RECONCILING","awaiting_user":true}`})

	status, err := client.ReportUserAttention(context.Background(), "ext-1", tenantID)
	require.NoError(t, err)
	assert.Equal(t, agrirouter.InitialLoadStateReconciling, status.State)
	assert.Equal(t, "/endpoints/ext-1/masterdata-initial-load/user-attention", rec.path)
	assert.Equal(t, tenantID.String(), rec.headers.Get("X-Agrirouter-Tenant-Id"))
}

func TestMasterdataErrors(t *testing.T) {
	ctx := context.Background()
	put := func(c *agrirouter.Client) error {
		_, err := c.PutFarm(ctx, "farm-1", endpointID, tenantID, nil, &agrirouter.Farm{Name: "Hof"})
		return err
	}

	t.Run("412 carries the current revision", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusPreconditionFailed,
			body: `{"message":"stale","current_revision":7}`})
		err := put(client)
		assert.ErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)
		assert.ErrorIs(t, err, agrirouter.ErrRevisionConflict)
		var conflict *agrirouter.RevisionConflict
		require.True(t, errors.As(err, &conflict))
		assert.Equal(t, 7, conflict.CurrentRevision)
	})

	t.Run("409 on a binding carries the rejection", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusConflict,
			body: `{"message":"taken","rejection":{"local_id":"b-1",` +
				`"agrirouter_id":"33333333-3333-3333-3333-333333333333","reason":"LOCAL_ID_ALREADY_BOUND"}}`})
		err := client.BindMapping(ctx, agrirouter.EntityTypeFieldBoundary, "b-1", objectID, endpointID, tenantID)
		assert.ErrorIs(t, err, agrirouter.ErrMappingConflict)
		var conflict *agrirouter.MappingConflict
		require.True(t, errors.As(err, &conflict))
		assert.Equal(t, agrirouter.ReasonLocalIDAlreadyBound, conflict.Rejection.Reason)
		assert.True(t, agrirouter.NeedsUser(conflict.Rejection))
	})

	t.Run("409 on initial load is not a mapping conflict", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusConflict, body: `{"message":"out of order"}`})
		_, err := client.SetInitialLoadState(ctx, "ext-1", tenantID, agrirouter.InitialLoadStateUpdate{})
		assert.ErrorIs(t, err, agrirouter.ErrInitialLoadConflict)
		assert.NotErrorIs(t, err, agrirouter.ErrMappingConflict)
	})

	t.Run("400", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusBadRequest, body: `{"message":"name is required"}`})
		err := put(client)
		assert.ErrorIs(t, err, agrirouter.ErrMasterdataValidation)
		var apiErr *agrirouter.MasterdataAPIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, "name is required", apiErr.Message)
	})
}

func TestStreamMasterdataEventsDecodesFrames(t *testing.T) {
	frames := "id: pos-1\nevent: MASTERDATA_CHANGED\n" +
		`data: {"type":"farm","agrirouter_id":"33333333-3333-3333-3333-333333333333","name":"Hof","revision":2}` + "\n\n" +
		"id: pos-1\nevent: CAUGHT_UP\ndata: {}\n\n"
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frames, contentType: "text/event-stream"})

	var got []agrirouter.MasterdataEvent
	err := client.StreamMasterdataEvents(context.Background(), "pos-0", func(_ context.Context, ev *agrirouter.MasterdataEvent) {
		got = append(got, *ev)
	})
	// The live stream never ends in an orderly way: its end is reported, so the
	// participant reconnects rather than silently stops receiving changes.
	require.ErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
	assert.ErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)

	assert.Equal(t, "/masterdata/events", rec.path)
	assert.Equal(t, "pos-0", rec.headers.Get("Last-Event-ID"))
	require.Len(t, got, 2)
	assert.Equal(t, agrirouter.MasterdataEventChanged, got[0].Type)
	assert.Equal(t, "pos-1", got[0].ID)
	assert.Equal(t, agrirouter.EntityTypeFarm, got[0].Envelope.Type)
	require.NotNil(t, got[0].Farm)
	assert.Equal(t, "Hof", got[0].Farm.Name)
	assert.Equal(t, agrirouter.MasterdataEventCaughtUp, got[1].Type)
	assert.False(t, got[1].HasEntity())
}

func TestStreamMasterdataEventsReadsFramesLargerThan64KB(t *testing.T) {
	// A field boundary's geometry alone can exceed go-sse's 64KB default; a frame
	// over the limit would fail every resume from before it.
	name := strings.Repeat("x", 1<<20)
	frames := "id: pos-1\nevent: MASTERDATA_CHANGED\n" +
		`data: {"type":"farm","agrirouter_id":"33333333-3333-3333-3333-333333333333","name":"` + name + `","revision":2}` + "\n\n"
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frames, contentType: "text/event-stream"})

	var got []agrirouter.MasterdataEvent
	err := client.StreamMasterdataEvents(context.Background(), "", func(_ context.Context, ev *agrirouter.MasterdataEvent) {
		got = append(got, *ev)
	})
	require.ErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].Farm)
	assert.Len(t, got[0].Farm.Name, len(name))
}

func TestWithMasterdataMaxEventSizeBoundsFrames(t *testing.T) {
	frame := "id: pos-1\nevent: MASTERDATA_CHANGED\n" +
		`data: {"type":"farm","agrirouter_id":"33333333-3333-3333-3333-333333333333",` +
		`"name":"` + strings.Repeat("x", 2048) + `","revision":2}` + "\n\n"
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frame, contentType: "text/event-stream"},
		agrirouter.WithMasterdataMaxEventSize(1024))

	err := client.StreamMasterdataEvents(context.Background(), "", func(context.Context, *agrirouter.MasterdataEvent) {
		t.Fatal("a frame over the configured bound must not be delivered")
	})
	require.ErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)
	assert.NotErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
}

func TestWithMasterdataMaxEventSizeRejectsNonPositive(t *testing.T) {
	_, err := agrirouter.NewClient("http://localhost", agrirouter.WithMasterdataMaxEventSize(0))
	require.Error(t, err)
}

func TestStreamMasterdataEventsReturnsContextErrorOnCancel(t *testing.T) {
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: "", contentType: "text/event-stream"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.StreamMasterdataEvents(ctx, "", func(context.Context, *agrirouter.MasterdataEvent) {})
	require.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
}

func TestStreamInitialLoadEventsEndsWithoutError(t *testing.T) {
	frames := "event: MASTERDATA_CHANGED\n" +
		`data: {"type":"farm","agrirouter_id":"33333333-3333-3333-3333-333333333333","name":"Hof","revision":2}` + "\n\n"
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frames, contentType: "text/event-stream"})

	var got []agrirouter.MasterdataEvent
	err := client.StreamInitialLoadEvents(context.Background(), "ext-1", tenantID, func(_ context.Context, ev *agrirouter.MasterdataEvent) {
		got = append(got, *ev)
	})
	require.NoError(t, err)
	require.Len(t, got, 1)
}

func TestMasterdataUsesTheClientsRequestEditors(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)
	client, err := agrirouter.NewClient(srv.URL, agrirouter.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer token")
		return nil
	}))
	require.NoError(t, err)

	err = client.RequestEntity(context.Background(), agrirouter.EntityTypeFarm, endpointID, tenantID, objectID)
	require.NoError(t, err)
	assert.Equal(t, "Bearer token", auth)
}
