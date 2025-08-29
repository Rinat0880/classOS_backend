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
	enabled  bool
}

func NewADService() *ADService {
	// Проверяем, включен ли AD
	enabled := os.Getenv("AD_DOMAIN") != "" && 
			   os.Getenv("AD_USERNAME") != ""

	service := &ADService{
		domain:   os.Getenv("AD_DOMAIN"),
		username: os.Getenv("AD_USERNAME"),
		password: os.Getenv("AD_PASSWORD"),
		baseOU:   os.Getenv("AD_BASE_OU"),
		enabled:  enabled,
	}

	// Если baseOU не указан, используем стандартный
	if service.baseOU == "" && service.enabled {
		service.baseOU = fmt.Sprintf("CN=Users,DC=%s", strings.Replace(service.domain, ".", ",DC=", -1))
	}

	return service
}

func (ads *ADService) CreateUser(user ADUser, password string) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	fmt.Printf("Creating AD user: %s in OU: %s\n", user.SamAccountName, ads.baseOU)
	
	escapedPassword := strings.ReplaceAll(password, `"`, `""`)
	
	cmd := fmt.Sprintf(`
		Import-Module ActiveDirectory;
		$SecurePassword = ConvertTo-SecureString "%s" -AsPlainText -Force;
		New-ADUser -Name "%s" -SamAccountName "%s" -DisplayName "%s" -EmailAddress "%s" -AccountPassword $SecurePassword -Enabled $true -Path "%s";
		Write-Output "User created successfully"`,
		escapedPassword,
		user.DisplayName,
		user.SamAccountName,
		user.DisplayName,
		user.EmailAddress,
		ads.baseOU,
	)

	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to create AD user: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) UpdateUser(username string, updates ADUser) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

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

	cmd := fmt.Sprintf(`Import-Module ActiveDirectory; Set-ADUser -Identity "%s" %s; Write-Output "User updated successfully"`, 
		username, strings.Join(setParts, " "))
	
	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to update AD user: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) DeleteUser(username string) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	cmd := fmt.Sprintf(`Import-Module ActiveDirectory; Remove-ADUser -Identity "%s" -Confirm:$false; Write-Output "User deleted successfully"`, username)
	
	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to delete AD user: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) ChangeUserPassword(username, newPassword string) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	escapedPassword := strings.ReplaceAll(newPassword, `"`, `""`)
	
	cmd := fmt.Sprintf(`
		Import-Module ActiveDirectory;
		$SecurePassword = ConvertTo-SecureString "%s" -AsPlainText -Force;
		Set-ADAccountPassword -Identity "%s" -NewPassword $SecurePassword -Reset;
		Write-Output "Password changed successfully"`,
		escapedPassword, username)

	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to change AD user password: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) CreateGroup(group ADGroup) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	cmd := fmt.Sprintf(`Import-Module ActiveDirectory; New-ADGroup -Name "%s" -GroupScope Global -Description "%s"; Write-Output "Group created successfully"`, 
		group.Name, group.Description)

	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to create AD group: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) UpdateGroup(groupName string, updates ADGroup) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	if updates.Description == "" {
		return errors.New("no fields to update")
	}

	cmd := fmt.Sprintf(`Import-Module ActiveDirectory; Set-ADGroup -Identity "%s" -Description "%s"; Write-Output "Group updated successfully"`, 
		groupName, updates.Description)
	
	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to update AD group: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) DeleteGroup(groupName string) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	cmd := fmt.Sprintf(`Import-Module ActiveDirectory; Remove-ADGroup -Identity "%s" -Confirm:$false; Write-Output "Group deleted successfully"`, groupName)
	
	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to delete AD group: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) AddUserToGroup(username, groupName string) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	cmd := fmt.Sprintf(`Import-Module ActiveDirectory; Add-ADGroupMember -Identity "%s" -Members "%s"; Write-Output "User added to group successfully"`, 
		groupName, username)
	
	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to add user to AD group: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) RemoveUserFromGroup(username, groupName string) error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	cmd := fmt.Sprintf(`Import-Module ActiveDirectory; Remove-ADGroupMember -Identity "%s" -Members "%s" -Confirm:$false; Write-Output "User removed from group successfully"`, 
		groupName, username)
	
	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove user from AD group: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
	return nil
}

func (ads *ADService) GetAllUsers() ([]ADUser, error) {
	if !ads.enabled {
		return nil, errors.New("AD service is disabled")
	}

	cmd := `Import-Module ActiveDirectory; Get-ADUser -Filter * -Properties DisplayName,EmailAddress,Enabled | Select-Object SamAccountName,DisplayName,EmailAddress,Enabled | ConvertTo-Json`
	
	output, err := ads.executePowerShellCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get AD users: %v", err)
	}

	var users []ADUser
	if err := json.Unmarshal(output, &users); err != nil {
		return nil, fmt.Errorf("failed to parse AD users JSON: %v", err)
	}

	return users, nil
}

