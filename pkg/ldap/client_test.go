// +build integration

// Package ldap provides ldap client access to our DS deployment. Used to manage users, etc.
// This is in an integration test that requires a running ldap server
package ldap

import "testing"

func TestDSConnection_Connect_test(t *testing.T) {
	type fields struct {
		url      string
		dn       string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Pick the password up from a file
		{"localhost test", fields{"ldap://localhost:1389", "uid=admin", "xetvjwgos5e75pty0e5w3vnbpk3nwt1e"}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DSConnection{
				DN:       tt.fields.dn,
				Password: tt.fields.password,
				URL:      tt.fields.url,
			}
			if err := ds.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("DSConnection.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer ds.Close()
			err := ds.getEntry("am-identity-bind-account")
			if err != nil {
				t.Errorf("Get Entry failed %v", err)
			}
			// When testing against DJ make sure to use a strong password that passes the policy (>8, special chars, upper/lower)
			err = ds.UpdatePassword("uid=am-identity-bind-account,ou=admins,ou=identities", "Password123!")
			if err != nil {
				t.Errorf("Get Entry failed %v", err)
			}
		})
	}
}
