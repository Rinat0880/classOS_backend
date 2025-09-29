package service

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"unicode/utf16"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
)

type ADUser struct {
	SamAccountName    string `json:"sam_account_name"`
	DisplayName       string `json:"display_name"`
	EmailAddress      string `json:"email_address"`
	UserPrincipalName string `json:"user_principal_name"`
	Enabled           bool   `json:"enabled"`
	DistinguishedName string `json:"distinguished_name"`
	Password          string `json:"password"`
}

type ADGroup struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	DistinguishedName string `json:"distinguished_name"`
}

type ADService struct {
	host     string
	port     string
	baseDN   string
	bindUser string
	bindPass string
	useTLS   bool
	enabled  bool
}

func NewADService() *ADService {
	// Проверяем, включен ли AD
	enabled := os.Getenv("AD_HOST") != "" &&
		os.Getenv("AD_BIND_USER") != "" &&
		os.Getenv("AD_BIND_PASS") != ""

	port := os.Getenv("AD_PORT")
	if port == "" {
		port = "389" // default LDAP port
	}

	useTLS := port == "636" || strings.ToLower(os.Getenv("AD_USE_TLS")) == "true"

	service := &ADService{
		host:     os.Getenv("AD_HOST"),
		port:     port,
		baseDN:   os.Getenv("AD_BASE_DN"),
		bindUser: os.Getenv("AD_BIND_USER"),
		bindPass: os.Getenv("AD_BIND_PASS"),
		useTLS:   useTLS,
		enabled:  enabled,
	}

	logrus.WithFields(logrus.Fields{
		"enabled": service.enabled,
		"host":    service.host,
		"port":    service.port,
		"useTLS":  service.useTLS,
		"baseDN":  service.baseDN,
	}).Info("AD Service initialized")

	return service
}

func (ads *ADService) connect() (*ldap.Conn, error) {
	if !ads.enabled {
		return nil, fmt.Errorf("AD service is disabled")
	}

	var conn *ldap.Conn
	var err error

	address := fmt.Sprintf("%s:%s", ads.host, ads.port)

	conn, err = ldap.DialTLS("tcp", address, &tls.Config{
		InsecureSkipVerify: false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to AD: %w", err)
	}

	if err := conn.Bind(ads.bindUser, ads.bindPass); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind user(%s) to AD: %w", ads.bindUser, err)
	}

	return conn, nil
}

func (ads *ADService) TestConnection() error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled - missing AD configuration")
	}

	logrus.Info("Testing AD connection...")

	conn, err := ads.connect()
	if err != nil {
		logrus.WithError(err).Error("AD connection test failed")
		return err
	}
	defer conn.Close()

	// Попытка поиска базового DN для проверки доступности
	searchRequest := ldap.NewSearchRequest(
		ads.baseDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		1, 0, false,
		"(objectClass=*)",
		[]string{"dn"},
		nil,
	)

	_, err = conn.Search(searchRequest)
	if err != nil {
		logrus.WithError(err).Error("AD search test failed")
		return fmt.Errorf("AD search test failed: %w", err)
	}

	logrus.Info("AD connection test successful")
	return nil
}

func (ads *ADService) CreateGroup(group ADGroup) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is not enabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Print("user: ", ads.bindUser, "|  pswrd: ", ads.bindPass, "|  baseDN: ", ads.baseDN)

	groupDN := fmt.Sprintf("CN=%s, OU=classos_groups, %s", group.Name, ads.baseDN)

	logrus.WithFields(logrus.Fields{
		"groupDN": groupDN,
	}).Info("Creating AD group")

	addReq := ldap.NewAddRequest(groupDN, []ldap.Control{})

	addReq.Attribute("objectClass", []string{"top", "group"})
	addReq.Attribute("name", []string{group.Name})
	addReq.Attribute("sAMAccountName", []string{group.Name})
	addReq.Attribute("instanceType", []string{fmt.Sprintf("%d", 0x00000004)})
	addReq.Attribute("groupType", []string{fmt.Sprintf("%d", 0x00000004|0x80000000)})

	if err := conn.Add(addReq); err != nil {
		return fmt.Errorf("failed to create group in AD: %w", err)
	}

	logrus.WithField("groupDN", groupDN).Info("AD group created successfully")
	return nil
}

