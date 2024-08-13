package generator

import (
	"github.com/cloudwego-contrib/rgo/config"
	"github.com/cloudwego-contrib/rgo/utils"
	"path/filepath"
)

func GenerateRemoteRGOCode(idlConfig *config.IDLConfig, rgoRepoPath string) error {
	repoName, err := utils.GetGitRepoName(idlConfig.Repository)
	if err != nil {
		return err
	}

	filePath := filepath.Join(rgoRepoPath, filepath.Join("remote", repoName, idlConfig.Branch))

	err = utils.CloneGitRepo(idlConfig.Repository, idlConfig.Branch, filePath)
	if err != nil {
		return err
	}

	err = GenerateRGOCode(filepath.Join(filePath, idlConfig.IDLPath), rgoRepoPath)
	if err != nil {
		return err
	}

	return nil
}
