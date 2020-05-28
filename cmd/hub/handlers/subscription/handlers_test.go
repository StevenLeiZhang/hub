package subscription

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/artifacthub/hub/cmd/hub/handlers/helpers"
	"github.com/artifacthub/hub/internal/hub"
	"github.com/artifacthub/hub/internal/subscription"
	"github.com/artifacthub/hub/internal/tests"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}

func TestAdd(t *testing.T) {
	t.Run("invalid subscription provided", func(t *testing.T) {
		testCases := []struct {
			description      string
			subscriptionJSON string
			smErr            error
		}{
			{
				"no subscription provided",
				"",
				nil,
			},
			{
				"invalid json",
				"-",
				nil,
			},
			{
				"invalid package id",
				`{"package_id": "invalid"}`,
				subscription.ErrInvalidInput,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.description, func(t *testing.T) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("POST", "/", strings.NewReader(tc.subscriptionJSON))
				r = r.WithContext(context.WithValue(r.Context(), hub.UserIDKey, "userID"))

				hw := newHandlersWrapper()
				if tc.smErr != nil {
					hw.sm.On("Add", r.Context(), mock.Anything).Return(tc.smErr)
				}
				hw.h.Add(w, r)
				resp := w.Result()
				defer resp.Body.Close()

				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				hw.sm.AssertExpectations(t)
			})
		}
	})

	t.Run("valid subscription provided", func(t *testing.T) {
		subscriptionJSON := `
		{
			"package_id": "00000000-0000-0000-0000-000000000001",
			"event_kind": 0
		}
		`
		s := &hub.Subscription{}
		_ = json.Unmarshal([]byte(subscriptionJSON), &s)

		testCases := []struct {
			description        string
			err                error
			expectedStatusCode int
		}{
			{
				"add subscription succeeded",
				nil,
				http.StatusOK,
			},
			{
				"error adding subscription",
				tests.ErrFakeDatabaseFailure,
				http.StatusInternalServerError,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.description, func(t *testing.T) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("POST", "/", strings.NewReader(subscriptionJSON))
				r = r.WithContext(context.WithValue(r.Context(), hub.UserIDKey, "userID"))

				hw := newHandlersWrapper()
				hw.sm.On("Add", r.Context(), s).Return(tc.err)
				hw.h.Add(w, r)
				resp := w.Result()
				defer resp.Body.Close()

				assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
				hw.sm.AssertExpectations(t)
			})
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("invalid subscription provided", func(t *testing.T) {
		testCases := []struct {
			description      string
			subscriptionJSON string
			smErr            error
		}{
			{
				"no subscription provided",
				"",
				nil,
			},
			{
				"invalid json",
				"-",
				nil,
			},
			{
				"invalid package id",
				`{"package_id": "invalid"}`,
				subscription.ErrInvalidInput,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.description, func(t *testing.T) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("DELETE", "/", strings.NewReader(tc.subscriptionJSON))
				r = r.WithContext(context.WithValue(r.Context(), hub.UserIDKey, "userID"))

				hw := newHandlersWrapper()
				if tc.smErr != nil {
					hw.sm.On("Delete", r.Context(), mock.Anything).Return(tc.smErr)
				}
				hw.h.Delete(w, r)
				resp := w.Result()
				defer resp.Body.Close()

				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				hw.sm.AssertExpectations(t)
			})
		}
	})

	t.Run("valid subscription provided", func(t *testing.T) {
		subscriptionJSON := `
		{
			"package_id": "00000000-0000-0000-0000-000000000001",
			"event_kind": 0
		}
		`
		s := &hub.Subscription{}
		_ = json.Unmarshal([]byte(subscriptionJSON), &s)

		testCases := []struct {
			description        string
			err                error
			expectedStatusCode int
		}{
			{
				"delete subscription succeeded",
				nil,
				http.StatusOK,
			},
			{
				"error deleting subscription",
				tests.ErrFakeDatabaseFailure,
				http.StatusInternalServerError,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.description, func(t *testing.T) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("DELETE", "/", strings.NewReader(subscriptionJSON))
				r = r.WithContext(context.WithValue(r.Context(), hub.UserIDKey, "userID"))

				hw := newHandlersWrapper()
				hw.sm.On("Delete", r.Context(), s).Return(tc.err)
				hw.h.Delete(w, r)
				resp := w.Result()
				defer resp.Body.Close()

				assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
				hw.sm.AssertExpectations(t)
			})
		}
	})
}

func TestGetByPackage(t *testing.T) {
	rctx := &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"packageID"},
			Values: []string{"packageID"},
		},
	}

	t.Run("error getting package subscriptions", func(t *testing.T) {
		testCases := []struct {
			smErr              error
			expectedStatusCode int
		}{
			{
				subscription.ErrInvalidInput,
				http.StatusBadRequest,
			},
			{
				tests.ErrFakeDatabaseFailure,
				http.StatusInternalServerError,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.smErr.Error(), func(t *testing.T) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "/", nil)
				r = r.WithContext(context.WithValue(r.Context(), hub.UserIDKey, "userID"))
				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

				hw := newHandlersWrapper()
				hw.sm.On("GetByPackageJSON", r.Context(), "packageID").Return(nil, tc.smErr)
				hw.h.GetByPackage(w, r)
				resp := w.Result()
				defer resp.Body.Close()

				assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
				hw.sm.AssertExpectations(t)
			})
		}
	})

	t.Run("get package subscriptions succeeded", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), hub.UserIDKey, "userID"))
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		hw := newHandlersWrapper()
		hw.sm.On("GetByPackageJSON", r.Context(), "packageID").Return([]byte("dataJSON"), nil)
		hw.h.GetByPackage(w, r)
		resp := w.Result()
		defer resp.Body.Close()
		h := resp.Header
		data, _ := ioutil.ReadAll(resp.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", h.Get("Content-Type"))
		assert.Equal(t, helpers.BuildCacheControlHeader(0), h.Get("Cache-Control"))
		assert.Equal(t, []byte("dataJSON"), data)
		hw.sm.AssertExpectations(t)
	})
}

func TestGetByUser(t *testing.T) {
	t.Run("error getting user subscriptions", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), hub.UserIDKey, "userID"))

		hw := newHandlersWrapper()
		hw.sm.On("GetByUserJSON", r.Context()).Return(nil, tests.ErrFakeDatabaseFailure)
		hw.h.GetByUser(w, r)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		hw.sm.AssertExpectations(t)
	})

	t.Run("get user subscriptions succeeded", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), hub.UserIDKey, "userID"))

		hw := newHandlersWrapper()
		hw.sm.On("GetByUserJSON", r.Context()).Return([]byte("dataJSON"), nil)
		hw.h.GetByUser(w, r)
		resp := w.Result()
		defer resp.Body.Close()
		h := resp.Header
		data, _ := ioutil.ReadAll(resp.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", h.Get("Content-Type"))
		assert.Equal(t, helpers.BuildCacheControlHeader(0), h.Get("Cache-Control"))
		assert.Equal(t, []byte("dataJSON"), data)
		hw.sm.AssertExpectations(t)
	})
}

type handlersWrapper struct {
	sm *subscription.ManagerMock
	h  *Handlers
}

func newHandlersWrapper() *handlersWrapper {
	sm := &subscription.ManagerMock{}

	return &handlersWrapper{
		sm: sm,
		h:  NewHandlers(sm),
	}
}