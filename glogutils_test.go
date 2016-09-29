package glogutils

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "launchpad.net/gocheck"
)

func Test(t *testing.T) { TestingT(t) }

type LogUtilsSuite struct {
	currentDir string
}

var _ = Suite(&LogUtilsSuite{})

func (s *LogUtilsSuite) SetUpTest(c *C) {
	// /tmp a symlink to /private/tmp on Mac OS X
	// so use local directory to avoid skipping all files
	dir, _ := filepath.Abs(filepath.Dir("."))
	tempDir, err := ioutil.TempDir(dir, "vulcan_test")
	if err != nil {
		panic(err)
	}
	s.currentDir = tempDir
}

func (s *LogUtilsSuite) TearDownTest(c *C) {
	os.RemoveAll(s.currentDir)
}

func (s *LogUtilsSuite) createFiles(files []string, symlinks map[string]string, folders []string) {
	for _, fileName := range files {
		err := ioutil.WriteFile(filepath.Join(s.currentDir, fileName), []byte("hi"), 0644)
		if err != nil {
			panic(err)
		}
	}

	for in, out := range symlinks {
		err := os.Symlink(filepath.Join(s.currentDir, out), filepath.Join(s.currentDir, in))
		if err != nil {
			panic(err)
		}
	}

	for _, folderName := range folders {
		err := os.Mkdir(filepath.Join(s.currentDir, folderName), 0755)
		if err != nil {
			panic(err)
		}
	}
}

func (s *LogUtilsSuite) TestLogDir(c *C) {
	flag.Set("log_dir", "")
	c.Assert(LogDir(), Equals, "")
	flag.Set("log_dir", "/tmp/google")
	c.Assert(LogDir(), Equals, "/tmp/google")
}

func (s *LogUtilsSuite) TestProgramName(c *C) {
	c.Assert(programName(), Equals, "glogutils.test")
}

func (s *LogUtilsSuite) TestRemoveFiles(c *C) {
	s.createFiles(
		[]string{
			// old logs to be removed
			"vulcan.radar1.mg.log.INFO.20131004-221312.25058",
			"vulcan.radar1.mg.log.ERROR.20131005-005231.31443",
			"vulcan.radar1.mg.log.WARNING.20131005-005231.31443",

			// active logs referenced by symlinks
			"vulcan.radar1.mg.log.INFO.20131009-135124.365",
			"vulcan.radar1.mg.log.WARNING.20131005-011339.365",
			"vulcan.radar1.mg.log.ERROR.20131005-011339.365",

			// totally unrellated logs
			"mgcore-0.log",
			"mongo-radar.log.2012-12-07T05-26-56",
			"redis-test.log",
		},
		map[string]string{
			"vulcan.INFO":  "vulcan.radar1.mg.log.INFO.20131009-135124.365",
			"vulcan.WARN":  "vulcan.radar1.mg.log.WARNING.20131005-011339.365",
			"vulcan.ERROR": "vulcan.radar1.mg.log.ERROR.20131005-011339.365",
		},
		[]string{"vulcan.directory", "other.directory"},
	)
	removeFiles(s.currentDir, "vulcan", 0)
	files, err := filepath.Glob(fmt.Sprintf("%s/*", s.currentDir))
	if err != nil {
		panic(err)
	}
	filesMap := make(map[string]bool, len(files))
	for _, path := range files {
		_, fileName := filepath.Split(path)
		filesMap[fileName] = true
	}
	expected := map[string]bool{
		"vulcan.radar1.mg.log.INFO.20131009-135124.365":    true,
		"vulcan.radar1.mg.log.WARNING.20131005-011339.365": true,
		"vulcan.radar1.mg.log.ERROR.20131005-011339.365":   true,
		"vulcan.INFO":                                      true,
		"vulcan.WARN":                                      true,
		"vulcan.ERROR":                                     true,
		"vulcan.directory":                                 true,
		"other.directory":                                  true,
	}
	for fileName, _ := range expected {
		if filesMap[fileName] != true {
			c.Errorf("Expected %s to be present", fileName)
		}
	}
}

