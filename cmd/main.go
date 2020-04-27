package main

import (
	"flag"
  "errors"
	"fmt"
	"io/ioutil"
  "runtime"
	"log"
	"net"
	"os"
	"path"

	"github.com/labs/scp/cmd/console"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/tmc/scp"
)

var targetServerName = flag.String("servername", "", "Target Servername*")
var configSshFilePath = flag.String("configSsh", "", "Configuration ssh file location path")

func usageAndExit(message string) {

	if message != "" {
		fmt.Fprintln(os.Stderr, message)
	}

	flag.Usage()
	fmt.Fprint(os.Stderr, "\n")

	os.Exit(1)
}

func getAgent() (agent.Agent, error) {
	agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	return agent.NewClient(agentConn), err
}

var mossSep = ".--. --- .-- . .-. . -..   -... -.--   -- -.-- .-.. -..- ... .-- \n"
var welcomeMessage = getWelcomeMessage() + console.ColorfulText(console.TextMagenta, mossSep)

func getWelcomeMessage() string {
	return "Upload local file to remote server \n"

}

func main() {

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, welcomeMessage)
		fmt.Fprint(os.Stderr, "Options:\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

  var configFilePath string
	if *targetServerName == "" {
		usageAndExit("")
  }
  if *configSshFilePath == ""{
	  configFilePath = UserHomeDir() + "/.ssh/config"
  } else {
    configFilePath = *configSshFilePath
  }

	tempFile, _ := ioutil.TempFile("", "")
	fmt.Fprintln(tempFile, "hello world")
	_, tempFileName := path.Split(tempFile.Name())

  var err error
  pemFile, user, server, err := getSshAuth(*targetServerName, configFilePath)
	if err != nil {
    fmt.Println(err.Error())
    os.Exit(9)
	}
	remoteDest := "/tmp/" + tempFileName + "-copy.log"

  var pemBytes []byte
	pemBytes = errorHandler(ioutil.ReadFile(pemFile)).([]byte)

	tempFile.Close()

	defer os.Remove(tempFile.Name())
	//**defer os.Remove(tempFile.Name() + "-copy")


	signer, _ := ssh.ParsePrivateKey(pemBytes)
	client, err := ssh.Dial("tcp", server, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		log.Fatalln("Failed to dial:", err)
	}

  var session *ssh.Session
	session = errorHandler(client.NewSession()).(*ssh.Session)

	err = scp.CopyPath(tempFile.Name(), remoteDest, session)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Apparently everything is fine!")
	}

	/*if _, err := os.Stat(remoteDest); os.IsNotExist(err) {
		fmt.Printf("no such file or directory: %s", remoteDest)
	} else {
		fmt.Println("success")
	}*/

}

func errorHandler(handler interface{}, err error) interface{} {
	if err != nil {
		//log.Fatal(err)
		log.Fatalln("Failed:", err.Error())
	}
	return handler
}

func getSshAuth(serverName string, configFilePath string) (string, string, string, error) {
	f, _ := os.Open(configFilePath)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return "", "", "", err
	}
	cfg, _ := ssh_config.Decode(f)

	identity, _ := cfg.Get(serverName, "IdentityFile")
  if identity == "" {
		return "", "", "",errors.New("Configuration not found!")
  }
	user, _ := cfg.Get(serverName, "User")
	port, _ := cfg.Get(serverName, "Port")
	server, _ := cfg.Get(serverName, "HostName")
	server = server + ":" + port

	return identity, user, server, nil
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
