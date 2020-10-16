// Package ldap provides ldap client access to our DS deployment. Used to manage users, etc.
package ldap

import (
	"fmt"

	ldap "github.com/go-ldap/ldap"
)

// DSConnection parameters for managing the DS ldap service
type DSConnection struct {
	URL      string
	DN       string
	Password string
	ldap     *ldap.Conn
}

// Connect to LDAP server via admin credentials
func (ds *DSConnection) Connect() error {
	l, err := ldap.DialURL(ds.URL)

	if err != nil {
		return fmt.Errorf("Cant open ldap connection to %s using dn %s :  %s", ds.URL, ds.DN, err.Error())
	}

	err = l.Bind(ds.DN, ds.Password)

	fmt.Printf("Connection status = %v", err)

	if err != nil {
		defer l.Close()
		return fmt.Errorf("Cant bind ldap connection to %s wiht %s: %s ", ds.URL, ds.DN, err.Error())
	}
	ds.ldap = l
	return nil
}

// GetEntry get an ldap entry.
// This doesn't do much right now ... just searches for an entry. Just for testing
func (ds *DSConnection) getEntry(dn string) (*ldap.Entry, error) {

	req := ldap.NewSearchRequest("ou=admins,ou=identities",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(uid="+dn+")",
		[]string{"dn", "cn", "uid"}, // A list attributes to retrieve
		nil)

	res, err := ds.ldap.Search(req)
	if err != nil {
		return nil, err
	}

	// just for info...
	for _, entry := range res.Entries {
		fmt.Printf("%s: %v cn=%s\n", entry.DN, entry.GetAttributeValue("uid"), entry.GetAttributeValue("cn"))
	}

	return res.Entries[0], err
}

// UpdatePassword changes the password for the user identified by the DN. This is done as an administrative password change
// The old password is not required.
func (ds *DSConnection) UpdatePassword(DN, newPassword string) error {
	req := ldap.NewPasswordModifyRequest(DN, "", newPassword)
	_, err := ds.ldap.PasswordModify(req)
	//fmt.Printf("res = %v gen pass=%v", res, res.GeneratedPassword)
	return err
}

// GetBackupTasks query the backup tasks
func (ds *DSConnection) GetBackupTasks() error {
	req := ldap.NewSearchRequest("cn=Recurring Tasks,cn=Tasks",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=ds-task-backup)",
		[]string{}, // return the default set of entries
		nil)

	// ./ldapmodify --useSSL -X -p 1444 -D "uid=admin" -w welcome1
	// 		dn: ds-recurring-task-id=NightlyBackup2,cn=Recurring Tasks,cn=Tasks
	// changetype: add
	// objectClass: top
	// objectClass: ds-task
	// objectClass: ds-recurring-task
	// objectClass: ds-task-backup

	// description: Nightly backup at 2 AM
	// ds-backup-location: bak
	// ds-recurring-task-id: NightlyBackup2
	// ds-recurring-task-schedule: 00 02 * * *
	// ds-task-class-name: org.opends.server.tasks.BackupTask
	// ds-task-id: NightlyBackup2
	// ds-task-state: RECURRING

	res, err := ds.ldap.Search(req)
	if err != nil {
		return err
	}

	fmt.Printf("\n********* Result %v\n", res.Entries[0])
	res.PrettyPrint(4)
	return nil
}

// ScheduleBackup - create a backup task
// This can be done over 1389.
func (ds *DSConnection) ScheduleBackup() error {

	var taskID = "test"

	// the dn needs to be unique for a recurring task
	req := ldap.NewAddRequest("ds-recurring-task-id=test,cn=Recurring Tasks,cn=Tasks", []ldap.Control{})
	req.Attribute("objectclass", []string{"top", "ds-task", "ds-recurring-task", "ds-task-backup"})
	req.Attribute("description", []string{"test"})
	req.Attribute("ds-backup-location", []string{"/var/tmp"})
	req.Attribute("ds-recurring-task-id", []string{taskID})
	req.Attribute("ds-task-id", []string{taskID}) // needed?
	req.Attribute("ds-task-state", []string{"RECURRING"})
	req.Attribute("ds-recurring-task-schedule", []string{"25 * * * *"})
	req.Attribute("ds-task-class-name", []string{"org.opends.server.tasks.BackupTask"})

	err := ds.ldap.Add(req)
	if err != nil {
		return err
	}
	fmt.Printf("LDAP added ok")
	return nil
}

// Close the ldap connection
func (ds *DSConnection) Close() {
	ds.ldap.Close()
}
