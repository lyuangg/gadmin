package utils

import (
	"context"
	"testing"

	"github.com/lyuangg/gadmin/config"
)

const testJWTSecret = "test-secret-for-jwt-unit-test"

func initTestJWT(t *testing.T) {
	t.Helper()
	InitJWT(&config.Config{JWTSecret: testJWTSecret})
}

func TestGenerateToken_ParseToken_Roundtrip(t *testing.T) {
	initTestJWT(t)
	token, err := GenerateToken(1, "user1", "用户1", 0, false, []uint{1, 2}, 3)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != 1 || claims.Username != "user1" || claims.Nickname != "用户1" {
		t.Errorf("claims = UserID %d Username %q Nickname %q", claims.UserID, claims.Username, claims.Nickname)
	}
	if claims.Type != 0 || claims.IsSuperAdmin != false {
		t.Errorf("Type=%d IsSuperAdmin=%v", claims.Type, claims.IsSuperAdmin)
	}
	if len(claims.RoleIDs) != 2 || claims.RoleIDs[0] != 1 || claims.RoleIDs[1] != 2 {
		t.Errorf("RoleIDs = %v", claims.RoleIDs)
	}
	if claims.ID != "1:3" {
		t.Errorf("jti = %q, want 1:3", claims.ID)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	initTestJWT(t)
	_, err := ParseToken("invalid-token")
	if err == nil {
		t.Fatal("ParseToken(invalid) should return error")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	initTestJWT(t)
	token, _ := GenerateToken(1, "u", "n", 0, false, nil, 0)
	InitJWT(&config.Config{JWTSecret: "other-secret"})
	_, err := ParseToken(token)
	if err == nil {
		t.Fatal("ParseToken with wrong secret should return error")
	}
}

func TestGetTokenVersion_Valid(t *testing.T) {
	claims := &Claims{}
	claims.ID = "42:5"
	userID, tokenVersion, err := GetTokenVersion(claims)
	if err != nil {
		t.Fatalf("GetTokenVersion: %v", err)
	}
	if userID != 42 || tokenVersion != 5 {
		t.Errorf("userID=%d tokenVersion=%d", userID, tokenVersion)
	}
}

func TestGetTokenVersion_EmptyJti(t *testing.T) {
	claims := &Claims{}
	claims.ID = ""
	_, _, err := GetTokenVersion(claims)
	if err == nil {
		t.Fatal("empty jti should return error")
	}
}

func TestGetTokenVersion_BadFormat(t *testing.T) {
	claims := &Claims{}
	claims.ID = "only-one-part"
	_, _, err := GetTokenVersion(claims)
	if err == nil {
		t.Fatal("bad jti format should return error")
	}
}

func TestGetTokenVersion_InvalidUserID(t *testing.T) {
	claims := &Claims{}
	claims.ID = "abc:1"
	_, _, err := GetTokenVersion(claims)
	if err == nil {
		t.Fatal("non-numeric user_id should return error")
	}
}

func TestGetTokenVersion_InvalidTokenVersion(t *testing.T) {
	claims := &Claims{}
	claims.ID = "1:xyz"
	_, _, err := GetTokenVersion(claims)
	if err == nil {
		t.Fatal("non-numeric token_version should return error")
	}
}

func TestClaimsFromContext_NilContext(t *testing.T) {
	claims, ok := ClaimsFromContext(nil)
	if ok || claims != nil {
		t.Errorf("ClaimsFromContext(nil) = %v, %v", claims, ok)
	}
}

func TestClaimsFromContext_NoClaims(t *testing.T) {
	ctx := context.Background()
	claims, ok := ClaimsFromContext(ctx)
	if ok || claims != nil {
		t.Errorf("ClaimsFromContext(no claims) = %v, %v", claims, ok)
	}
}

func TestClaimsFromContext_ValidClaims(t *testing.T) {
	c := &Claims{UserID: 2, Username: "test"}
	ctx := context.WithValue(context.Background(), "claims", c)
	claims, ok := ClaimsFromContext(ctx)
	if !ok || claims == nil {
		t.Fatalf("ClaimsFromContext = %v, %v", claims, ok)
	}
	if claims.UserID != 2 || claims.Username != "test" {
		t.Errorf("claims = %+v", claims)
	}
}

func TestClaimsFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), "claims", "not-claims")
	claims, ok := ClaimsFromContext(ctx)
	if ok || claims != nil {
		t.Errorf("wrong type should return nil, false: %v, %v", claims, ok)
	}
}