func (ads *ADService) CreateUser(user ADUser, password string, groupname string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Print("user: ", ads.bindUser, "|  pswrd: ", ads.bindPass, "|  baseDN: ", ads.baseDN)

	userDN := fmt.Sprintf("CN=%s, OU=classos_users, %s", user.DisplayName, ads.baseDN)

	logrus.WithFields(logrus.Fields{
		"userDN": userDN,
		"sam":    user.SamAccountName,
	}).Info("Creating AD user")

	// addRequest := ldap.NewAddRequest(userDN, []ldap.Control{})
	// addRequest.Attribute("objectClass", []string{"top", "organizationalPerson", "user", "person"})
	// addRequest.Attribute("cn", []string{user.DisplayName})
	// addRequest.Attribute("sn", []string{user.DisplayName}) // фамилия
	// addRequest.Attribute("displayName", []string{user.DisplayName})
	// addRequest.Attribute("sAMAccountName", []string{user.SamAccountName})
	// addRequest.Attribute("userPrincipalName", []string{user.UserPrincipalName})
	// addRequest.Attribute("userAccountControl", []string{"514"}) // disabled account

	addRequest := ldap.NewAddRequest(userDN, []ldap.Control{})
	addRequest.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
	addRequest.Attribute("cn", []string{user.DisplayName})
	addRequest.Attribute("sn", []string{"User"})
	addRequest.Attribute("displayName", []string{user.DisplayName})
	addRequest.Attribute("sAMAccountName", []string{user.SamAccountName})
	addRequest.Attribute("userPrincipalName", []string{
		fmt.Sprintf("%s@school.local", user.SamAccountName),
	})
	addRequest.Attribute("userAccountControl", []string{"514"})

	if err := conn.Add(addRequest); err != nil {
		return fmt.Errorf("failed to create user in AD: %w", err)
	}

	if password != "" {
		if err := ads.setUserPassword(conn, userDN, password); err != nil {
			ads.deleteUserByDN(conn, userDN)
			return fmt.Errorf("failed to set password: %w", err)
		}
	}

	// modReq := ldap.NewModifyRequest(userDN, []ldap.Control{})
    // modReq.Replace("pwdLastSet", []string{"0"})
    // if err := conn.Modify(modReq); err != nil {
    //     ads.deleteUserByDN(conn, userDN)
    //     return fmt.Errorf("failed to force password reset: %w", err)
    // }

	if user.Enabled {
		if err := ads.enableUser(conn, userDN); err != nil {
			ads.deleteUserByDN(conn, userDN)
			return fmt.Errorf("failed to enable user: %w", err)
		}
	}
	//здесь нужно добавить логику добавления инста в группу при создании
	if err := ads.AddUserToGroup(user.SamAccountName, groupname); err != nil {
		ads.deleteUserByDN(conn, userDN)
		return fmt.Errorf("failed user to add to a group: %w", err)
	}

	logrus.WithField("userDN", userDN).Info("AD user created successfully")

	return nil
}

func (ads *ADService) setUserPassword(conn *ldap.Conn, userDN, password string) error {
	passwordBytes := ads.encodePasswordForAD(password)

	modifyRequest := ldap.NewModifyRequest(userDN, nil)
	modifyRequest.Replace("unicodePwd", []string{string(passwordBytes)})

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	return nil
}

func (ads *ADService) enableUser(conn *ldap.Conn, userDN string) error {
	modifyRequest := ldap.NewModifyRequest(userDN, nil)
	modifyRequest.Replace("userAccountControl", []string{"66048"}) 

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to enable user: %w", err)
	}

	return nil
}

func (ads *ADService) encodePasswordForAD(password string) []byte {
	quotedPassword := fmt.Sprintf("\"%s\"", password)
	utf16Password := utf16.Encode([]rune(quotedPassword))

	// Конвертируем в little-endian bytes
	passwordBytes := make([]byte, len(utf16Password)*2)
	for i, r := range utf16Password {
		passwordBytes[i*2] = byte(r)
		passwordBytes[i*2+1] = byte(r >> 8)
	}

	return passwordBytes
}