func (s *LogUtilsSuite) TestCleanupLogs(c *C) {
	flag.Set("log_dir", s.currentDir)
	c.Assert(CleanupLogs(), IsNil)

	flag.Set("log_dir", "")
	c.Assert(CleanupLogs(), IsNil)

	flag.Set("log_dir", "/something/that/does/not/exist")
	c.Assert(CleanupLogs(), IsNil)
}

func (s *LogUtilsSuite) TestRemoveFilesOlderThanMaxAge(c *C) {
	t := time.Now() // now
	today := fmt.Sprintf("%04d%02d%02d-%02d%02d%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
	)
	t = t.Add(-time.Duration(int64(24*time.Hour) * int64(3))) // 3 days ago from now
	threeDaysAgo := fmt.Sprintf("%04d%02d%02d-%02d%02d%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
	)

	s.createFiles(
		[]string{
			// very old logs to be removed
			"vulcan.radar1.mg.log.INFO.20131004-221312.25058",
			"vulcan.radar1.mg.log.ERROR.20131005-005231.31443",
			"vulcan.radar1.mg.log.WARNING.20131005-005231.31443",

			// 3 days ago logs to be retained
			fmt.Sprintf("vulcan.radar1.mg.log.INFO.%s.12345", threeDaysAgo),
			fmt.Sprintf("vulcan.radar1.mg.log.WARNING.%s.12345", threeDaysAgo),
			fmt.Sprintf("vulcan.radar1.mg.log.ERROR.%s.12345", threeDaysAgo),

			// active logs referenced by symlinks
			fmt.Sprintf("vulcan.radar1.mg.log.INFO.%s.365", today),
			fmt.Sprintf("vulcan.radar1.mg.log.WARNING.%s.365", today),
			fmt.Sprintf("vulcan.radar1.mg.log.ERROR.%s.365", today),

			// totally unrellated logs
			"mgcore-0.log",
			"mongo-radar.log.2012-12-07T05-26-56",
			"redis-test.log",
		},
		map[string]string{
			"vulcan.INFO":  fmt.Sprintf("vulcan.radar1.mg.log.INFO.%s.365", today),
			"vulcan.WARN":  fmt.Sprintf("vulcan.radar1.mg.log.WARNING.%s.365", today),
			"vulcan.ERROR": fmt.Sprintf("vulcan.radar1.mg.log.ERROR.%s.365", today),
		},
		[]string{"vulcan.directory", "other.directory"},
	)
	removeFiles(s.currentDir, "vulcan", 4) // retain logs within 4 days
	files, err := filepath.Glob(fmt.Sprintf("%s/*", s.currentDir))
	if err != nil {
		panic(err)
	}
	filesMap := make(map[string]bool, len(files))
	for _, path := range files {
		_, fileName := filepath.Split(path)
		filesMap[fileName] = true
	}
	expected := map[string]bool{
		fmt.Sprintf("vulcan.radar1.mg.log.INFO.%s.12345", threeDaysAgo):    true,
		fmt.Sprintf("vulcan.radar1.mg.log.WARNING.%s.12345", threeDaysAgo): true,
		fmt.Sprintf("vulcan.radar1.mg.log.ERROR.%s.12345", threeDaysAgo):   true,
		fmt.Sprintf("vulcan.radar1.mg.log.INFO.%s.365", today):             true,
		fmt.Sprintf("vulcan.radar1.mg.log.WARNING.%s.365", today):          true,
		fmt.Sprintf("vulcan.radar1.mg.log.ERROR.%s.365", today):            true,
		"vulcan.INFO":      true,
		"vulcan.WARN":      true,
		"vulcan.ERROR":     true,
		"vulcan.directory": true,
		"other.directory":  true,
	}
	for fileName, _ := range expected {
		if filesMap[fileName] != true {
			c.Errorf("Expected %s to be present", fileName)
		}
	}
}
