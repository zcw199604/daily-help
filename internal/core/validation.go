package core

import (
	"errors"
	"regexp"
	"strings"
)

var containerNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)

// ValidateContainerName 校验容器名输入，用于交互式参数输入场景。
func ValidateContainerName(input string) (string, error) {
	name := strings.TrimSpace(input)
	if name == "" {
		return "", errors.New("不能为空")
	}
	if !containerNamePattern.MatchString(name) {
		return "", errors.New("仅允许字母数字及 . _ - ，且长度≤128")
	}
	return name, nil
}
