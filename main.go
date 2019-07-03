package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"net"
	"strconv"
	"time"
	"github.com/mvo5/goconfigparser"
	"golang.org/x/crypto/ssh"
)

type HostStruct struct {
	IP     string
	Port   int
	User   string
	Passwd string
	status bool
	cmd    string
}

func (h *HostStruct) setstatus(wg *sync.WaitGroup) {
	if CheckServer(h) {
		h.status = true
	} else {
		h.status = false
	}
	wg.Done()
}

func CheckServer(h *HostStruct) bool {
	timeout := time.Duration(2 * time.Second)
	portStrings := strconv.Itoa(h.Port)
	ip_port := h.IP + ":" + portStrings
	_, err := net.DialTimeout("tcp", ip_port, timeout)
	if err != nil {
		return false
	} else {
		return true
	}
}

func configtest(cmd_str string) []HostStruct {
	configfile := "hosts_info.conf"
	var hostSlice []HostStruct
	cfg := goconfigparser.New()
	cfg.ReadFile(configfile)
	val, err := cfg.Get("test_hosts", "hosts_nums")
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println("get value: ", val)
	cmd := ""	
	if len(cmd_str) != 0 {
		cmd,_ = cfg.Get("test_hosts", cmd_str)
	} 
	hostName := strings.Split(val, ";")
	// fmt.Println(hostSlice)
	for _, host := range hostName {
		// fmt.Println(host)
		hostinfo, _ := cfg.Get("test_hosts", host)
		// fmt.Println(hostinfo)
		hostinfoslice := strings.Split(hostinfo, ";")
		ip := hostinfoslice[0]
		port, _ := strconv.Atoi(hostinfoslice[1])
		user := hostinfoslice[2]
		passwd := hostinfoslice[3]
		// fmt.Printf("ip %T\t%v, port %T\t%v, user %T\t%v, passwd %T\t%v\n", ip, ip, port, port, user, user, passwd, passwd)
		hoststmp := HostStruct{
			ip, port, user, passwd, false, cmd,
		}
		hostSlice = append(hostSlice, hoststmp)
		// fmt.Println(hostSlice)
	}
	return hostSlice
}

func main() {
	cmd := os.Args[1:]
	cmd_str := strings.Join(cmd, " ")
	strat := time.Now()
	var wgStatus sync.WaitGroup
	x := configtest(cmd_str)
	for i := 0; i < len(x); i++ {
		wgStatus.Add(1)
		go x[i].setstatus(&wgStatus)
	}
	wgStatus.Wait() // 等待状态信息更新完毕
	if len(cmd) == 0 {
		for i := 0; i < len(x); i++ {
			if x[i].status == true {
				fmt.Printf("%v  True\n", x[i].IP)
			} else {
				fmt.Printf("%v  False\n", x[i].IP)
			}
		}
		fmt.Println("Total Time：  ", time.Now().Sub(strat))
	} else {
		var wg sync.WaitGroup
		for i := 0; i < len(x); i++ {
			if x[i].status == true {
				// fmt.Printf("%v is alive...it's command is << %v >>\n", x[i].IP, x[i].cmd)
				// fmt.Printf("%v is running command...\n", x[i].IP)
				wg.Add(1)
				go remoteExcute(&wg, x[i])

			} else {
				// fmt.Printf("%v is dead...it's command is << %v >>\n", x[i].IP, x[i].cmd)
				fmt.Printf("%v is dead...\n", x[i].IP)
			}
		}
		wg.Wait() //等待命令执行完毕
		fmt.Println("Total Time： ", time.Now().Sub(strat))
	}
}

func connect(h HostStruct) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		session      *ssh.Session
		err          error
	)
	config = ssh.Config{
		Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
	}

	auth = append(auth, ssh.Password(h.Passwd))

	clientConfig = &ssh.ClientConfig{
		User:    h.User,
		Auth:    auth,
		Timeout: 30 * time.Second,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr = fmt.Sprintf("%s:%d", h.IP, h.Port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, err
	}

	return session, nil

}

func remoteExcute(wg *sync.WaitGroup, h HostStruct) {
	session, err := connect(h)
	if err != nil {
		log.Fatalf("\nERROR: \n\tIP: \n\t%v\n\tINFO: \n\t%v\n", h.IP, err)
		wg.Done() // 计数器减一
		return
	}
	defer session.Close()
	if strings.Contains(h.cmd, "reboot") {
		session.Run(h.cmd)
		wg.Done()
		return
	} else {
		var stdoutBuff bytes.Buffer
		session.Stdout = &stdoutBuff
		session.Run(h.cmd)
		fmt.Printf("%v -->\t %v", h.IP, session.Stdout)
		wg.Done() // 计数器减一
		return
	}
}
