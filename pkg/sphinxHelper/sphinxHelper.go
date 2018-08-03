package sphinxHelper

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "bufio"
    "bytes"
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "os/user"
    "strings"
    "time"

    "github.com/google/go-github/github"
    "gopkg.in/src-d/go-git.v4"
    "golang.org/x/crypto/ssh"    
    ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
    "gopkg.in/src-d/go-git.v4/plumbing/object"
)

const CLR_RED =   "\x1b[31;1m"
const CLR_GRN =   "\x1b[32;1m"
const CLR_YLW =   "\x1b[33;1m"
const CLR_END =   "\x1b[0m"
const ERR_ICON =  "[!]"
const INFO_ICON = "[+]"

// GenKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func genKeyPair() (string, string, error) {
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return "", "", err
    }

    privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
    var private bytes.Buffer
    if err := pem.Encode(&private, privateKeyPEM); err != nil {
        return "", "", err
    }

    // generate public key
    pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
    if err != nil {
        return "", "", err
    }

    public := ssh.MarshalAuthorizedKey(pub)
    return string(public), private.String(), nil
}

func PushSite(remoteName string, auth *ssh2.PublicKeys, r *git.Repository, w *git.Worktree) {
    fmt.Printf("%s%s%s Beginning to push site.\n", CLR_GRN, INFO_ICON, CLR_END)
    _, addErr := w.Add(".")
    CheckErr(addErr)

    commit, comErr := w.Commit("pushing generated site", &git.CommitOptions{
        Author: &object.Signature{
            Name:  "Sphynx",
            Email: "sphynxGit@gmail.com",
            When: time.Now(),
        },
    })
    CheckErr(comErr)

    fmt.Printf("%s%s%s Creating Commit...\n", CLR_GRN, INFO_ICON, CLR_END)
    obj, headErr := r.CommitObject(commit)
    CheckErr(headErr)
    fmt.Printf("%s%s%s Commit Head\n%s\n", CLR_GRN, INFO_ICON, CLR_END, obj)

    fmt.Printf("%s%s%s Pushing generated site to repo\n", CLR_GRN, INFO_ICON, CLR_END)
    pushErr := r.Push(&git.PushOptions {
        Auth:       auth,
        RemoteName: remoteName,
        Progress:   os.Stdout,
    })
    CheckErr(pushErr)
    fmt.Printf("%s%s%s Site successfully pushed!\n", CLR_GRN, INFO_ICON, CLR_END)
}

func CreateSSHKey(client *github.Client, ctx context.Context) {
    // Host github.com-sphinx
    //     HostName github.com
    //     User git
    //     IdentityFile ~/.ssh/id_rsa_sphinx

    homeDir, hmErr := user.Current()
    CheckErr(hmErr)
    hostName, hnErr := os.Hostname()
    CheckErr(hnErr)

	if _, err := os.Stat(homeDir.HomeDir + "/.ssh/id_rsa_sphinx_" + hostName); os.IsNotExist(err) {
		pubKey, priKey, kgErr := genKeyPair()
		sphinxHelper.CheckErr(kgErr)

        fPub, pubErr := os.Create(homeDir.HomeDir + "/.ssh/id_rsa_sphinx_" + hostName + ".pub")
        CheckErr(pubErr)
        fPri, priErr := os.Create(homeDir.HomeDir + "/.ssh/id_rsa_sphinx_" + hostName)
        CheckErr(priErr)

        fmt.Fprintln(fPub, pubKey)
        fmt.Fprintln(fPri, priKey)

        chmodErr := os.Chmod(homeDir.HomeDir + "/.ssh/id_rsa_sphinx_" + hostName, 0400)
        CheckErr(chmodErr)

        cfgString := "Host github.com-sphinx\n\tHostName github.com\n\tUser git\n\tIdentityFile ~/.ssh/id_rsa_sphinx_" + hostName + "\n"

        if _, err2 := os.Stat(homeDir.HomeDir + "/.ssh/config"); os.IsNotExist(err2) {
            fCfg, cfgErr := os.Create(homeDir.HomeDir + "/.ssh/config")
            CheckErr(cfgErr)
            fmt.Fprintln(fCfg, cfgString)
        } else {
            fCfg, cfgErr := os.OpenFile(homeDir.HomeDir + "/.ssh/config", os.O_APPEND|os.O_WRONLY, 0600)
            CheckErr(cfgErr)
            _, apdErr := fCfg.WriteString(cfgString)
            CheckErr(apdErr)
        }
        pubKeyFil, pubErr := ioutil.ReadFile(homeDir.HomeDir + "/.ssh/id_rsa_sphinx_" + hostName + ".pub")
        CheckErr(pubErr)

        keyTitle := "sphinxKey"
        keyKey   := string(pubKeyFil)

        key := &github.Key {
            Title: &keyTitle,
            Key:   &keyKey,
        }

        _, _, keyErr := client.Users.CreateKey(ctx, key)
        CheckErr(keyErr)

	}
}

func ReadUserInput() (resp string) {
	reader := bufio.NewReader(os.Stdin)
    in, _ := reader.ReadString('\n')
    in = strings.TrimSuffix(in, "\n")
    return in
}

func CheckErr(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s%s Sphinx%s: %v\n", CLR_RED, ERR_ICON, CLR_END, err)
        os.Exit(1)
    }
}