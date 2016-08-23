package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

const (
	VERSION = "1.1.2"
	AUTHOR  = "FANG LI <surivlee+rancherssh@gmail.com>"
	USAGE   = `
Example:
    rancherssh my-server-1
    rancherssh "my-server*"  (equals to) rancherssh my-server%
    rancherssh %proxy%
    rancherssh "projectA-app-*" (equals to) rancherssh projectA-app-%

Configuration:
    We read configuration from config.json or config.yml in ./, /etc/rancherssh/ and ~/.rancherssh/ folders.

    If you want to use JSON format, create a config.json in the folders with content:
        {
            "endpoint": "https://rancher.server/v1", // Or "https://rancher.server/v1/projects/xxxx"
            "user": "your_access_key",
            "password": "your_access_password"
        }

    If you want to use YAML format, create a config.yml with content:
        endpoint: https://your.rancher.server // Or https://rancher.server/v1/projects/xxxx
        user: your_access_key
        password: your_access_password

    We accept environment variables as well:
        SSHRANCHER_ENDPOINT=https://your.rancher.server   // Or https://rancher.server/v1/projects/xxxx
        SSHRANCHER_USER=your_access_key
        SSHRANCHER_PASSWORD=your_access_password
`
)

type Config struct {
	Container string
	Endpoint  string
	User      string
	Password  string
}

type RancherAPI struct {
	Endpoint string
	User     string
	Password string
}

type WebTerm struct {
	SocketConn *websocket.Conn
	ttyState   *terminal.State
	errChn     chan error
}

func (w *WebTerm) wsWrite() {
	var payload string
	var err error
	var keybuf [1]byte
	for {
		_, err = os.Stdin.Read(keybuf[0:1])
		if err != nil {
			w.errChn <- err
			return
		}

		payload = base64.StdEncoding.EncodeToString(keybuf[0:1])
		err = w.SocketConn.WriteMessage(websocket.BinaryMessage, []byte(payload))
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				w.errChn <- nil
			} else {
				w.errChn <- err
			}
			return
		}
	}
}

func (w *WebTerm) wsRead() {
	var err error
	var raw []byte
	var out []byte
	for {
		_, raw, err = w.SocketConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				w.errChn <- nil
			} else {
				w.errChn <- err
			}
			return
		}
		out, err = base64.StdEncoding.DecodeString(string(raw))
		if err != nil {
			w.errChn <- err
			return
		}
		os.Stdout.Write(out)
	}
}

func (w *WebTerm) SetRawtty(isRaw bool) {
	if isRaw {
		state, err := terminal.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		w.ttyState = state
	} else {
		terminal.Restore(int(os.Stdin.Fd()), w.ttyState)
	}
}

func (w *WebTerm) Run() {
	w.errChn = make(chan error)
	w.SetRawtty(true)

	go w.wsRead()
	go w.wsWrite()

	err := <-w.errChn
	w.SetRawtty(false)

	if err != nil {
		panic(err)
	}
}

func (r *RancherAPI) formatEndpoint() string {
	if r.Endpoint[len(r.Endpoint)-1:len(r.Endpoint)] == "/" {
		return r.Endpoint[0 : len(r.Endpoint)-1]
	} else {
		return r.Endpoint
	}
}

func (r *RancherAPI) makeReq(req *http.Request) (map[string]interface{}, error) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(r.User, r.Password)

	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var tokenResp map[string]interface{}
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	return tokenResp, nil
}

