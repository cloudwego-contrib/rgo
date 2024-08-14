package generator

import (
	"github.com/cloudwego-contrib/rgo/config"
	"github.com/cloudwego-contrib/rgo/consts"
	"github.com/cloudwego-contrib/rgo/utils"
	"path/filepath"
)

func GenerateRemoteRGOCode(idlConfig *config.IDLConfig, filePath string) error {
	err := utils.CloneOrUpdateGitRepo(idlConfig.Repository, idlConfig.Branch, filePath)
	if err != nil {
		return err
	}

	return nil
}

func (rg *RGOGenerator) updateRemoteRGOCode(idlConfig *config.IDLConfig) (string, error) {
	mu := rg.getOrCreateMutex(idlConfig.Repository)

	mu.Lock()
	defer mu.Unlock()

	repoName, err := utils.GetGitRepoName(idlConfig.Repository)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(rg.rgoRepoPath, filepath.Join(consts.IDLRemotePath, repoName, idlConfig.Branch))

	if rg.ic.RemoteRepoFlag[idlConfig.Repository] {
		return filePath, nil
	}

	err = rg.UpdateRemoteRepo(filePath, idlConfig)
	if err != nil {
		return "", err
	}

	rg.ic.RemoteRepoFlag[idlConfig.Repository] = true

	return filePath, nil
}

func (rg *RGOGenerator) UpdateRemoteRepo(filePath string, idlConfig *config.IDLConfig) error {
	err := GenerateRemoteRGOCode(idlConfig, filePath)
	if err != nil {
		return err
	}
	return nil
}
