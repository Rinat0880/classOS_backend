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

// Устанавливает соединение с AD
func (ads *ADService) connect() (*ldap.Conn, error) {
	if !ads.enabled {
		return nil, fmt.Errorf("AD service is disabled")
	}

	var conn *ldap.Conn
	var err error

	address := fmt.Sprintf("%s:%s", ads.host, ads.port)

	if ads.useTLS {
		// LDAPS соединение
		conn, err = ldap.DialTLS("tcp", address, &tls.Config{
			ServerName:         ads.host,
			InsecureSkipVerify: false, // в продакшене должно быть false
		})
	} else {
		// Обычное LDAP соединение
		conn, err = ldap.Dial("tcp", address)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to AD: %w", err)
	}

	// Авторизация
	if err := conn.Bind(ads.bindUser, ads.bindPass); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind to AD: %w", err)
	}

	return conn, nil
}

// Тестирует соединение с AD
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

// Создает пользователя в AD
func (ads *ADService) CreateUser(user ADUser, password string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Формируем DN пользователя
	userDN := fmt.Sprintf("CN=%s,OU=Users,OU=Managed,%s", user.DisplayName, ads.baseDN)

	logrus.WithFields(logrus.Fields{
		"userDN": userDN,
		"sam":    user.SamAccountName,
	}).Info("Creating AD user")

	// Создаем пользователя (сначала disabled)
	addRequest := ldap.NewAddRequest(userDN, nil)
	addRequest.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
	addRequest.Attribute("cn", []string{user.DisplayName})
	addRequest.Attribute("sn", []string{user.DisplayName}) // фамилия
	addRequest.Attribute("displayName", []string{user.DisplayName})
	addRequest.Attribute("sAMAccountName", []string{user.SamAccountName})
	addRequest.Attribute("userPrincipalName", []string{user.UserPrincipalName})
	addRequest.Attribute("userAccountControl", []string{"514"}) // disabled account

	if err := conn.Add(addRequest); err != nil {
		return fmt.Errorf("failed to create user in AD: %w", err)
	}

	// Устанавливаем пароль
	if password != "" {
		if err := ads.setUserPassword(conn, userDN, password); err != nil {
			// Пытаемся удалить пользователя при ошибке
			ads.deleteUserByDN(conn, userDN)
			return fmt.Errorf("failed to set password: %w", err)
		}
	}

	// Включаем учетку
	if user.Enabled {
		if err := ads.enableUser(conn, userDN); err != nil {
			// Пытаемся удалить пользователя при ошибке
			ads.deleteUserByDN(conn, userDN)
			return fmt.Errorf("failed to enable user: %w", err)
		}
	}

	logrus.WithField("userDN", userDN).Info("AD user created successfully")
	return nil
}

// Устанавливает пароль пользователя
func (ads *ADService) setUserPassword(conn *ldap.Conn, userDN, password string) error {
	// Конвертируем пароль в UTF-16LE формат для AD
	passwordBytes := ads.encodePasswordForAD(password)

	modifyRequest := ldap.NewModifyRequest(userDN, nil)
	modifyRequest.Replace("unicodePwd", []string{string(passwordBytes)})

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	return nil
}

// Включает пользователя
func (ads *ADService) enableUser(conn *ldap.Conn, userDN string) error {
	modifyRequest := ldap.NewModifyRequest(userDN, nil)
	modifyRequest.Replace("userAccountControl", []string{"512"}) // normal account

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to enable user: %w", err)
	}

	return nil
}

// Кодирует пароль для AD (UTF-16LE)
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

// Обновляет пользователя в AD
func (ads *ADService) UpdateUser(username string, updates ADUser) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Найдем пользователя по sAMAccountName
	userDN, err := ads.findUserDN(conn, username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	modifyRequest := ldap.NewModifyRequest(userDN, nil)
	hasChanges := false

	if updates.DisplayName != "" {
		modifyRequest.Replace("displayName", []string{updates.DisplayName})
		modifyRequest.Replace("cn", []string{updates.DisplayName})
		hasChanges = true
	}

	if updates.EmailAddress != "" {
		modifyRequest.Replace("mail", []string{updates.EmailAddress})
		hasChanges = true
	}

	if updates.UserPrincipalName != "" {
		modifyRequest.Replace("userPrincipalName", []string{updates.UserPrincipalName})
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("no fields to update")
	}

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to update user in AD: %w", err)
	}

	logrus.WithField("userDN", userDN).Info("AD user updated successfully")
	return nil
}

