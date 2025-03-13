package ldapServer

import (
	"fmt"
	"os"

	"github.com/go-ldap/ldap"
)

type LDAPServer struct {
	LDAPConn           *ldap.Conn
	FQDN               string
	BaseDN             string
	UserAccountControl string
}

func New(fqdn, baseDN, userAccountControl string) (*LDAPServer, error) {
	const op = "storage.ldapServer.New" // Имя текущей функции для логов и ошибок

	// You can also use IP instead of FQDN
	ldapConn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:389", fqdn))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &LDAPServer{LDAPConn: ldapConn, FQDN: fqdn, BaseDN: baseDN, UserAccountControl: userAccountControl}, nil
}

// Check user exists and get info in []string{userDN, name, position, department, mobile, mail}
func (ldapsrv *LDAPServer) GetUserInfo(username string) ([]string, error) {
	const op = "storage.ldapServer.GetUserInfo"

	srvUsername, exists := os.LookupEnv("LDAP_USERNAME")
	if !exists {
		return nil, fmt.Errorf("%s: username for LDAP does not exists in env", op)
	}
	srvPassword, exists := os.LookupEnv("LDAP_PASSWORD")
	if !exists {
		return nil, fmt.Errorf("%s: password for LDAP does not exists in env", op)
	}

	err := ldapsrv.LDAPConn.Bind(srvUsername, srvPassword)
	if err != nil && ldap.IsErrorWithCode(err, 200) {
		ldapsrv.LDAPConn, err = ldap.DialURL(fmt.Sprintf("ldap://%s:389", ldapsrv.FQDN))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	filter := fmt.Sprintf("(&(objectCategory=Person)(sAMAccountName=%s)(!(UserAccountControl:%s))!", username, ldapsrv.UserAccountControl)

	searchReq := ldap.NewSearchRequest(
		ldapsrv.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{"Name", "Title", "Department"},
		nil,
	)
	result, err := ldapsrv.LDAPConn.Search(searchReq)
	if err != nil && ldap.IsErrorWithCode(err, 200) {
		ldapsrv.LDAPConn, err = ldap.DialURL(fmt.Sprintf("ldap://%s:389", ldapsrv.FQDN))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("%s: empty search result", op)
	}

	userInfo := []string{result.Entries[0].DN}

	name := result.Entries[0].GetAttributeValues("name")
	if len(name) != 0 {
		userInfo = append(userInfo, name[0])
	} else {
		userInfo = append(userInfo, "")
	}

	title := result.Entries[0].GetAttributeValues("title")
	if len(title) != 0 {
		userInfo = append(userInfo, title[0])
	} else {
		userInfo = append(userInfo, "")
	}

	department := result.Entries[0].GetAttributeValues("department")
	if len(department) != 0 {
		userInfo = append(userInfo, department[0])
	} else {
		userInfo = append(userInfo, "")
	}

	mobile := result.Entries[0].GetAttributeValues("Mobile")
	if len(mobile) != 0 {
		userInfo = append(userInfo, mobile[0])
	} else {
		userInfo = append(userInfo, "")
	}

	mail := result.Entries[0].GetAttributeValues("mail")
	if len(mail) != 0 {
		userInfo = append(userInfo, mail[0])
	} else {
		userInfo = append(userInfo, "")
	}

	return userInfo, nil
}

func (ldapsrv *LDAPServer) IsUserMemberOf(username, group string) (bool, error) {
	const op = "storage.ldapServer.IsUserMemberOf"

	srvUsername, exists := os.LookupEnv("LDAP_USERNAME")
	if !exists {
		return false, fmt.Errorf("%s: username for LDAP does not exists in env", op)
	}
	srvPassword, exists := os.LookupEnv("LDAP_PASSWORD")
	if !exists {
		return false, fmt.Errorf("%s: password for LDAP does not exists in env", op)
	}

	err := ldapsrv.LDAPConn.Bind(srvUsername, srvPassword)
	if err != nil && ldap.IsErrorWithCode(err, 200) {
		ldapsrv.LDAPConn, err = ldap.DialURL(fmt.Sprintf("ldap://%s:389", ldapsrv.FQDN))
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
	} else if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	filter := fmt.Sprintf("(&(sAMAccountName=%s)(memberof=CN=%s,CN=Users,%s))", username, group, ldapsrv.BaseDN)

	searchReq := ldap.NewSearchRequest(
		ldapsrv.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{},
		nil,
	)
	result, err := ldapsrv.LDAPConn.Search(searchReq)
	if err != nil && ldap.IsErrorWithCode(err, 200) {
		ldapsrv.LDAPConn, err = ldap.DialURL(fmt.Sprintf("ldap://%s:389", ldapsrv.FQDN))
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
	} else if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if len(result.Entries) == 0 {
		return false, nil
	}

	return true, nil
}
