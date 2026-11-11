package agrirouter_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		body: `{"type":"farm","revision":4,"content":{"id":{"referenceId":"farm-1","uniqueIds":[` +
			`{"idText":"33333333-3333-3333-3333-333333333333","idTypeCode":"UUID","idSource":"https://agrirouter.com"},` +
			`{"idText":"farm-1","idTypeCode":"STRING"}]},"name":"Hof"}}`})

	// A merge patch: grower removed by null, every other optional attribute left out.
	base := 3
	farm, err := client.PutFarm(context.Background(), "farm-1", endpointID, tenantID, true, &base,
		&agrirouter.Farm{Content: agrirouter.FarmContent{
			Name: "Hof", UnderscoreAgmasyncGrowerRef: nullable.NewNullNullable[agrirouter.Ref](),
		}})
	require.NoError(t, err)
	assert.Equal(t, "Hof", farm.Content.Name)
	require.NotNil(t, farm.Revision)
	assert.Equal(t, 4, *farm.Revision)

	assert.Equal(t, http.MethodPut, rec.method)
	assert.Equal(t, "/masterdata/farms/farm-1", rec.path)
	assert.JSONEq(t, `{"type":"farm","active":true,`+
		`"content":{"id":{"referenceId":"farm-1"},"name":"Hof","_agmasync_growerRef":null}}`, rec.body)
	assert.Equal(t, endpointID.String(), rec.headers.Get("X-Agrirouter-Endpoint-Id"))
	assert.Equal(t, tenantID.String(), rec.headers.Get("X-Agrirouter-Tenant-Id"))
	assert.Equal(t, "3", rec.headers.Get("X-Agrirouter-Base-Revision"))
}

func TestDeactivateEntityDecodesTheObject(t *testing.T) {
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusOK,
		body: `{"type":"field","content":{"id":{"referenceId":"field-1"},"name":"Acker"},"active":false}`})

	base := 2
	obj, err := client.DeactivateEntity(context.Background(), agrirouter.EntityTypeField, "field-1", endpointID, tenantID, &base)
	require.NoError(t, err)
	assert.Equal(t, agrirouter.EntityTypeField, obj.Envelope.Type)
	require.NotNil(t, obj.Field)
	assert.False(t, obj.Field.Active)
	assert.Equal(t, "/masterdata/fields/field-1/deactivation", rec.path)
}

func TestDeactivateEntityDecodesByTheRequestedType(t *testing.T) {
	t.Run("body without type", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK,
			body: `{"content":{"id":{"referenceId":"field-1"},"name":"Acker"},"active":false}`})

		obj, err := client.DeactivateEntity(context.Background(), agrirouter.EntityTypeField, "field-1", endpointID, tenantID, nil)
		require.NoError(t, err)
		assert.Equal(t, agrirouter.EntityTypeField, obj.Envelope.Type)
		require.NotNil(t, obj.Field)
		assert.Equal(t, "Acker", obj.Field.Content.Name)
	})

	t.Run("body of another type", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK,
			body: `{"type":"farm","content":{"id":{"referenceId":"field-1"},"name":"Hof"},"active":false}`})

		obj, err := client.DeactivateEntity(context.Background(), agrirouter.EntityTypeField, "field-1", endpointID, tenantID, nil)
		require.ErrorIs(t, err, agrirouter.ErrEntityTypeMismatch)
		assert.ErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)
		assert.Nil(t, obj)
	})
}