func (ads *ADService) UpdateUser(username string, updates ADUser, groupname string) error {
    if !ads.enabled {
        return fmt.Errorf("AD service is disabled")
    }

    conn, err := ads.connect()
    if err != nil {
        return err
    }
    defer conn.Close()

    userDN, err := ads.findUserDN(conn, username)
    if err != nil {
        return fmt.Errorf("user not found: %w", err)
    }

    modifyReq := ldap.NewModifyRequest(userDN, nil)

    if updates.DisplayName != "" {
        modifyReq.Replace("displayName", []string{updates.DisplayName})
    }

    if len(modifyReq.Changes) > 0 {
        if err := conn.Modify(modifyReq); err != nil {
            return fmt.Errorf("failed to modify user attributes: %w", err)
        }
    }

    if updates.Password != "" {
        if err := ads.setUserPassword(conn, userDN, updates.Password); err != nil {
            return fmt.Errorf("failed to update password: %w", err)
        }
    }

	if groupname != "" {
		if err := ads.MoveUserToAnotherGroup(username, groupname); err != nil {
			return fmt.Errorf("failed to move to another group: %w", err)
		}
	}

    logrus.WithField("userDN", userDN).Info("AD user updated successfully")
    return nil
}

func (ads *ADService) DeleteUser(username string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	userDN, err := ads.findUserDN(conn, username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return ads.deleteUserByDN(conn, userDN)
}

func (ads *ADService) deleteUserByDN(conn *ldap.Conn, userDN string) error {
	delRequest := ldap.NewDelRequest(userDN, nil)
	if err := conn.Del(delRequest); err != nil {
		return fmt.Errorf("failed to delete user from AD: %w", err)
	}

	logrus.WithField("userDN", userDN).Info("AD user deleted successfully")
	return nil
}

func (ads *ADService) ChangeUserPassword(username, newPassword string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	userDN, err := ads.findUserDN(conn, username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := ads.setUserPassword(conn, userDN, newPassword); err != nil {
		return err
	}

	logrus.WithField("userDN", userDN).Info("AD user password changed successfully")
	return nil
}

// Находит DN пользователя по sAMAccountName
func (ads *ADService) findUserDN(conn *ldap.Conn, username string) (string, error) {
	searchRequest := ldap.NewSearchRequest(
		ads.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1, 0, false,
		fmt.Sprintf("(&(objectClass=user)(sAMAccountName=%s))", ldap.EscapeFilter(username)),
		[]string{"dn"},
		nil,
	)

	searchResult, err := conn.Search(searchRequest)
	if err != nil {
		return "", err
	}

	if len(searchResult.Entries) == 0 {
		return "", fmt.Errorf("user %s not found", username)
	}

	return searchResult.Entries[0].DN, nil
}

// Создает группу в AD
// func (ads *ADService) CreateGroup(group ADGroup) error {
// 	if !ads.enabled {
// 		return fmt.Errorf("AD service is disabled")
// 	}

// 	conn, err := ads.connect()
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	// Формируем DN группы
// 	groupDN := fmt.Sprintf("CN=%s, %s", group.Name, ads.baseDN)

// 	logrus.WithField("groupDN", groupDN).Info("Creating AD group")

// 	addRequest := ldap.NewAddRequest(groupDN, nil)
// 	addRequest.Attribute("objectClass", []string{"top", "group"})
// 	addRequest.Attribute("cn", []string{group.Name})
// 	addRequest.Attribute("sAMAccountName", []string{group.Name})
// 	addRequest.Attribute("groupType", []string{"-2147483646"}) // Global security group

// 	if group.Description != "" {
// 		addRequest.Attribute("description", []string{group.Description})
// 	}

// 	if err := conn.Add(addRequest); err != nil {
// 		return fmt.Errorf("failed to create group in AD: %w", err)
// 	}

// 	logrus.WithField("groupDN", groupDN).Info("AD group created successfully")
// 	return nil
// }

func (ads *ADService) UpdateGroup(groupName string, updates ADGroup) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	groupDN, err := ads.findGroupDN(conn, groupName)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	logrus.Infof("ModifyDN: oldDN=%s, and new name: %s", groupDN, updates.Name)
	newRDN := fmt.Sprintf("CN=%s", ldap.EscapeFilter(updates.Name))
	logrus.Infof("newRDN: %s", newRDN)

	modifyRequest := ldap.NewModifyDNRequest(groupDN, newRDN, true, "") //остановился на изменении имени группы (CN)

	if err := conn.ModifyDN(modifyRequest); err != nil {
		return fmt.Errorf("failed to update group in AD: %w", err)
	}
	
	logrus.Infof("ModifyDN: oldDN=%s, newRDN=%s", groupDN, newRDN)

	logrus.WithField("groupDN", groupDN).Info("AD group updated successfully")
	return nil
}

func (ads *ADService) DeleteGroup(groupName string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	groupDN, err := ads.findGroupDN(conn, groupName)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	delRequest := ldap.NewDelRequest(groupDN, nil)
	if err := conn.Del(delRequest); err != nil {
		return fmt.Errorf("failed to delete group from AD: %w", err)
	}

	logrus.WithField("groupDN", groupDN).Info("AD group deleted successfully")
	return nil
}

func (ads *ADService) findGroupDN(conn *ldap.Conn, groupName string) (string, error) {
	searchRequest := ldap.NewSearchRequest(
		ads.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1, 0, false,
		fmt.Sprintf("(&(objectClass=group)(cn=%s))", ldap.EscapeFilter(groupName)),
		[]string{"dn"},
		nil,
	)

	searchResult, err := conn.Search(searchRequest)
	if err != nil {
		return "", err
	}

	if len(searchResult.Entries) == 0 {
		return "", fmt.Errorf("group %s not found", groupName)
	}

	return searchResult.Entries[0].DN, nil
}

func (ads *ADService) AddUserToGroup(username, groupName string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	userDN, err := ads.findUserDN(conn, username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	groupDN, err := ads.findGroupDN(conn, groupName)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Add("member", []string{userDN})

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to add user to group in AD: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"userDN":  userDN,
		"groupDN": groupDN,
	}).Info("User added to AD group successfully")

	return nil
}

