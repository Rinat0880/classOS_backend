package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"errors"
	"runtime"
)

type ADUser struct {
	SamAccountName string `json:"sam_account_name"`
	DisplayName    string `json:"display_name"`
	EmailAddress   string `json:"email_address"`
	Enabled        bool   `json:"enabled"`
}

type ADGroup struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	DistinguishedName string `json:"distinguished_name"`
}

type ADService struct {
	domain   string
	username string
	password string
	baseOU   string
}

func NewADService() *ADService {
	return &ADService{
		domain:   os.Getenv("AD_DOMAIN"),
		username: os.Getenv("AD_USERNAME"),
		password: os.Getenv("AD_PASSWORD"),
		baseOU:   os.Getenv("AD_BASE_OU"),
	}
}

func (ads *ADService) CreateUser(user ADUser, password string) error {
	escapedPassword := strings.ReplaceAll(password, `"`, `""`)
	
	cmd := fmt.Sprintf(`
		$SecurePassword = ConvertTo-SecureString "%s" -AsPlainText -Force;
		New-ADUser -Name "%s" -SamAccountName "%s" -DisplayName "%s" -EmailAddress "%s" -AccountPassword $SecurePassword -Enabled $true -Path "%s"`,
		escapedPassword,
		user.DisplayName,
		user.SamAccountName,
		user.DisplayName,
		user.EmailAddress,
		ads.baseOU,
	)

	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) UpdateUser(username string, updates ADUser) error {
	var setParts []string

	if updates.DisplayName != "" {
		setParts = append(setParts, fmt.Sprintf(`-DisplayName "%s"`, updates.DisplayName))
	}
	if updates.EmailAddress != "" {
		setParts = append(setParts, fmt.Sprintf(`-EmailAddress "%s"`, updates.EmailAddress))
	}

	if len(setParts) == 0 {
		return errors.New("no fields to update")
	}

	cmd := fmt.Sprintf(`Set-ADUser -Identity "%s" %s`, username, strings.Join(setParts, " "))
	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) DeleteUser(username string) error {
	cmd := fmt.Sprintf(`Remove-ADUser -Identity "%s" -Confirm:$false`, username)
	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) ChangeUserPassword(username, newPassword string) error {
	escapedPassword := strings.ReplaceAll(newPassword, `"`, `""`)
	
	cmd := fmt.Sprintf(`
		$SecurePassword = ConvertTo-SecureString "%s" -AsPlainText -Force;
		Set-ADAccountPassword -Identity "%s" -NewPassword $SecurePassword -Reset`,
		escapedPassword, username)

	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) CreateGroup(group ADGroup) error {
	cmd := fmt.Sprintf(`New-ADGroup -Name "%s" -GroupScope Global -Description "%s"`, 
		group.Name, group.Description)

	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) UpdateGroup(groupName string, updates ADGroup) error {
	if updates.Description == "" {
		return errors.New("no fields to update")
	}

	cmd := fmt.Sprintf(`Set-ADGroup -Identity "%s" -Description "%s"`, groupName, updates.Description)
	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) DeleteGroup(groupName string) error {
	cmd := fmt.Sprintf(`Remove-ADGroup -Identity "%s" -Confirm:$false`, groupName)
	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) AddUserToGroup(username, groupName string) error {
	cmd := fmt.Sprintf(`Add-ADGroupMember -Identity "%s" -Members "%s"`, groupName, username)
	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) RemoveUserFromGroup(username, groupName string) error {
	cmd := fmt.Sprintf(`Remove-ADGroupMember -Identity "%s" -Members "%s" -Confirm:$false`, 
		groupName, username)
	_, err := ads.executePowerShellCommand(cmd)
	return err
}

func (ads *ADService) GetAllUsers() ([]ADUser, error) {
	cmd := `Get-ADUser -Filter * -Properties DisplayName,EmailAddress,Enabled | 
			Select-Object SamAccountName,DisplayName,EmailAddress,Enabled | 
			ConvertTo-Json`
	
	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return nil, err
	}

	var users []ADUser
	if err := json.Unmarshal(output, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (ads *ADService) SyncAllUsersFromAD() error {
	adUsers, err := ads.GetAllUsers()
	if err != nil {
		return err
	}

	fmt.Printf("Synced %d users from AD\n", len(adUsers))
	return nil
}

func (ads *ADService) TestConnection() error {
	cmd := `Get-ADDomain | Select-Object -ExpandProperty DNSRoot`
	_, err := ads.executePowerShellCommand(cmd)
	return err
}

// КЛЮЧЕВАЯ ФУНКЦИЯ: Выполнение PowerShell на хосте
func (ads *ADService) executePowerShellCommand(command string) ([]byte, error) {
	fullCommand := fmt.Sprintf(`Import-Module ActiveDirectory; %s`, command)
	
	var cmd *exec.Cmd
	
	// Определяем как запускать PowerShell в зависимости от окружения
	if ads.isRunningInContainer() {
		// В контейнере на Windows Server - вызываем PowerShell хоста
		cmd = ads.createHostPowerShellCommand(fullCommand)
	} else {
		// Запуск напрямую на хосте
		cmd = exec.Command("powershell", "-NoProfile", "-Command", fullCommand)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("PowerShell command failed: %s, output: %s", err, string(output))
	}

	return output, nil
}

// Проверяем запущены ли мы в контейнере
func (ads *ADService) isRunningInContainer() bool {
	// Проверяем наличие /.dockerenv файла или cgroup
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	
	// Дополнительная проверка через cgroup (для Linux контейнеров)
	if runtime.GOOS == "linux" {
		if cgroup, err := os.ReadFile("/proc/1/cgroup"); err == nil {
			return strings.Contains(string(cgroup), "docker") || strings.Contains(string(cgroup), "containerd")
		}
	}
	
	return false
}

// Создаем команду для вызова PowerShell на хосте из контейнера
func (ads *ADService) createHostPowerShellCommand(command string) *exec.Cmd {
	// На Windows Server с Docker Desktop можно вызывать команды хоста
	// через специальные механизмы
	
	if runtime.GOOS == "windows" {
		// Прямой вызов PowerShell (если доступен в контейнере Windows)
		return exec.Command("powershell", "-NoProfile", "-Command", command)
	} else {
		// В Linux контейнере на Windows хосте используем docker exec
		// для вызова команд на хосте (это требует специальной настройки)
		return exec.Command("docker", "run", "--rm", 
			"-v", "//var/run/docker.sock:/var/run/docker.sock",
			"mcr.microsoft.com/windows/servercore:ltsc2022",
			"powershell", "-NoProfile", "-Command", command)
	}
}