func TestReactivateEntitySendsAPOSTForEveryType(t *testing.T) {
	for _, tc := range []struct {
		entityType agrirouter.EntityType
		collection string
		name       func(*agrirouter.MasterdataObject) string
	}{
		{agrirouter.EntityTypeParty, "parties", func(o *agrirouter.MasterdataObject) string { return o.Party.Content.Name }},
		{agrirouter.EntityTypeFarm, "farms", func(o *agrirouter.MasterdataObject) string { return o.Farm.Content.Name }},
		{agrirouter.EntityTypeField, "fields", func(o *agrirouter.MasterdataObject) string { return o.Field.Content.Name }},
		{agrirouter.EntityTypeFieldBoundary, "field-boundaries",
			func(o *agrirouter.MasterdataObject) string { return o.FieldBoundary.Content.Name.MustGet() }},
	} {
		t.Run(string(tc.entityType), func(t *testing.T) {
			client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusOK,
				body: `{"type":"` + string(tc.entityType) + `","revision":5,"active":true,` +
					`"content":{"id":{"referenceId":"x-1"},"name":"Back"}}`})

			base := 4
			obj, err := client.ReactivateEntity(context.Background(), tc.entityType, "x-1", endpointID, tenantID, &base)
			require.NoError(t, err)

			assert.Equal(t, http.MethodPost, rec.method)
			assert.Equal(t, "/masterdata/"+tc.collection+"/x-1/reactivation", rec.path)
			assert.Empty(t, rec.body)
			assert.Equal(t, endpointID.String(), rec.headers.Get("X-Agrirouter-Endpoint-Id"))
			assert.Equal(t, tenantID.String(), rec.headers.Get("X-Agrirouter-Tenant-Id"))
			assert.Equal(t, "4", rec.headers.Get("X-Agrirouter-Base-Revision"))

			assert.Equal(t, tc.entityType, obj.Envelope.Type)
			assert.True(t, obj.Envelope.Active)
			require.NotNil(t, obj.Envelope.Revision)
			assert.Equal(t, 5, *obj.Envelope.Revision)
			assert.Equal(t, "Back", tc.name(obj))
		})
	}
}

func TestReactivateEntityWithoutABaseSendsNoBaseRevision(t *testing.T) {
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusOK,
		body: `{"type":"farm","active":true,"content":{"id":{"referenceId":"farm-1"},"name":"Hof"}}`})

	_, err := client.ReactivateEntity(context.Background(), agrirouter.EntityTypeFarm, "farm-1", endpointID, tenantID, nil)
	require.NoError(t, err)
	_, sent := rec.headers["X-Agrirouter-Base-Revision"]
	assert.False(t, sent)
}

func TestPutSendsActiveAsGiven(t *testing.T) {
	// The model's own Active is set to the opposite of the argument: the
	// argument is what the write states.
	puts := map[agrirouter.EntityType]func(c *agrirouter.Client, active bool) error{
		agrirouter.EntityTypeParty: func(c *agrirouter.Client, active bool) error {
			_, err := c.PutParty(context.Background(), "x-1", endpointID, tenantID, active, nil,
				&agrirouter.Party{Active: !active, Content: agrirouter.PartyContent{Name: "Hof"}})
			return err
		},
		agrirouter.EntityTypeFarm: func(c *agrirouter.Client, active bool) error {
			_, err := c.PutFarm(context.Background(), "x-1", endpointID, tenantID, active, nil,
				&agrirouter.Farm{Active: !active, Content: agrirouter.FarmContent{Name: "Hof"}})
			return err
		},
		agrirouter.EntityTypeField: func(c *agrirouter.Client, active bool) error {
			_, err := c.PutField(context.Background(), "x-1", endpointID, tenantID, active, nil,
				&agrirouter.Field{Active: !active, Content: agrirouter.FieldContent{Name: "Acker"}})
			return err
		},
		agrirouter.EntityTypeFieldBoundary: func(c *agrirouter.Client, active bool) error {
			_, err := c.PutFieldBoundary(context.Background(), "x-1", endpointID, tenantID, active, nil,
				&agrirouter.FieldBoundary{Active: !active})
			return err
		},
	}
	for entityType, put := range puts {
		for _, active := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s active=%t", entityType, active), func(t *testing.T) {
				client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusCreated,
					body: fmt.Sprintf(`{"type":%q,"active":%t,"content":{"id":{"referenceId":"x-1"}}}`, entityType, active)})

				require.NoError(t, put(client, active))

				var sent struct {
					Active *bool `json:"active"`
				}
				require.NoError(t, json.Unmarshal([]byte(rec.body), &sent))
				require.NotNil(t, sent.Active, "active must be sent on every write")
				assert.Equal(t, active, *sent.Active)
			})
		}
	}
}

func TestMasterdataErrorsCarryTheServersMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("stream refused", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusForbidden,
			body: `{"message":"endpoint not opted into field_boundary"}`})
		err := client.StreamMasterdataEvents(ctx, "", func(context.Context, *agrirouter.MasterdataEvent) error { return nil })
		assert.ErrorIs(t, err, agrirouter.ErrMasterdataForbidden)
		var apiErr *agrirouter.MasterdataAPIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, "endpoint not opted into field_boundary", apiErr.Message)
	})

	t.Run("400 on an operation that declares none", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusBadRequest,
			body: `{"message":"invalid base revision"}`})
		_, err := client.DeactivateEntity(ctx, agrirouter.EntityTypeFarm, "farm-1", endpointID, tenantID, nil)
		assert.ErrorIs(t, err, agrirouter.ErrMasterdataValidation)
		var apiErr *agrirouter.MasterdataAPIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, "invalid base revision", apiErr.Message)
	})
}

func TestStreamRejectsAResponseThatIsNotAnEventStream(t *testing.T) {
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: `{"hello":"proxy"}`})
	err := client.StreamMasterdataEvents(context.Background(), "", func(context.Context, *agrirouter.MasterdataEvent) error {
		t.Fatal("no frame expected")
		return nil
	})
	require.ErrorIs(t, err, agrirouter.ErrNotEventStream)
	assert.ErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)
	assert.NotErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
	assert.ErrorContains(t, err, "application/json")
}

func TestPutWithoutLocalIDIsMatchable(t *testing.T) {
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK})
	farm := &agrirouter.Farm{Content: agrirouter.FarmContent{Name: "Hof"}}
	_, err := client.PutFarm(context.Background(), "", endpointID, tenantID, true, nil, farm)
	assert.ErrorIs(t, err, agrirouter.ErrLocalIDRequired)
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
		_, err := c.PutFarm(ctx, "farm-1", endpointID, tenantID, true, nil, &agrirouter.Farm{Content: agrirouter.FarmContent{Name: "Hof"}})
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

// TestPutConflictsAreToldApartByTheirRequiredField pins the two 409 bodies of a
// put: current_active for a lifecycle conflict, rejection for a mapping one.
func TestPutConflictsAreToldApartByTheirRequiredField(t *testing.T) {
	put := func(c *agrirouter.Client) error {
		_, err := c.PutFarm(context.Background(), "farm-1", endpointID, tenantID, true, nil,
			&agrirouter.Farm{Content: agrirouter.FarmContent{Name: "Hof"}})
		return err
	}

	t.Run("409 on a put with current_active is a lifecycle conflict", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusConflict,
			body: `{"message":"object is inactive; use /reactivation","current_active":false}`})
		err := put(client)
		assert.ErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)
		assert.ErrorIs(t, err, agrirouter.ErrLifecycleConflict)
		assert.NotErrorIs(t, err, agrirouter.ErrMappingConflict)
		var conflict *agrirouter.LifecycleConflict
		require.True(t, errors.As(err, &conflict))
		assert.False(t, conflict.CurrentActive)
		assert.Equal(t, "object is inactive; use /reactivation", conflict.Message)
	})

	t.Run("409 on a put with rejection is a mapping conflict", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusConflict,
			body: `{"message":"taken","rejection":{"local_id":"farm-1",` +
				`"agrirouter_id":"33333333-3333-3333-3333-333333333333","reason":"AGRIROUTER_ID_ALREADY_BOUND"}}`})
		err := put(client)
		assert.ErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)
		assert.ErrorIs(t, err, agrirouter.ErrMappingConflict)
		assert.NotErrorIs(t, err, agrirouter.ErrLifecycleConflict)
		var conflict *agrirouter.MappingConflict
		require.True(t, errors.As(err, &conflict))
		assert.Equal(t, agrirouter.ReasonAgrirouterIDAlreadyBound, conflict.Rejection.Reason)
		assert.Equal(t, "farm-1", conflict.Rejection.LocalId)
	})

	t.Run("409 on a put with neither field is neither conflict", func(t *testing.T) {
		client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusConflict, body: `{"message":"odd"}`})
		err := put(client)
		assert.ErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)
		assert.NotErrorIs(t, err, agrirouter.ErrLifecycleConflict)
		assert.NotErrorIs(t, err, agrirouter.ErrMappingConflict)
		var apiErr *agrirouter.MasterdataAPIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
	})
}