func (ads *ADService) SyncAllUsersFromAD() error {
	if !ads.enabled {
		return errors.New("AD service is disabled")
	}

	adUsers, err := ads.GetAllUsers()
	if err != nil {
		return err
	}

	fmt.Printf("Synced %d users from AD\n", len(adUsers))
	return nil
}

func (ads *ADService) TestConnection() error {
	if !ads.enabled {
		return errors.New("AD service is disabled - missing AD configuration")
	}

	fmt.Printf("Testing AD connection with domain: %s\n", ads.domain)
	fmt.Printf("Base OU: %s\n", ads.baseOU)
	
	cmd := `Import-Module ActiveDirectory; Get-ADDomain | Select-Object -ExpandProperty DNSRoot`
	output, err := ads.executePowerShellCommand(cmd)
	
	if err != nil {
		return fmt.Errorf("AD connection test failed: %v", err)
	}
	
	fmt.Printf("AD Domain response: %s\n", strings.TrimSpace(string(output)))
	return nil
}

// Основная функция выполнения PowerShell команд
func (ads *ADService) executePowerShellCommand(command string) ([]byte, error) {
	if !ads.enabled {
		return nil, errors.New("AD service is disabled")
	}

	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		// На Windows используем прямой вызов PowerShell
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	} else {
		// На Linux в контейнере используем удаленное выполнение через SSH
		return ads.executeRemotePowerShell(command)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("PowerShell command failed: %s, output: %s", err, string(output))
	}

	return output, nil
}

// Выполнение PowerShell команд удаленно через SSH (из Linux контейнера на Windows сервер)
func (ads *ADService) executeRemotePowerShell(command string) ([]byte, error) {
	// Вариант 1: SSH к Windows серверу (если настроен OpenSSH)
	if ads.isSSHAvailable() {
		return ads.executeViaSSH(command)
	}
	
	// Вариант 2: WinRM (если настроен)
	if ads.isWinRMAvailable() {
		return ads.executeViaWinRM(command)
	}
	
	// Вариант 3: Docker exec к Windows контейнеру
	return ads.executeViaDockerExec(command)
}

func (ads *ADService) isSSHAvailable() bool {
	// Проверяем наличие ssh клиента и переменных окружения для SSH
	_, err := exec.LookPath("ssh")
	return err == nil && os.Getenv("AD_SSH_HOST") != ""
}

func (ads *ADService) executeViaSSH(command string) ([]byte, error) {
	sshHost := os.Getenv("AD_SSH_HOST")
	sshUser := os.Getenv("AD_SSH_USER")
	
	if sshHost == "" || sshUser == "" {
		return nil, errors.New("SSH configuration incomplete (AD_SSH_HOST, AD_SSH_USER required)")
	}
	
	// Экранируем кавычки для SSH
	escapedCommand := strings.ReplaceAll(command, `"`, `\"`)
	
	cmd := exec.Command("ssh", 
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", sshUser, sshHost),
		fmt.Sprintf(`powershell -NoProfile -NonInteractive -Command "%s"`, escapedCommand))
	
	return cmd.CombinedOutput()
}

func (ads *ADService) isWinRMAvailable() bool {
	// Проверяем наличие winrm клиента
	_, err := exec.LookPath("winrm")
	return err == nil && os.Getenv("AD_WINRM_HOST") != ""
}

func (ads *ADService) executeViaWinRM(command string) ([]byte, error) {
	// Реализация WinRM вызова (упрощенная версия)
	winrmHost := os.Getenv("AD_WINRM_HOST")
	
	cmd := exec.Command("winrm", "invoke", "powershell", 
		"-r", winrmHost,
		"-u", ads.username,
		"-p", ads.password,
		command)
	
	return cmd.CombinedOutput()
}

func (ads *ADService) executeViaDockerExec(command string) ([]byte, error) {
	// Попытка выполнить PowerShell через docker exec (если есть Windows контейнер)
	containerName := os.Getenv("AD_WINDOWS_CONTAINER")
	if containerName == "" {
		containerName = "windows_ad_container" // значение по умолчанию
	}
	
	// Проверяем наличие docker
	_, err := exec.LookPath("docker")
	if err != nil {
		return nil, errors.New("docker CLI not available for remote PowerShell execution")
	}
	
	cmd := exec.Command("docker", "exec", containerName, 
		"powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	
	return cmd.CombinedOutput()
}