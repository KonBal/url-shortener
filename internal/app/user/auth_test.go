package user

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type keyStoreMock struct{}

func (s *keyStoreMock) Secret() []byte {
	return []byte("secret")
}

func TestAuthenticate(t *testing.T) {
	tests := map[string]struct {
		token string

		key []byte

		wantUser    *User
		wantErr     bool
		expectedErr string
	}{
		"empty_token": {
			token:       "",
			wantErr:     true,
			expectedErr: ErrAuthenticationFailed.Error(),
		},
		"wrong_key": {
			token:       "5b2d14385bf93ea58949c4f5c533261cdf4b3b9794604f6af4203201165b86ea078189db807a1942a6c1463a94f367bf",
			key:         []byte("another key"),
			wantErr:     true,
			expectedErr: ErrAuthenticationFailed.Error(),
		},
		"wrong_signature": {
			token:       "aa2d14385bf93ea58949c4f5c533261cdf4b3b9794604f6af4203201165b86ea078189db807a1942a6c1463a94f367bf",
			key:         []byte("key"),
			wantErr:     true,
			expectedErr: ErrAuthenticationFailed.Error(),
		},
		"correct": {
			token:    "7573657231323334b4879144b509c9c892ab48e0041d7e095c0e4efcc659654dd83e8fb1fe7cf88231089d7c4ed5882fef0c6b5c83673f7b",
			key:      []byte("key"),
			wantUser: &User{UserID: "user1234"},
			wantErr:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			auth := Authenticator{SecretKeyStore: NewKeyStore(func() []byte { return tt.key })}

			got, err := auth.Authenticate(tt.token)
			if tt.wantErr {
				require.EqualError(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.wantUser, got)
		})
	}
}

func BenchmarkAuthenticate(b *testing.B) {
	auth := Authenticator{SecretKeyStore: NewKeyStore(func() []byte { return []byte("secret key of moderate size") })}
	token := "7573657231323334b4879144b509c9c892ab48e0041d7e095c0e4efcc659654dd83e8fb1fe7cf88231089d7c4ed5882fef0c6b5c83673f7b"

	for i := 0; i < b.N; i++ {
		auth.Authenticate(token)
	}
}

func TestSign(t *testing.T) {
	tests := map[string]struct {
		userID string

		key []byte

		wantToken   string
		wantErr     bool
		expectedErr string
	}{
		"empty": {
			userID:    "",
			wantToken: "5b2d14385bf93ea58949c4f5c533261cdf4b3b9794604f6af4203201165b86ea078189db807a1942a6c1463a94f367bf",
			key:       []byte("key"),
			wantErr:   false,
		},
		"correct": {
			userID:    "user1234",
			wantToken: "7573657231323334b4879144b509c9c892ab48e0041d7e095c0e4efcc659654dd83e8fb1fe7cf88231089d7c4ed5882fef0c6b5c83673f7b",
			key:       []byte("key"),
			wantErr:   false,
		},
		"nil_key": {
			userID:    "user1234",
			wantToken: "7573657231323334da96545f310144563c305b1cab825592e22445aed03e24e35b6340b1031510b883b35c5599bd36396f3bbff14acc0ae4",
			key:       nil,
			wantErr:   false,
		},
		"empty_key": {
			userID:    "user1234",
			wantToken: "7573657231323334da96545f310144563c305b1cab825592e22445aed03e24e35b6340b1031510b883b35c5599bd36396f3bbff14acc0ae4",
			key:       []byte{},
			wantErr:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			auth := Authenticator{SecretKeyStore: NewKeyStore(func() []byte { return tt.key })}

			got, err := auth.Sign(tt.userID)
			if tt.wantErr {
				require.EqualError(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.wantToken, got)
		})
	}
}

func BenchmarkSign(b *testing.B) {
	auth := Authenticator{SecretKeyStore: NewKeyStore(func() []byte { return []byte("secret key of moderate size") })}
	token := "userID12345"

	for i := 0; i < b.N; i++ {
		auth.Sign(token)
	}
}

func TestSignAndAuth(t *testing.T) {
	tests := map[string]struct {
		userID  string
		key     []byte
		wantErr bool
	}{
		"correct": {
			userID:  "1234",
			key:     []byte("key"),
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			auth := Authenticator{SecretKeyStore: NewKeyStore(func() []byte { return tt.key })}

			token, err := auth.Sign(tt.userID)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			u, err := auth.Authenticate(token)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			require.Equal(t, tt.userID, u.UserID)
		})
	}
}
