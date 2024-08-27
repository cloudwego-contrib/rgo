package utils

import (
	"os"
	"testing"
)

func TestCloneGitRepo(t *testing.T) {
	err := CloneGitRepo("https://github.com/cloudwego/hertz.git", "develop", "./tmp/hertz")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateGitRepo(t *testing.T) {
	_, err := os.Stat("./tmp/hertz")
	if err != nil {
		if os.IsNotExist(err) {
			TestCloneGitRepo(t)
		}
		t.Fatal(err)
	}
	err = UpdateGitRepo("develop", "./tmp/hertz")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetLatestCommitID(t *testing.T) {
	_, err := os.Stat("./tmp/hertz")
	if err != nil {
		if os.IsNotExist(err) {
			TestCloneGitRepo(t)
		}
		t.Fatal(err)
	}
	s, err := GetLatestCommitID("./tmp/hertz")
	t.Logf(s)
	if err != nil {
		t.Fatal(err)
	}
}
