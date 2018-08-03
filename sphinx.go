package main

import (
    "context"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "os/exec"
    "os/user"
    "strings"

    "git.praetorianlabs.com/mars/Sphinx/pkg/generate"
    "git.praetorianlabs.com/mars/Sphinx/pkg/sphinxHelper"
    "github.com/google/go-github/github"
    "golang.org/x/oauth2"
    "golang.org/x/crypto/ssh"
    "gopkg.in/src-d/go-git.v4"
    "gopkg.in/src-d/go-git.v4/config"
    ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const CLR_RED =   "\x1b[31;1m"
const CLR_GRN =   "\x1b[32;1m"
const CLR_YLW =   "\x1b[33;1m"
const CLR_END =   "\x1b[0m"
const ERR_ICON =  "[!]"
const INFO_ICON = "[+]"


func main() {
    if os.Getenv("SPHINX_EDITOR") == "" {
        setEditor()
    }
    setGithubAccount()
}

func setEditor() {
	editor := "vim"
	fmt.Printf("%s%s%s Set your perferred editor (default: vim): ", CLR_GRN, INFO_ICON, CLR_END)
    in := sphinxHelper.ReadUserInput()
    if in != "" {
        editor = in
    }

    os.Setenv("SPHINX_EDITOR", editor)
	fmt.Printf("%s%s%s Editor set to %s.\n", CLR_GRN, INFO_ICON, CLR_END, editor)
}

func setGithubAccount() {
    for true {
        fmt.Printf("%s%s%s Enter your GitHub username: ", CLR_GRN, INFO_ICON, CLR_END)
        gitHubUser := sphinxHelper.ReadUserInput()
        if gitHubUser == "" {
            fmt.Printf("%s%s%s You must enter a GitHub username!\n", CLR_RED, ERR_ICON, CLR_END)
        } else {
            accExistCheck, existErr := http.Get("https://github.com/" + gitHubUser)
            sphinxHelper.CheckErr(existErr)
            if accExistCheck.StatusCode == 200 {
                for true {
                    fmt.Printf("%s%s%s Enter the personal access token associated with this account: ", CLR_GRN, INFO_ICON, CLR_END)
                    token := sphinxHelper.ReadUserInput()
                    os.Setenv("GITHUB_TOKEN", token)
                    ctx := context.Background()
                    ts := oauth2.StaticTokenSource(
                        &oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
                    )
                    tc := oauth2.NewClient(ctx, ts)
    
                    client := github.NewClient(tc)
                    _, _, err := client.Repositories.List(ctx, "", nil)
                    if err == nil {
                        cmdHandler(gitHubUser)
                        break
                    } else {
                        fmt.Printf("%s%s%s That key does not work with the given account!\n", CLR_RED, ERR_ICON, CLR_END)
                    }
                }
            } else {
                fmt.Printf("%s%s%s Could not find that GitHub user. Try Again.\n", CLR_RED, ERR_ICON, CLR_END)
            }
        }
    }
}

func cmdHandler(gitHubUser string) {
    homeDir, hmErr := user.Current()
    sphinxHelper.CheckErr(hmErr)

	var r *git.Repository
    if _, err := os.Stat(homeDir.HomeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/.git"); os.IsNotExist(err) {
        r, err = git.PlainInit(homeDir.HomeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/", false)
        sphinxHelper.CheckErr(err)
    } else {
		r, err = git.PlainOpen(homeDir.HomeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/")
		sphinxHelper.CheckErr(err)
	}

    w, wtErr := r.Worktree()
    sphinxHelper.CheckErr(wtErr)

    ctx := context.Background()
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
    )
    tc := oauth2.NewClient(ctx, ts)

    client := github.NewClient(tc)

    sphinxHelper.CreateSSHKey(client, ctx)

    hostName, hnErr := os.Hostname()
    sphinxHelper.CheckErr(hnErr)

    pem, pemErr := ioutil.ReadFile(homeDir.HomeDir + "/.ssh/id_rsa_sphinx_" + hostName)
    sphinxHelper.CheckErr(pemErr)
    signer, signErr := ssh.ParsePrivateKey(pem)
    sphinxHelper.CheckErr(signErr)
    auth := &ssh2.PublicKeys{User: "git", Signer: signer}

    for true {
        fmt.Printf("%sSphinx%s > ", CLR_YLW, CLR_END)
        c := sphinxHelper.ReadUserInput()
        cmd := strings.Fields(c)

        switch cmd[0] {
            case "help":
                if len(cmd) == 1 {
                    fmt.Println("site\texit")
                } else if cmd[1] == "site" {
                    fmt.Println("list\tcreate\tedit\tregen")
                }
                break
            case "exit":
                os.Exit(0)
                break
            case "site":
                if len(cmd) == 1 {
                    fmt.Printf("%s%s%s 'site' requires an argument.\n", CLR_RED, ERR_ICON, CLR_END)
                } else {
                    switch cmd[1] {
                        case "list":
                            repos, _, err := client.Repositories.List(ctx, "", nil)
                            sphinxHelper.CheckErr(err)

                            for _, repo := range(repos) {
                                fmt.Println(repo.GetName())
                            }
                            break
                        case "create":
                            if len(cmd) < 3 {
                                fmt.Printf("%s%s%s Unrecognized argument: usage 'site create <domain> [<category>]'\n", CLR_RED, ERR_ICON, CLR_END)
                            } else {
                                repoExistCheck, existErr := http.Get("https://github.com/" + gitHubUser + "/" + cmd[2])
                                sphinxHelper.CheckErr(existErr)
                                if repoExistCheck.StatusCode != 200 {
                                    newRepo := &github.Repository {
                                        Name:    github.String(cmd[2]),
                                        Private: github.Bool(false),
                                    }
                                    _, _, createErr := client.Repositories.Create(ctx, "", newRepo)
                                    sphinxHelper.CheckErr(createErr)
                                    if len(cmd) == 3 {
                                        generate.Generate(cmd[2], "", "")
                                    } else {
                                        generate.Generate(cmd[2], cmd[3], "")
                                    }

                                    _, remErr := r.CreateRemote(&config.RemoteConfig {
                                        Name: cmd[2],
                                        URLs: []string{"git@github.com-sphinx:" + gitHubUser + "/" + cmd[2] + ".git"},
                                    })
                                    sphinxHelper.CheckErr(remErr)

                                    sphinxHelper.PushSite(cmd[2], auth, r, w)

                                    // TODO
                                    // Figure out the best way to set the
                                    // source for the repo so that it will
                                    // start hosting on the domain

                                    fmt.Printf("%s%s%s Until the go-github package implements GitHub sites calls\n    you will have to set the source manually!\n\n", CLR_YLW, ERR_ICON, CLR_END)


                                    fmt.Printf( " ------------------------------------------------------------------\n" +
                                                "| %sBlueCoat%s               |  https://sitereview.bluecoat.com/       |\n" +
                                                "|------------------------+-----------------------------------------|\n" +
                                                "| %sIBM X-Force Exchange%s   |  https://exchange.xforce.ibmcloud.com/  |\n" +
                                                "|------------------------+-----------------------------------------|\n" +
                                                "| %sTrustedSource%s          |  https://www.trustedsource.org/         |\n" +
                                                " ------------------------------------------------------------------\n\n", CLR_YLW, CLR_END, CLR_YLW, CLR_END, CLR_YLW, CLR_END) 
                                    
                                } else {
                                    fmt.Printf("%s%s%s A repo named %s already exists for user %s.\n", CLR_RED, ERR_ICON, CLR_END, cmd[2], gitHubUser)
                                }
                            }

                            break
                        case "edit":
                            if len(cmd) != 3 {
                                fmt.Printf("%s%s%s Unrecognized argument: usage 'site edit <domain>'\n", CLR_RED, ERR_ICON, CLR_END)
                            } else {
                                repoExistCheck, existErr := http.Get("https://github.com/" + gitHubUser + "/" + cmd[2])
                                sphinxHelper.CheckErr(existErr)
                                if repoExistCheck.StatusCode == 200 {
                                    _, remErr := r.Remote(cmd[2])
                                    if remErr != nil {
                                        _, rmtErr := r.CreateRemote(&config.RemoteConfig {
                                            Name:  cmd[2],
                                            URLs: []string{"git@github.com-sphinx:" + gitHubUser + "/" + cmd[2] + ".git"},
                                        })
                                        sphinxHelper.CheckErr(rmtErr)
                                    }

                                    indResp, getErr := http.Get("https://raw.githubusercontent.com/" + gitHubUser + "/" + cmd[2] + "/master/index.md")
                                    sphinxHelper.CheckErr(getErr)
                                    defer indResp.Body.Close()

                                    delErr := os.Remove("index.md")
                                    sphinxHelper.CheckErr(delErr)
                                    f, crtErr := os.Create("index.md")
                                    sphinxHelper.CheckErr(crtErr)

                                    io.Copy(f, indResp.Body)

                                    editor := exec.Command(os.Getenv("SPHINX_EDITOR"), "index.md")
                                    editor.Stdin  = os.Stdin
                                    editor.Stdout = os.Stdout
                                    edtErr := editor.Run()
                                    sphinxHelper.CheckErr(edtErr)

                                    sphinxHelper.PushSite(cmd[2], auth, r, w)

                                } else {
                                    fmt.Printf("%s%s%s Could not find repository named %s associated with %s.\n", CLR_RED, ERR_ICON, CLR_END, cmd[2], gitHubUser)
                                }
                            }
                            break
                        case "regen":
                            if len(cmd) < 3 {
                                fmt.Printf("%s%s%s Unrecognized argument: usage 'site regen <domain>'\n", CLR_RED, ERR_ICON, CLR_END)
                            } else {
                                repoExistCheck, existErr := http.Get("https://github.com/" + gitHubUser + "/" + cmd[2])
                                sphinxHelper.CheckErr(existErr)
                                if repoExistCheck.StatusCode == 200 {
                                    _, remErr := r.Remote(cmd[2])
                                    if remErr != nil {
                                        _, rmtErr := r.CreateRemote(&config.RemoteConfig {
                                            Name:  cmd[2],
                                            URLs: []string{"git@github.com-sphinx:" + gitHubUser + "/" + cmd[2] + ".git"},
                                        })
                                        sphinxHelper.CheckErr(rmtErr)
                                    }
                                    if len(cmd) == 3 {
                                        generate.Generate(cmd[2], "", "")
                                    } else {
                                        generate.Generate(cmd[2], cmd[3], "")
                                    }
                                    sphinxHelper.PushSite(cmd[2], auth, r, w)
                                } else {
                                    fmt.Printf("%s%s%s Could not find repository named %s associated with %s.\n", CLR_RED, ERR_ICON, CLR_END, cmd[2], gitHubUser)
                                }
                            }
                            break
                        case "delete":
                            if len(cmd) < 3 {
                                fmt.Printf("%s%s%s Unrecognized argument: usage 'site delete <domain>'\n", CLR_RED, ERR_ICON, CLR_END)
                            } else {
                                repoExistCheck, existErr := http.Get("https://github.com/" + gitHubUser + "/" + cmd[2])
                                sphinxHelper.CheckErr(existErr)
                                if repoExistCheck.StatusCode == 200 {
                                    _, delErr := client.Repositories.Delete(ctx, gitHubUser, cmd[2])
                                    sphinxHelper.CheckErr(delErr)

                                    _, remErr := r.Remote(cmd[2])
                                    if remErr == nil {
                                        remDelErr := r.DeleteRemote(cmd[2])
                                        sphinxHelper.CheckErr(remDelErr)
                                    }

                                    fmt.Printf("%s%s%s Site successfully deleted.\n", CLR_GRN, INFO_ICON, CLR_END)
                                } else {
                                    fmt.Printf("%s%s%s Could not find repository named %s associated with %s.\n", CLR_RED, ERR_ICON, CLR_END, cmd[2], gitHubUser)
                                }
                            }
                            break
                        case "help":
                            fmt.Printf("\n%susage:%s site <command> [options]\n" +
                                       "\tcommands:                         \n" +
                                       "\t\tlist:                          List sites associated with the GitHub account in use.\n" +
                                       "\t\tcreate <domain> [template] :   Create a new site. <domain> is the registered domain, and [template] can be either 'healthcare' or 'finance'\n" + 
                                       "\t\tdelete <domain> :              Delete site associated with the supplied domain.\n" +
                                       "\t\tregen <domain> [template} :    Regenerate and push a new template for an existing domain.\n\n", CLR_YLW, CLR_END)
                            break
                        default:
                            fmt.Printf("%s%s%s Argument '%s' not recognized.\n", CLR_RED, ERR_ICON, CLR_END, cmd[1])
                            break
                    }
                }
                break
            default:
                fmt.Printf("%s%s%s Command not recognized.\n", CLR_RED, ERR_ICON, CLR_END)
                break
        }
    }
}