func TestStreamMasterdataEventsDecodesFrames(t *testing.T) {
	frames := "id: pos-1\nevent: MASTERDATA_CHANGED\n" +
		"data: " + farmJSON("33333333-3333-3333-3333-333333333333", "Hof", 2) + "\n\n" +
		"id: pos-1\nevent: CAUGHT_UP\ndata: {}\n\n"
	client, rec := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frames, contentType: "text/event-stream"})

	var got []agrirouter.MasterdataEvent
	err := client.StreamMasterdataEvents(context.Background(), "pos-0", func(_ context.Context, ev *agrirouter.MasterdataEvent) error {
		got = append(got, *ev)
		return nil
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
	assert.Equal(t, "Hof", got[0].Farm.Content.Name)
	assert.Equal(t, agrirouter.MasterdataEventCaughtUp, got[1].Type)
	assert.False(t, got[1].HasEntity())
}

func TestStreamMasterdataEventsDeliversAnInactiveObjectAsChanged(t *testing.T) {
	frames := "id: pos-1\nevent: MASTERDATA_CHANGED\n" +
		`data: {"type":"farm","revision":3,"active":false,"content":{"id":{"uniqueIds":[{"idText":` +
		`"33333333-3333-3333-3333-333333333333","idTypeCode":"UUID","idSource":"https://agrirouter.com"}]},` +
		`"name":"Hof"}}` + "\n\n"
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frames, contentType: "text/event-stream"})

	var got []agrirouter.MasterdataEvent
	err := client.StreamMasterdataEvents(context.Background(), "", func(_ context.Context, ev *agrirouter.MasterdataEvent) error {
		got = append(got, *ev)
		return nil
	})
	require.ErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
	require.Len(t, got, 1)
	assert.Equal(t, agrirouter.MasterdataEventChanged, got[0].Type)
	require.NotNil(t, got[0].Farm)
	assert.False(t, got[0].Farm.Active)
	assert.False(t, got[0].Envelope.Active)
}

func TestStreamMasterdataEventsReadsFramesLargerThan64KB(t *testing.T) {
	// A field boundary's geometry alone can exceed go-sse's 64KB default; a frame
	// over the limit would fail every resume from before it.
	name := strings.Repeat("x", 1<<20)
	frames := "id: pos-1\nevent: MASTERDATA_CHANGED\n" +
		`data: {"type":"farm","revision":2,"content":{"id":{"uniqueIds":[{"idText":` +
		`"33333333-3333-3333-3333-333333333333","idTypeCode":"UUID","idSource":"https://agrirouter.com"}]},` +
		`"name":"` + name + `"}}` + "\n\n"
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frames, contentType: "text/event-stream"})

	var got []agrirouter.MasterdataEvent
	err := client.StreamMasterdataEvents(context.Background(), "", func(_ context.Context, ev *agrirouter.MasterdataEvent) error {
		got = append(got, *ev)
		return nil
	})
	require.ErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].Farm)
	assert.Len(t, got[0].Farm.Content.Name, len(name))
}

var errApply = errors.New("not applied")

// farmJSON is a delivered farm: its canonical identifier, its name, and its revision.
func farmJSON(agrirouterID, name string, revision int) string {
	return fmt.Sprintf(`{"type":"farm","revision":%d,"content":{"id":{"uniqueIds":[`+
		`{"idText":%q,"idTypeCode":"UUID","idSource":"https://agrirouter.com"}]},"name":%q}}`,
		revision, agrirouterID, name)
}

