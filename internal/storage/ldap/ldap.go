package ldapServer

import (
	"fmt"

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

// Check user exists and get info in []string{userDN, name, position, department}
func (ldapsrv *LDAPServer) GetUserInfo(username string) ([]string, error) {
	const op = "storage.ldapServer.GetUserDN"

	srvUsername, exists := os.LookupEnv("LDAP_USERNAME")
	if !exists {
		return "", fmt.Errorf("%s: username for LDAP does not exists in env", op)
	}
	srvPassword, exists := os.LookupEnv("LDAP_PASSWORD")
	if !exists {
		return "", fmt.Errorf("%s: password for LDAP does not exists in env", op)
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

	userInfo := []string{
		result.Entries[0].DN,
		result.Entries[0].GetAttributeValues("name")[0],
		result.Entries[0].GetAttributeValues("title")[0],
		result.Entries[0].GetAttributeValues("department")[0],
	}

	return userInfo, nil
}

// Normal Bind and Search (TO DO: ERROR FORMAT)
/*func (ldapsrv *LDAPServer) myBindAndSearch() (*ldap.SearchResult, error) {
	const op = "storage.ldapServer.BindAndSearch"

	ldapsrv.LDAPConn.Bind(BindUsername, BindPassword)

	searchReq := ldap.NewSearchRequest(
		BaseDN,
		ldap.ScopeWholeSubtree, //ldap.ScopeBaseObject, // you can also use ldap.ScopeWholeSubtree
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		Filter,
		[]string{"Name", "Title", "Department", "Mobile", "mail"},
		nil,
	)
	result, err := ldapsrv.LDAPConn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("Search Error: %s", err)
	}

	if len(result.Entries) > 0 {
		return result, nil
	} else {
		return nil, fmt.Errorf("Couldn't fetch search entries")
	}
}*/
