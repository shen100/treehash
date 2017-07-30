package main

import (
	"testing"
)

// TestTraverse 测试目录遍历
func TestTraverse(t *testing.T) {
	if result := Traverse("", "", ""); result != PathNullErr {
		t.Error("没传目录时，测试失败")
	}

	if result := Traverse("/afasf/sadfasf/rerfef", "", ""); result != InvalidPathErr {
		t.Error("传的目录不存在时，测试失败")
	}

	if result := Traverse("/Users/liushen/Pictures/普吉岛/IMG_8151.JPG", "", ""); result != FileNotDIR {
		t.Error("传一个文件，而不是目录时，测试失败")
	}

	if result := Traverse("/Users/liushen/dev/backup", "", ""); result != PermissionErr {
		t.Error("传的目录没有访问权限时，测试失败")
	}

	if result := Traverse("/Users/liushen/dev/emptydir", "", ""); result != NoChildrenErr {
		t.Error("传的目录下即没有子目录，也没有文件时，测试失败")
	}

	if result := Traverse("/Users/liushen/dev/docs", "", "/adf/asdf/"); result != OutputPathErr {
		t.Error("传的目录下即没有子目录，也没有文件时，测试失败")
	}
}
