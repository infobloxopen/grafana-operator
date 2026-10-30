package grafanadashboard

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func newFolderTestServer(t *testing.T, listStatus int, listBody string, createCalls *int32) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/folders", r.URL.Path)

		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(listStatus)
			_, _ = w.Write([]byte(listBody))
		case http.MethodPost:
			atomic.AddInt32(createCalls, 1)
			_, _ = w.Write([]byte(`{"id": 99, "title": "created", "uid": "new"}`))
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	}))
}

func TestCreateOrUpdateFolderDoesNotCreateWhenFolderListIsUnavailable(t *testing.T) {
	var createCalls int32
	srv := newFolderTestServer(t, http.StatusServiceUnavailable, ``, &createCalls)
	defer srv.Close()

	client := NewGrafanaClient(srv.URL, "user", "pass", &http.Transport{}, 5)

	_, err := client.CreateOrUpdateFolder("TDI")

	require.Error(t, err)
	require.Contains(t, err.Error(), "503")
	require.Zero(t, atomic.LoadInt32(&createCalls), "must not create a folder when the existing folders could not be listed")
}

func TestCreateOrUpdateFolderReusesExistingFolder(t *testing.T) {
	var createCalls int32
	srv := newFolderTestServer(t, http.StatusOK, `[{"id": 7, "title": "TDI", "uid": "abc"}]`, &createCalls)
	defer srv.Close()

	client := NewGrafanaClient(srv.URL, "user", "pass", &http.Transport{}, 5)

	folder, err := client.CreateOrUpdateFolder("tdi")

	require.NoError(t, err)
	require.NotNil(t, folder.ID)
	require.Equal(t, int64(7), *folder.ID)
	require.Zero(t, atomic.LoadInt32(&createCalls))
}

func TestCreateOrUpdateFolderCreatesMissingFolder(t *testing.T) {
	var createCalls int32
	srv := newFolderTestServer(t, http.StatusOK, `[{"id": 7, "title": "Other", "uid": "abc"}]`, &createCalls)
	defer srv.Close()

	client := NewGrafanaClient(srv.URL, "user", "pass", &http.Transport{}, 5)

	folder, err := client.CreateOrUpdateFolder("TDI")

	require.NoError(t, err)
	require.NotNil(t, folder.ID)
	require.Equal(t, int64(99), *folder.ID)
	require.Equal(t, int32(1), atomic.LoadInt32(&createCalls))
}