func TestStreamsStopAtTheHandlersError(t *testing.T) {
	frames := "id: pos-1\nevent: MASTERDATA_CHANGED\n" +
		"data: " + farmJSON("33333333-3333-3333-3333-333333333333", "Hof", 2) + "\n\n" +
		"id: pos-2\nevent: MASTERDATA_CHANGED\n" +
		"data: " + farmJSON("44444444-4444-4444-4444-444444444444", "Gut", 1) + "\n\n"
	streams := map[string]func(*agrirouter.Client, agrirouter.MasterdataEventHandler) error{
		"live": func(c *agrirouter.Client, h agrirouter.MasterdataEventHandler) error {
			return c.StreamMasterdataEvents(context.Background(), "", h)
		},
		"initial load": func(c *agrirouter.Client, h agrirouter.MasterdataEventHandler) error {
			return c.StreamInitialLoadEvents(context.Background(), "ext-1", tenantID, h)
		},
	}
	for name, stream := range streams {
		t.Run(name, func(t *testing.T) {
			client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frames, contentType: "text/event-stream"})

			var got []string
			err := stream(client, func(_ context.Context, ev *agrirouter.MasterdataEvent) error {
				got = append(got, ev.Farm.Content.Name)
				return errApply
			})
			// The handler's own error, as is: the participant failed, not the call.
			require.ErrorIs(t, err, errApply)
			assert.NotErrorIs(t, err, agrirouter.ErrMasterdataCallFailed)
			assert.NotErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
			assert.Equal(t, []string{"Hof"}, got, "no frame may follow one that was not applied")
		})
	}
}

func TestDependencyOrderIsACopy(t *testing.T) {
	order := agrirouter.DependencyOrder()
	assert.Equal(t, []agrirouter.EntityType{agrirouter.EntityTypeParty, agrirouter.EntityTypeFarm,
		agrirouter.EntityTypeField, agrirouter.EntityTypeFieldBoundary}, order)
	order[0] = agrirouter.EntityTypeFieldBoundary
	assert.Equal(t, agrirouter.EntityTypeParty, agrirouter.DependencyOrder()[0])
}

func TestWithMasterdataMaxEventSizeBoundsFrames(t *testing.T) {
	frame := "id: pos-1\nevent: MASTERDATA_CHANGED\n" +
		`data: {"type":"farm","revision":2,"content":{"id":{"uniqueIds":[{"idText":` +
		`"33333333-3333-3333-3333-333333333333","idTypeCode":"UUID","idSource":"https://agrirouter.com"}]},` +
		`"name":"` + strings.Repeat("x", 2048) + `"}}` + "\n\n"
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frame, contentType: "text/event-stream"},
		agrirouter.WithMasterdataMaxEventSize(1024))

	err := client.StreamMasterdataEvents(context.Background(), "", func(context.Context, *agrirouter.MasterdataEvent) error {
		t.Fatal("a frame over the configured bound must not be delivered")
		return nil
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

	err := client.StreamMasterdataEvents(ctx, "", func(context.Context, *agrirouter.MasterdataEvent) error { return nil })
	require.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, agrirouter.ErrMasterdataStreamEnded)
}

func TestStreamInitialLoadEventsEndsWithoutError(t *testing.T) {
	frames := "event: MASTERDATA_CHANGED\n" +
		"data: " + farmJSON("33333333-3333-3333-3333-333333333333", "Hof", 2) + "\n\n"
	client, _ := newMasterdataClient(t, fakeResponse{status: http.StatusOK, body: frames, contentType: "text/event-stream"})

	var got []agrirouter.MasterdataEvent
	err := client.StreamInitialLoadEvents(context.Background(), "ext-1", tenantID,
		func(_ context.Context, ev *agrirouter.MasterdataEvent) error {
			got = append(got, *ev)
			return nil
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
