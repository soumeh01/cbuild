/*
 * Copyright (c) 2022-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/inittest"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../test"
const testDir = "utils"

func init() {
	inittest.TestInitialization(testRoot, testDir)
}

func TestGetExecutablePath(t *testing.T) {
	assert := assert.New(t)

	t.Run("get executable path", func(t *testing.T) {
		_, err := GetExecutablePath()
		assert.Error(err)
	})
}

func TestUpdateEnvVars(t *testing.T) {
	assert := assert.New(t)

	t.Run("test update environment variables", func(t *testing.T) {
		binPath := testRoot + "/bin"
		etcPath := testRoot + "/etc"
		env := UpdateEnvVars(binPath, etcPath)
		binPath, _ = filepath.Abs(binPath)
		etcPath, _ = filepath.Abs(etcPath)
		assert.Equal(env.BuildRoot, binPath)
		assert.Equal(env.CompilerRoot, etcPath)
		assert.NotEmpty(env.PackRoot)
	})

	t.Run("test update environment variables, with CMSIS_PACK_ROOT", func(t *testing.T) {
		binPath := testRoot + "/bin"
		etcPath := testRoot + "/etc"
		packRoot, _ := filepath.Abs(testRoot + "/packs")
		_ = os.Setenv("CMSIS_PACK_ROOT", packRoot)
		env := UpdateEnvVars(binPath, etcPath)
		binPath, _ = filepath.Abs(binPath)
		etcPath, _ = filepath.Abs(etcPath)
		assert.Equal(env.BuildRoot, binPath)
		assert.Equal(env.CompilerRoot, etcPath)
		assert.NotEmpty(env.PackRoot)
	})
}

func TestGetDefaultCmsisPackRoot(t *testing.T) {
	assert := assert.New(t)

	t.Run("get default cmsis pack root", func(t *testing.T) {
		root := GetDefaultCmsisPackRoot()
		assert.NotEmpty(root)
	})
}

func TestParseContext(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		Input           string
		ExpectError     bool
		ExpectedContext ContextItem
	}{
		// negative test cases
		{"", true, ContextItem{}},
		//{".+", true, ContextItem{}},

		{".Build.Build2+Target", true, ContextItem{}},
		{".Build+Target+Test", true, ContextItem{}},
		{"+Target.Build", true, ContextItem{}},
		{"Project+Target.Build", true, ContextItem{}},

		// positive test cases
		{".Build+", false, ContextItem{ProjectName: "", BuildType: "Build", TargetType: ""}},
		{".+Target", false, ContextItem{ProjectName: "", BuildType: "", TargetType: "Target"}},
		{"+Target", false, ContextItem{ProjectName: "", BuildType: "", TargetType: "Target"}},
		{".Build", false, ContextItem{ProjectName: "", BuildType: "Build", TargetType: ""}},
		{".Build+Target", false, ContextItem{ProjectName: "", BuildType: "Build", TargetType: "Target"}},
		{"Project", false, ContextItem{ProjectName: "Project", BuildType: "", TargetType: ""}},
		{"Project.Build", false, ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}},
		{"Project.Build+", false, ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}},
		{"Project.+Target", false, ContextItem{ProjectName: "Project", BuildType: "", TargetType: "Target"}},
		{"Project+Target", false, ContextItem{ProjectName: "Project", BuildType: "", TargetType: "Target"}},
		{"Project.Build+Target", false, ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: "Target"}},
	}
	for _, test := range testCases {
		contextItem, err := ParseContext(test.Input)
		if test.ExpectError {
			assert.Error(err)
		} else {
			assert.Nil(err)
		}
		assert.Equal(contextItem.ProjectName, test.ExpectedContext.ProjectName)
		assert.Equal(contextItem.BuildType, test.ExpectedContext.BuildType)
		assert.Equal(contextItem.TargetType, test.ExpectedContext.TargetType)
	}
}

func TestCreateContext(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		Input          ContextItem
		ExpectError    bool
		ExpectedOutput string
	}{
		{ContextItem{ProjectName: "", BuildType: "", TargetType: ""}, false, ""},
		{ContextItem{ProjectName: "Project", BuildType: "", TargetType: ""}, false, "Project"},
		{ContextItem{ProjectName: "", BuildType: "Build", TargetType: ""}, false, ".Build"},
		{ContextItem{ProjectName: "", BuildType: "", TargetType: "Target"}, false, "+Target"},
		{ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}, false, "Project.Build"},
		{ContextItem{ProjectName: "", BuildType: "Build", TargetType: "Target"}, false, ".Build+Target"},
		{ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}, false, "Project.Build"},
		{ContextItem{ProjectName: "Project", BuildType: "", TargetType: "Target"}, false, "Project+Target"},
		{ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: "Target"}, false, "Project.Build+Target"},
	}
	for _, test := range testCases {
		context := CreateContext(test.Input)
		assert.Equal(context, test.ExpectedOutput)
	}
}

func TestParseCbuildIndexFile(t *testing.T) {
	assert := assert.New(t)

	t.Run("test file not available", func(t *testing.T) {
		_, err := ParseCbuildIndexFile("Unknown.cbuild-idx.yml")
		assert.Error(err)
	})

	t.Run("test cbuild-idx file parsing", func(t *testing.T) {
		data, err := ParseCbuildIndexFile(filepath.Join(testRoot, testDir, "Test.cbuild-idx.yml"))
		assert.Nil(err)
		var re = regexp.MustCompile(`^csolution\s[\d]+.[\d+]+.[\d+].*`)
		assert.True(re.MatchString(data.BuildIdx.GeneratedBy))
		assert.Equal(data.BuildIdx.Cdefault, "HelloWorld.cdefault.yml")
		assert.Equal(data.BuildIdx.Csolution, "HelloWorld.csolution.yml")
		assert.Equal(len(data.BuildIdx.Cprojects), 2)
		assert.Equal(data.BuildIdx.Cprojects[0].Cproject, "cm0plus/HelloWorld_cm0plus.cproject.yml")
		assert.Equal(data.BuildIdx.Cprojects[1].Cproject, "cm4/HelloWorld_cm4.cproject.yml")
		assert.Equal(data.BuildIdx.Licenses, "test123")
		assert.Equal(len(data.BuildIdx.Cbuilds), 4)
		assert.Equal(data.BuildIdx.Cbuilds[0].Cbuild, "cm0plus/HelloWorld_cm0plus.Debug+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[1].Cbuild, "cm0plus/HelloWorld_cm0plus.Release+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[2].Cbuild, "cm4/HelloWorld_cm4.Debug+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[3].Cbuild, "cm4/HelloWorld_cm4.Release+FRDM-K32L3A6.cbuild.yml")
	})
}

func TestParseCbuildSetFile(t *testing.T) {
	assert := assert.New(t)

	t.Run("test file not available", func(t *testing.T) {
		_, err := ParseCbuildSetFile("Unknown.cbuild-idx.yml")
		assert.Error(err)
	})

	t.Run("test cbuild-set file parsing", func(t *testing.T) {
		data, err := ParseCbuildSetFile(filepath.Join(testRoot, testDir, "Test.cbuild-set.yml"))
		assert.Nil(err)
		var re = regexp.MustCompile(`^csolution\sversion\s[\d]+.[\d+]+.[\d+].*`)
		assert.True(re.MatchString(data.ContextSet.GeneratedBy))
		assert.Equal(len(data.ContextSet.Contexts), 4)
		assert.Equal(data.ContextSet.Contexts[0].Context, "test2.Debug+CM0")
		assert.Equal(data.ContextSet.Contexts[1].Context, "test2.Debug+CM3")
		assert.Equal(data.ContextSet.Contexts[2].Context, "test1.Debug+CM0")
		assert.Equal(data.ContextSet.Contexts[3].Context, "test1.Release+CM0")
		assert.Equal(data.ContextSet.Compiler, "AC6")
	})
}

func TestAppendUnique(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		input            []string
		addElement       string
		expectedSliceLen int
		expectedOutput   []string
	}{
		{[]string{"one", "two", "three"}, "four", 4, []string{"one", "two", "three", "four"}},
		{[]string{"one", "two", "three"}, "one", 3, []string{"one", "two", "three"}},
	}
	for _, test := range testCases {
		output := AppendUnique(test.input, test.addElement)
		assert.Len(output, test.expectedSliceLen)
		assert.Equal(output, test.expectedOutput)
	}
}

func TestContains(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		slice          []string
		element        string
		expectedResult bool
	}{
		{[]string{"one", "two", "three"}, "four", false},
		{[]string{""}, "one", false},
		{[]string{"one", "two", "three"}, "one", true},
	}

	for _, test := range testCases {
		output := Contains(test.slice, test.element)
		assert.Equal(output, test.expectedResult)
	}
}

func TestGetInstalledExePath(t *testing.T) {
	assert := assert.New(t)
	t.Run("test to get invalid executable path", func(t *testing.T) {
		path, err := GetInstalledExePath("testunknown")
		assert.Equal(path, "")
		assert.Error(err)
	})
}

func TestNormalizePath(t *testing.T) {
	assert := assert.New(t)
	t.Run("test with backslash path", func(t *testing.T) {
		path := NormalizePath("test\\input\\test.csolution.yml")
		assert.Equal(path, "test/input/test.csolution.yml")
	})

	t.Run("test NormalizePath", func(t *testing.T) {
		path := NormalizePath("test/input/test.csolution.yml")
		assert.Equal(path, "test/input/test.csolution.yml")
	})
}

func TestResolveContexts(t *testing.T) {
	assert := assert.New(t)

	allContexts := []string{
		"Project1.Debug+Target",
		"Project1.Release+Target",
		"Project1.Debug+Target2",
		"Project1.Release+Target2",
		"Project2.Debug+Target",
		"Project2.Release+Target",
		"Project2.Debug+Target2",
		"Project2.Release+Target2",
		"Project3.Debug",
		"Project4+Target",
	}

	testCases := []struct {
		contextFilters           []string
		expectedResolvedContexts []string
		ExpectError              bool
	}{
		{[]string{"Project1"}, []string{"Project1.Debug+Target", "Project1.Release+Target", "Project1.Debug+Target2", "Project1.Release+Target2"}, false},
		{[]string{".Debug"}, []string{"Project1.Debug+Target", "Project1.Debug+Target2", "Project2.Debug+Target", "Project2.Debug+Target2", "Project3.Debug"}, false},
		{[]string{"+Target"}, []string{"Project1.Debug+Target", "Project1.Release+Target", "Project2.Debug+Target", "Project2.Release+Target", "Project4+Target"}, false},
		{[]string{"Project1.Debug"}, []string{"Project1.Debug+Target", "Project1.Debug+Target2"}, false},
		{[]string{"Project1+Target"}, []string{"Project1.Debug+Target", "Project1.Release+Target"}, false},
		{[]string{".Release+Target2"}, []string{"Project1.Release+Target2", "Project2.Release+Target2"}, false},
		{[]string{"Project1.Release+Target2"}, []string{"Project1.Release+Target2"}, false},

		{[]string{"*"}, allContexts, false},
		{[]string{"*.*+*"}, allContexts, false},
		{[]string{"*.*"}, allContexts, false},
		{[]string{"Proj*"}, allContexts, false},
		{[]string{".De*"}, []string{"Project1.Debug+Target", "Project1.Debug+Target2", "Project2.Debug+Target", "Project2.Debug+Target2", "Project3.Debug"}, false},
		{[]string{"+Tar*"}, []string{"Project1.Debug+Target", "Project1.Release+Target", "Project1.Debug+Target2", "Project1.Release+Target2", "Project2.Debug+Target", "Project2.Release+Target", "Project2.Debug+Target2", "Project2.Release+Target2", "Project4+Target"}, false},
		{[]string{"Proj*.D*g"}, []string{"Project1.Debug+Target", "Project1.Debug+Target2", "Project2.Debug+Target", "Project2.Debug+Target2", "Project3.Debug"}, false},
		{[]string{"Proj*+Tar*"}, []string{"Project1.Debug+Target", "Project1.Release+Target", "Project1.Debug+Target2", "Project1.Release+Target2", "Project2.Debug+Target", "Project2.Release+Target", "Project2.Debug+Target2", "Project2.Release+Target2", "Project4+Target"}, false},
		{[]string{"Project2.Rel*+Tar*"}, []string{"Project2.Release+Target", "Project2.Release+Target2"}, false},
		{[]string{".Rel*+*2"}, []string{"Project1.Release+Target2", "Project2.Release+Target2"}, false},
		{[]string{"Project*.Release+*"}, []string{"Project1.Release+Target", "Project1.Release+Target2", "Project2.Release+Target", "Project2.Release+Target2"}, false},

		// negative tests
		{[]string{"Unknown"}, nil, true},
		{[]string{".UnknownBuild"}, nil, true},
		{[]string{"+UnknownTarget"}, nil, true},
		{[]string{"Project.UnknownBuild"}, nil, true},
		{[]string{"Project+UnknownTarget"}, nil, true},
		{[]string{".UnknownBuild+Target"}, nil, true},
		{[]string{"+Debug"}, nil, true},
		{[]string{".Target"}, nil, true},
		{[]string{"TestProject*"}, nil, true},
		{[]string{"Project.*Build"}, nil, true},
		{[]string{"Project.Debug+*H"}, nil, true},
		{[]string{"Project1.Release.Debug+Target"}, nil, true},
		{[]string{"Project1.Debug+Target+Target2"}, nil, true},
	}

	for _, test := range testCases {
		outResolvedContexts, err := ResolveContexts(allContexts, test.contextFilters)
		if test.ExpectError {
			assert.Error(err)
		} else {
			assert.Nil(err)
		}
		assert.Equal(test.expectedResolvedContexts, outResolvedContexts)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	assert := assert.New(t)

	inputList := []string{"apple", "banana", "apple", "orange", "banana", "grape"}
	UniqueList := []string{"apple", "banana", "orange", "grape"}

	outUniqueList := RemoveDuplicates(inputList)
	assert.Equal(UniqueList, outUniqueList)

	outUniqueList = RemoveDuplicates(UniqueList)
	assert.Equal(UniqueList, outUniqueList)
}

func TestFileExists(t *testing.T) {
	testFile := testRoot + "/" + testDir + "/testfile.txt"

	tests := []struct {
		name         string
		filePath     string
		expectedBool bool
		expectedErr  error
	}{
		{
			name:         "Existing File",
			filePath:     testFile,
			expectedBool: true,
			expectedErr:  nil,
		},
		{
			name:         "Non-Existing File",
			filePath:     "testdata/nonexistent.txt",
			expectedBool: false,
			expectedErr:  errutils.New(errutils.ErrFileNotExist, "testdata/nonexistent.txt"),
		},
		{
			name:         "Invalid Path",
			filePath:     "/invalid/path/here",
			expectedBool: false,
			expectedErr:  errutils.New(errutils.ErrFileNotExist, "/invalid/path/here"),
		},
	}

	// Create test files
	createTestFiles := func(testFile string) {
		// Create a dummy test file
		file, _ := os.Create(testFile)
		file.Close()
	}

	removeTestFiles := func(testFile string) {
		// Remove dummy test file
		os.Remove(testFile)
	}

	createTestFiles(testFile)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := FileExists(tt.filePath)
			if exists != tt.expectedBool {
				t.Errorf("Expected file existence %v, got %v", tt.expectedBool, exists)
			}
			if (err == nil && tt.expectedErr != nil) || (err != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}

	// Clean up test files
	removeTestFiles(testFile)
}

func TestComparePaths(t *testing.T) {
	t.Run("Windows", func(t *testing.T) {
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
			// Windows-like paths (case-insensitive)
			path1 := "C:\\Users\\Example"
			path2 := "C:/users/example"
			equal, err := ComparePaths(path1, path2)
			assert.NoError(t, err, "Error should be nil")
			assert.True(t, equal, "Paths should be considered equivalent")
		}
	})

	t.Run("Darwin", func(t *testing.T) {
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "darwin") {
			// macOS paths (case-insensitive)
			path1 := "/usr/Local/bin/Test"
			path2 := "/Usr/local/biN/tesT"
			equal, err := ComparePaths(path1, path2)
			assert.NoError(t, err, "Error should be nil")
			assert.True(t, equal, "Paths should be considered equivalent")
		}
	})

	t.Run("Linux", func(t *testing.T) {
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "linux") {
			// Linux-like paths (case-sensitive)
			path1 := "/home/user/file.txt"
			path2 := "/home/User/FILE.txt"
			equal, err := ComparePaths(path1, path2)
			assert.NoError(t, err, "Error should be nil")
			assert.False(t, equal, "Paths should not be considered equivalent")
		}
	})
}