func (ads *ADService) MoveUserToAnotherGroup(username, groupName string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	oldGroupName, err := ads.GetUserGroups(username)
	if err != nil {
		return fmt.Errorf("problem with finding group: %w", err)
	}

	userDN, err := ads.findUserDN(conn, username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	logrus.Infof("oldgroupname: %s", oldGroupName)
	groupDN, err := ads.findGroupDN(conn, oldGroupName)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Delete("member", []string{userDN})

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to remove user from group in AD: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"userDN":  userDN,
		"groupDN": groupDN,
	}).Info("User removed from AD group successfully")

	if err := ads.AddUserToGroup(username, groupname); err != nil {
		return fmt.Errorf("failed user to add to a group: %w", err)
	}

	return nil
}

func extractCN(dn string) string {
    parts := strings.Split(dn, ",")
    if len(parts) > 0 && strings.HasPrefix(parts[0], "CN=") {
        return strings.TrimPrefix(parts[0], "CN=")
    }
    return dn 
}

func (ads *ADService) GetUserGroups(username string) (string, error) {
	conn, err := ads.connect()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	searchRequest := ldap.NewSearchRequest(
		ads.baseDN, 
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(sAMAccountName=%s)", username),
		[]string{"memberOf"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return "", err
	}
	if len(sr.Entries) != 1 {
		return "", fmt.Errorf("user not found or multiple entries")
	}

	groups := sr.Entries[0].GetAttributeValues("memberOf")
	if len(groups) > 0 {
		return extractCN(groups[0]), nil
	}
	return "", fmt.Errorf("user has no group memberships")
}

func (ads *ADService) GetAllUsers() ([]ADUser, error) {
	if !ads.enabled {
		return nil, fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	searchRequest := ldap.NewSearchRequest(
		ads.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(&(objectClass=user)(!(objectClass=computer)))",
		[]string{"sAMAccountName", "displayName", "mail", "userPrincipalName", "userAccountControl", "distinguishedName"},
		nil,
	)

	searchResult, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search users in AD: %w", err)
	}

	var users []ADUser
	for _, entry := range searchResult.Entries {
		user := ADUser{
			SamAccountName:    entry.GetAttributeValue("sAMAccountName"),
			DisplayName:       entry.GetAttributeValue("displayName"),
			EmailAddress:      entry.GetAttributeValue("mail"),
			UserPrincipalName: entry.GetAttributeValue("userPrincipalName"),
			DistinguishedName: entry.DN,
			Enabled:           ads.isUserEnabled(entry.GetAttributeValue("userAccountControl")),
		}
		users = append(users, user)
	}

	return users, nil
}

func (ads *ADService) isUserEnabled(userAccountControl string) bool {
	// userAccountControl: 512 = enabled, 514 = disabled
	return userAccountControl == "512"
}

func (ads *ADService) SyncAllUsersFromAD() error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	users, err := ads.GetAllUsers()
	if err != nil {
		return err
	}

	logrus.WithField("count", len(users)).Info("Synced users from AD")
	// Здесь может быть логика синхронизации с БД
	return nil
}