func (r *RancherAPI) containerUrl(name string) string {
	req, _ := http.NewRequest("GET", r.formatEndpoint()+"/containers/", nil)
	q := req.URL.Query()
	q.Add("name_like", strings.Replace(name, "*", "%", -1))
	q.Add("state", "running")
	q.Add("kind", "container")
	req.URL.RawQuery = q.Encode()
	resp, err := r.makeReq(req)
	if err != nil {
		fmt.Println("Failed to communicate with rancher API: " + err.Error())
		os.Exit(1)
	}
	data := resp["data"].([]interface{})
	var choice = 1
	if len(data) == 0 {
		fmt.Println("Container " + name + " not existed in system, not running, or you don't have access permissions.")
		os.Exit(1)
	}
	if len(data) > 1 {
		fmt.Println("We found more than one containers in system:")
		for i, _ctn := range data {
			ctn := _ctn.(map[string]interface{})
			if _, ok := ctn["data"]; ok {
				fmt.Println(fmt.Sprintf("[%d] %s, Container ID %s in project %s, IP Address %s on Host %s", i+1, ctn["name"].(string), ctn["id"].(string), ctn["accountId"].(string), ctn["data"].(map[string]interface{})["fields"].(map[string]interface{})["primaryIpAddress"].(string), ctn["data"].(map[string]interface{})["fields"].(map[string]interface{})["dockerHostIp"].(string)))
			} else {
				fmt.Println(fmt.Sprintf("[%d] %s, Container ID %s in project %s, IP Address %s", i+1, ctn["name"].(string), ctn["id"].(string), ctn["accountId"].(string), ctn["primaryIpAddress"].(string)))
			}
		}
		fmt.Println("--------------------------------------------")
		fmt.Print("Which one you want to connect: ")
		fmt.Scan(&choice)
	}
	ctn := data[choice-1].(map[string]interface{})
	if _, ok := ctn["data"]; ok {
		fmt.Println(fmt.Sprintf("Target Container: %s, ID %s in project %s, Addr %s on Host %s", ctn["name"].(string), ctn["id"].(string), ctn["accountId"].(string), ctn["data"].(map[string]interface{})["fields"].(map[string]interface{})["primaryIpAddress"].(string), ctn["data"].(map[string]interface{})["fields"].(map[string]interface{})["dockerHostIp"].(string)))
	} else {
		fmt.Println(fmt.Sprintf("Target Container: %s, ID %s in project %s, Addr %s", ctn["name"].(string), ctn["id"].(string), ctn["accountId"].(string), ctn["primaryIpAddress"].(string)))
	}
	return r.formatEndpoint() + fmt.Sprintf(
		"/containers/%s/", ctn["id"].(string))
}

func (r *RancherAPI) getWsUrl(url string) string {
	cols, rows, _ := terminal.GetSize(int(os.Stdin.Fd()))
	req, _ := http.NewRequest("POST", url+"?action=execute",
		strings.NewReader(fmt.Sprintf(
			`{"attachStdin":true, "attachStdout":true,`+
				`"command":["/bin/sh", "-c", "TERM=xterm-256color; export TERM; `+
				`stty cols %d rows %d; `+
				`[ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh"], "tty":true}`, cols, rows)))
	resp, err := r.makeReq(req)
	if err != nil {
		fmt.Println("Failed to get access token: ", err.Error())
		os.Exit(1)
	}
	return resp["url"].(string) + "?token=" + resp["token"].(string)
}

func (r *RancherAPI) getWSConn(wsUrl string) *websocket.Conn {
	endpoint := r.formatEndpoint()
	header := http.Header{}
	header.Add("Origin", endpoint)
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, header)
	if err != nil {
		fmt.Println("We couldn't connect to this container: ", err.Error())
		os.Exit(1)
	}
	return conn
}

func (r *RancherAPI) GetContainerConn(name string) *websocket.Conn {
	fmt.Println("Searching for container " + name)
	url := r.containerUrl(name)
	fmt.Println("Getting access token")
	wsurl := r.getWsUrl(url)
	fmt.Println("SSH into container ...")
	return r.getWSConn(wsurl)
}

func ReadConfig() *Config {
	app := kingpin.New("rancherssh", USAGE)
	app.Author(AUTHOR)
	app.Version(VERSION)
	app.HelpFlag.Short('h')

	viper.SetDefault("endpoint", "")
	viper.SetDefault("user", "")
	viper.SetDefault("password", "")

	viper.SetConfigName("config")            // name of config file (without extension)
	viper.AddConfigPath(".")                 // call multiple times to add many search paths
	viper.AddConfigPath("$HOME/.rancherssh") // call multiple times to add many search paths
	viper.AddConfigPath("/etc/rancherssh/")  // path to look for the config file in
	viper.ReadInConfig()

	viper.SetEnvPrefix("rancherssh")
	viper.AutomaticEnv()

	var endpoint = app.Flag("endpoint", "Rancher server endpoint, https://your.rancher.server/v1 or https://your.rancher.server/v1/projects/xxx.").Default(viper.GetString("endpoint")).String()
	var user = app.Flag("user", "Rancher API user/accesskey.").Default(viper.GetString("user")).String()
	var password = app.Flag("password", "Rancher API password/secret.").Default(viper.GetString("password")).String()
	var container = app.Arg("container", "Container name, fuzzy match").Required().String()

	app.Parse(os.Args[1:])

	if *endpoint == "" || *user == "" || *password == "" || *container == "" {
		app.Usage(os.Args[1:])
		os.Exit(1)
	}

	return &Config{
		Container: *container,
		Endpoint:  *endpoint,
		User:      *user,
		Password:  *password,
	}

}

func main() {
	config := ReadConfig()
	rancher := RancherAPI{
		Endpoint: config.Endpoint,
		User:     config.User,
		Password: config.Password,
	}
	conn := rancher.GetContainerConn(config.Container)

	wt := WebTerm{
		SocketConn: conn,
	}
	wt.Run()

	fmt.Println("Good bye.")
}