// Удаляет пользователя из AD
func (ads *ADService) DeleteUser(username string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Найдем пользователя по sAMAccountName
	userDN, err := ads.findUserDN(conn, username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return ads.deleteUserByDN(conn, userDN)
}

// Удаляет пользователя по DN
func (ads *ADService) deleteUserByDN(conn *ldap.Conn, userDN string) error {
	delRequest := ldap.NewDelRequest(userDN, nil)
	if err := conn.Del(delRequest); err != nil {
		return fmt.Errorf("failed to delete user from AD: %w", err)
	}

	logrus.WithField("userDN", userDN).Info("AD user deleted successfully")
	return nil
}

// Изменяет пароль пользователя
func (ads *ADService) ChangeUserPassword(username, newPassword string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Найдем пользователя по sAMAccountName
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
func (ads *ADService) CreateGroup(group ADGroup) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Формируем DN группы
	groupDN := fmt.Sprintf("CN=%s,OU=Groups,OU=Managed,%s", group.Name, ads.baseDN)

	logrus.WithField("groupDN", groupDN).Info("Creating AD group")

	addRequest := ldap.NewAddRequest(groupDN, nil)
	addRequest.Attribute("objectClass", []string{"top", "group"})
	addRequest.Attribute("cn", []string{group.Name})
	addRequest.Attribute("sAMAccountName", []string{group.Name})
	addRequest.Attribute("groupType", []string{"-2147483646"}) // Global security group

	if group.Description != "" {
		addRequest.Attribute("description", []string{group.Description})
	}

	if err := conn.Add(addRequest); err != nil {
		return fmt.Errorf("failed to create group in AD: %w", err)
	}

	logrus.WithField("groupDN", groupDN).Info("AD group created successfully")
	return nil
}

// Обновляет группу в AD
func (ads *ADService) UpdateGroup(groupName string, updates ADGroup) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Найдем группу по имени
	groupDN, err := ads.findGroupDN(conn, groupName)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	hasChanges := false

	if updates.Description != "" {
		modifyRequest.Replace("description", []string{updates.Description})
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("no fields to update")
	}

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to update group in AD: %w", err)
	}

	logrus.WithField("groupDN", groupDN).Info("AD group updated successfully")
	return nil
}

// Удаляет группу из AD
func (ads *ADService) DeleteGroup(groupName string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Найдем группу по имени
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

// Находит DN группы по имени
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

// Добавляет пользователя в группу
func (ads *ADService) AddUserToGroup(username, groupName string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Найдем DN пользователя и группы
	userDN, err := ads.findUserDN(conn, username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	groupDN, err := ads.findGroupDN(conn, groupName)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	// Добавляем пользователя в группу
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

// Удаляет пользователя из группы
func (ads *ADService) RemoveUserFromGroup(username, groupName string) error {
	if !ads.enabled {
		return fmt.Errorf("AD service is disabled")
	}

	conn, err := ads.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Найдем DN пользователя и группы
	userDN, err := ads.findUserDN(conn, username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	groupDN, err := ads.findGroupDN(conn, groupName)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	// Удаляем пользователя из группы
	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Delete("member", []string{userDN})

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to remove user from group in AD: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"userDN":  userDN,
		"groupDN": groupDN,
	}).Info("User removed from AD group successfully")

	return nil
}

// Получает всех пользователей из AD
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
		fmt.Sprintf("OU=Users,OU=Managed,%s", ads.baseDN),
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

// Проверяет, включен ли пользователь
func (ads *ADService) isUserEnabled(userAccountControl string) bool {
	// userAccountControl: 512 = enabled, 514 = disabled
	return userAccountControl == "512"
}

// Синхронизирует всех пользователей из AD
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
