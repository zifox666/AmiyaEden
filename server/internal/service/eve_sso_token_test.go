package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/pkg/eve"
	"context"
	"encoding/base64"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	global.Logger = zap.NewNop()
	os.Exit(m.Run())
}

const testCharacterID int64 = 90000001

type fakeCharacterStore struct {
	char        model.EveCharacter
	markedWith  string
	updateCalls int
}

func (f *fakeCharacterStore) GetByCharacterID(int64) (*model.EveCharacter, error) {
	char := f.char
	return &char, nil
}

func (f *fakeCharacterStore) Update(char *model.EveCharacter) error {
	f.char = *char
	f.updateCalls++
	return nil
}

func (f *fakeCharacterStore) MarkTokenInvalid(_ int64, refreshToken string) error {
	f.markedWith = refreshToken
	f.char.TokenInvalid = true
	return nil
}

type fakeRefresher struct {
	resp  *eve.TokenResponse
	err   error
	calls atomic.Int32
	block chan struct{}
}

func (f *fakeRefresher) RefreshAccessToken(context.Context, string) (*eve.TokenResponse, error) {
	f.calls.Add(1)
	if f.block != nil {
		<-f.block
	}
	return f.resp, f.err
}

func fakeAccessToken() string {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"CHARACTER:EVE:90000001","scp":["esi-a.v1","esi-b.v1"]}`))
	return "header." + payload + ".signature"
}

func expiredStore(invalid bool) *fakeCharacterStore {
	return &fakeCharacterStore{char: model.EveCharacter{
		CharacterID:  testCharacterID,
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		Scopes:       "old-scope",
		TokenExpiry:  time.Now().Add(-time.Minute),
		TokenInvalid: invalid,
	}}
}

func TestRefreshKeepsCharacterValidOnTransientFailure(t *testing.T) {
	failures := map[string]error{
		"网络错误":    errors.New("token request: connection reset by peer"),
		"SSO 502": &eve.TokenError{StatusCode: 502, Body: "Bad Gateway"},
		"限流 429":  &eve.TokenError{StatusCode: 429},
		"客户端配置错误": &eve.TokenError{StatusCode: 401, Code: "invalid_client"},
		"请求取消":    context.Canceled,
	}

	for name, failure := range failures {
		t.Run(name, func(t *testing.T) {
			store := expiredStore(false)
			_, err := refreshCharacterToken(context.Background(), store, &fakeRefresher{err: failure}, testCharacterID)

			if err == nil {
				t.Fatal("want error")
			}
			if store.char.TokenInvalid || store.markedWith != "" {
				t.Error("暂时性失败不应标记角色失效")
			}
			if store.char.RefreshToken != "old-refresh" {
				t.Errorf("RefreshToken = %q, 应保留", store.char.RefreshToken)
			}
		})
	}
}

func TestRefreshKeepsCharacterValidOnUnparsableToken(t *testing.T) {
	store := expiredStore(false)
	refresher := &fakeRefresher{resp: &eve.TokenResponse{AccessToken: "not-a-jwt", RefreshToken: "new"}}

	if _, err := refreshCharacterToken(context.Background(), store, refresher, testCharacterID); err == nil {
		t.Fatal("want error")
	}
	if store.char.TokenInvalid || store.char.RefreshToken != "old-refresh" {
		t.Errorf("char = %+v, 不应被修改", store.char)
	}
}

func TestRefreshMarksInvalidOnlyOnInvalidGrant(t *testing.T) {
	store := expiredStore(false)
	refresher := &fakeRefresher{err: &eve.TokenError{StatusCode: 400, Code: "invalid_grant"}}

	_, err := refreshCharacterToken(context.Background(), store, refresher, testCharacterID)

	if !eve.IsRefreshTokenRevoked(err) {
		t.Fatalf("err = %v, want revoked", err)
	}
	if store.markedWith != "old-refresh" {
		t.Errorf("markedWith = %q, want old-refresh", store.markedWith)
	}
}

func TestRefreshSuccessRecoversInvalidCharacter(t *testing.T) {
	store := expiredStore(true)
	refresher := &fakeRefresher{resp: &eve.TokenResponse{AccessToken: fakeAccessToken(), RefreshToken: "new-refresh", ExpiresIn: 1199}}

	token, err := refreshCharacterToken(context.Background(), store, refresher, testCharacterID)

	if err != nil {
		t.Fatal(err)
	}
	if token != fakeAccessToken() || store.char.TokenInvalid || store.char.RefreshToken != "new-refresh" {
		t.Errorf("token = %q, char = %+v", token, store.char)
	}
	if store.char.Scopes != "esi-a.v1 esi-b.v1" {
		t.Errorf("Scopes = %q", store.char.Scopes)
	}
}

func TestRefreshKeepsOldRefreshTokenWhenResponseOmitsIt(t *testing.T) {
	store := expiredStore(false)
	refresher := &fakeRefresher{resp: &eve.TokenResponse{AccessToken: fakeAccessToken(), ExpiresIn: 1199}}

	if _, err := refreshCharacterToken(context.Background(), store, refresher, testCharacterID); err != nil {
		t.Fatal(err)
	}
	if store.char.RefreshToken != "old-refresh" {
		t.Errorf("RefreshToken = %q, want old-refresh", store.char.RefreshToken)
	}
}

func TestRefreshSkipsWhenAlreadyRefreshed(t *testing.T) {
	store := expiredStore(false)
	store.char.AccessToken = "fresh"
	store.char.TokenExpiry = time.Now().Add(time.Hour)
	refresher := &fakeRefresher{}

	token, err := refreshCharacterToken(context.Background(), store, refresher, testCharacterID)

	if err != nil || token != "fresh" {
		t.Fatalf("token = %q, err = %v", token, err)
	}
	if refresher.calls.Load() != 0 {
		t.Errorf("calls = %d, want 0", refresher.calls.Load())
	}
}

func TestConcurrentRefreshesAreMerged(t *testing.T) {
	store := expiredStore(false)
	refresher := &fakeRefresher{
		resp:  &eve.TokenResponse{AccessToken: fakeAccessToken(), RefreshToken: "new-refresh", ExpiresIn: 1199},
		block: make(chan struct{}),
	}

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := refreshCharacterToken(context.Background(), store, refresher, testCharacterID); err != nil {
				t.Error(err)
			}
		}()
	}
	time.Sleep(50 * time.Millisecond) // 等待所有调用进入 singleflight
	close(refresher.block)
	wg.Wait()

	if refresher.calls.Load() != 1 {
		t.Errorf("calls = %d, want 1", refresher.calls.Load())
	}
}
