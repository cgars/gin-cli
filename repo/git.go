package repo

import (
	"fmt"
	"path"

	git "github.com/libgit2/git2go"
)

// TODO: Load from config
const user = "git"
const githost = "gin.g-node.org"

// Git callbacks

func credsCB(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	_, cred := git.NewCredSshKeyFromAgent("git")
	return git.ErrOk, &cred
}

func certCB(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	if hostname != githost {
		return git.ErrCertificate
	}
	return git.ErrOk
}

func remoteCreateCB(repo *git.Repository, name, url string) (*git.Remote, git.ErrorCode) {
	remote, err := repo.Remotes.Create(name, url)
	if err != nil {
		return nil, 1 // TODO: Return proper error codes (git.ErrorCode)
	}
	return remote, git.ErrOk
}

// **************** //

// Clone downloads a repository and sets the remote fetch and push urls
func Clone(repopath string) (*git.Repository, error) {
	remotePath := fmt.Sprintf("%s@%s:%s", user, githost, repopath)
	localPath := path.Base(repopath)

	cbs := &git.RemoteCallbacks{
		CredentialsCallback:      credsCB,
		CertificateCheckCallback: certCB,
	}
	fetchopts := &git.FetchOptions{RemoteCallbacks: *cbs}
	opts := git.CloneOptions{
		Bare:                 false,
		CheckoutBranch:       "master",
		FetchOptions:         fetchopts,
		RemoteCreateCallback: remoteCreateCB,
	}
	fmt.Printf("Downloading into '%s'... ", localPath)
	repository, err := git.Clone(remotePath, localPath, &opts)

	if err != nil {
		fmt.Printf("failed!\n")
		return nil, err
	}
	fmt.Printf("done.\n")

	return repository, nil
}
